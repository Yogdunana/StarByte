import { describe, expect, it } from 'vitest';
import type { MonitorApp, MonitorDatabase, MonitorRedis, MonitorServer } from '@/api/monitor';
import { applyLiveSnapshot, applySettledIfCurrent, createPollSession, mergeSettledSnapshot } from './load';

const server = { cpu_percent: 10, mem_percent: 20, disk_percent: 30 } as MonitorServer;
const app = { goroutines: 8, uptime_seconds: 12 } as MonitorApp;
const database = { available: true, in_use: 1, idle: 3 } as MonitorDatabase;
const redis = { available: true, connected_clients: 2 } as MonitorRedis;

describe('mergeSettledSnapshot', () => {
  it('keeps fulfilled slices and records rejected keys', () => {
    const { next, failed, succeeded } = mergeSettledSnapshot({}, [
      { status: 'fulfilled', value: server },
      { status: 'fulfilled', value: app },
      { status: 'rejected', reason: new Error('db') },
      { status: 'fulfilled', value: redis },
      { status: 'rejected', reason: new Error('api') },
    ]);
    expect(next.server).toEqual(server);
    expect(next.app).toEqual(app);
    expect(next.database).toBeUndefined();
    expect(next.redis).toEqual(redis);
    expect(next.api).toBeUndefined();
    expect(failed).toEqual(['database', 'api']);
    expect(succeeded).toEqual(['server', 'app', 'redis']);
  });

  it('retains last good data when a later poll rejects that slice', () => {
    const prev = { server, app, database, redis };
    const newer = { ...server, cpu_percent: 40 };
    const { next, failed } = mergeSettledSnapshot(prev, [
      { status: 'fulfilled', value: newer },
      { status: 'rejected', reason: new Error('app') },
      { status: 'rejected', reason: new Error('db') },
      { status: 'fulfilled', value: redis },
      { status: 'rejected', reason: new Error('api') },
    ]);
    expect(next.server?.cpu_percent).toBe(40);
    expect(next.app).toEqual(app);
    expect(next.database).toEqual(database);
    expect(next.redis).toEqual(redis);
    expect(failed).toEqual(['app', 'database', 'api']);
  });

  it('does not treat an empty previous snapshot as success', () => {
    const { next, failed, succeeded } = mergeSettledSnapshot({}, [
      { status: 'rejected', reason: new Error('all') },
      { status: 'rejected', reason: new Error('all') },
      { status: 'rejected', reason: new Error('all') },
      { status: 'rejected', reason: new Error('all') },
      { status: 'rejected', reason: new Error('all') },
    ]);
    expect(next).toEqual({});
    expect(succeeded).toEqual([]);
    expect(failed).toHaveLength(5);
  });
});

function settledAll(value: MonitorServer) {
  return [
    { status: 'fulfilled' as const, value },
    { status: 'fulfilled' as const, value: app },
    { status: 'fulfilled' as const, value: database },
    { status: 'fulfilled' as const, value: redis },
    { status: 'rejected' as const, reason: new Error('api') },
  ];
}

describe('applyLiveSnapshot', () => {
  it('fills present slices and records WS errors', () => {
    const { next, failed, succeeded } = applyLiveSnapshot(
      { server, app },
      { server: { ...server, cpu_percent: 44 }, redis, errors: ['database', 'api'] },
    );
    expect(next.server?.cpu_percent).toBe(44);
    expect(next.app).toEqual(app);
    expect(next.redis).toEqual(redis);
    expect(succeeded).toEqual(['server', 'redis']);
    expect(failed).toEqual(['database', 'api']);
  });
});

describe('poll generation rejects stale overwrite', () => {
  it('keeps the newer snapshot when an older poll settles later', () => {
    const session = createPollSession();
    const older = session.begin();
    const newer = session.begin();
    expect(session.isCurrent(older.gen)).toBe(false);
    expect(session.isCurrent(newer.gen)).toBe(true);
    expect(older.signal.aborted).toBe(true);
    expect(newer.signal.aborted).toBe(false);

    const staleServer = { ...server, cpu_percent: 11 };
    const freshServer = { ...server, cpu_percent: 77 };
    let snapshot = {};

    const appliedNew = applySettledIfCurrent(session, newer.gen, snapshot, settledAll(freshServer));
    expect(appliedNew).not.toBeNull();
    snapshot = appliedNew!.next;
    expect(snapshot.server?.cpu_percent).toBe(77);

    const appliedOld = applySettledIfCurrent(session, older.gen, snapshot, settledAll(staleServer));
    expect(appliedOld).toBeNull();
    expect(snapshot.server?.cpu_percent).toBe(77);
  });

  it('drops a slow earlier poll after a faster refresh (async)', async () => {
    const session = createPollSession();
    let snapshot = { server, app, database, redis };
    const cpuSamples: number[] = [];

    const apply = (
      gen: number,
      results: ReturnType<typeof settledAll>,
    ) => {
      const merged = applySettledIfCurrent(session, gen, snapshot, results);
      if (!merged) return false;
      snapshot = merged.next;
      if (merged.succeeded.includes('server') && merged.next.server) {
        cpuSamples.push(merged.next.server.cpu_percent);
      }
      return true;
    };

    const oldTicket = session.begin();
    const slowOld = new Promise<void>((resolve) => {
      setTimeout(() => {
        apply(oldTicket.gen, settledAll({ ...server, cpu_percent: 15 }));
        resolve();
      }, 30);
    });

    const newTicket = session.begin();
    apply(newTicket.gen, settledAll({ ...server, cpu_percent: 88 }));
    await slowOld;

    expect(snapshot.server?.cpu_percent).toBe(88);
    expect(cpuSamples).toEqual([88]);
  });

  it('invalidate() makes the in-flight ticket stale (unmount)', () => {
    const session = createPollSession();
    const ticket = session.begin();
    session.invalidate();
    expect(session.isCurrent(ticket.gen)).toBe(false);
    expect(ticket.signal.aborted).toBe(true);
    expect(applySettledIfCurrent(session, ticket.gen, {}, settledAll(server))).toBeNull();
  });
});
