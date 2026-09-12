import type {
  MonitorAPIStats,
  MonitorApp,
  MonitorDatabase,
  MonitorLiveSnapshot,
  MonitorRedis,
  MonitorServer,
} from '@/api/monitor';

export type SnapshotKey = 'server' | 'app' | 'database' | 'redis' | 'api';

export const SNAPSHOT_KEYS: SnapshotKey[] = ['server', 'app', 'database', 'redis', 'api'];

export interface Snapshot {
  server?: MonitorServer;
  app?: MonitorApp;
  database?: MonitorDatabase;
  redis?: MonitorRedis;
  api?: MonitorAPIStats;
}

export interface MergeResult {
  next: Snapshot;
  failed: SnapshotKey[];
  succeeded: SnapshotKey[];
}

export function mergeSettledSnapshot(
  prev: Snapshot,
  results: Array<PromiseSettledResult<Snapshot[SnapshotKey]>>,
  keys: SnapshotKey[] = SNAPSHOT_KEYS,
): MergeResult {
  const next: Snapshot = { ...prev };
  const failed: SnapshotKey[] = [];
  const succeeded: SnapshotKey[] = [];
  keys.forEach((key, i) => {
    const result = results[i];
    if (result?.status === 'fulfilled' && result.value != null) {
      next[key] = result.value as never;
      succeeded.push(key);
      return;
    }
    failed.push(key);
  });
  return { next, failed, succeeded };
}

export interface PollTicket {
  gen: number;
  signal: AbortSignal;
}

export interface PollSession {
  begin(): PollTicket;
  isCurrent(gen: number): boolean;
  invalidate(): void;
}

/** Serializes monitor polls: a newer begin()/invalidate() makes older tickets stale. */
export function createPollSession(): PollSession {
  let generation = 0;
  let controller: AbortController | null = null;
  return {
    begin() {
      controller?.abort();
      controller = new AbortController();
      generation += 1;
      return { gen: generation, signal: controller.signal };
    },
    isCurrent(gen: number) {
      return gen === generation;
    },
    invalidate() {
      generation += 1;
      controller?.abort();
      controller = null;
    },
  };
}

export function applyLiveSnapshot(prev: Snapshot, live: MonitorLiveSnapshot): MergeResult {
  const next: Snapshot = { ...prev };
  const failed: SnapshotKey[] = [];
  const succeeded: SnapshotKey[] = [];
  const assign = (key: SnapshotKey, value: Snapshot[SnapshotKey] | undefined) => {
    if (value != null) {
      next[key] = value as never;
      succeeded.push(key);
    }
  };
  assign('server', live.server);
  assign('app', live.app);
  assign('database', live.database);
  assign('redis', live.redis);
  assign('api', live.api);
  const errSet = new Set((live.errors || []).filter((k): k is SnapshotKey => (
    k === 'server' || k === 'app' || k === 'database' || k === 'redis' || k === 'api'
  )));
  SNAPSHOT_KEYS.forEach((key) => {
    if (errSet.has(key) && !succeeded.includes(key)) failed.push(key);
  });
  return { next, failed, succeeded };
}

export function applySettledIfCurrent(
  session: Pick<PollSession, 'isCurrent'>,
  gen: number,
  prev: Snapshot,
  results: Array<PromiseSettledResult<Snapshot[SnapshotKey]>>,
  keys: SnapshotKey[] = SNAPSHOT_KEYS,
): MergeResult | null {
  if (!session.isCurrent(gen)) return null;
  return mergeSettledSnapshot(prev, results, keys);
}
