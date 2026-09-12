import { describe, expect, it } from 'vitest';
import type { MonitorApp, MonitorDatabase, MonitorRedis, MonitorServer } from '@/api/monitor';
import { mergeSettledSnapshot } from './load';

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
