import { describe, expect, it } from 'vitest';
import en from '../../locales/en-US.json';
import zh from '../../locales/zh-CN.json';

const keys = [
  'title', 'hint', 'refresh', 'autoRefresh', 'updatedAt', 'loadFail', 'loadPartial',
  'partSep', 'cardUnavailable', 'server', 'cpu', 'memory', 'disk', 'cpuTrend', 'load',
  'hostMeta', 'memDetail', 'diskDetail', 'app', 'uptime', 'goroutines', 'heap', 'numGC',
  'database', 'dbDown', 'dbInUse', 'dbIdle', 'dbOpen', 'dbWait', 'dbMax', 'dbWaitMs',
  'redis', 'redisDown', 'redisClients', 'redisMemory', 'redisHitRate', 'redisKeys',
  'redisHits', 'redisMisses', 'api', 'apiTotal', 'apiErrors', 'apiErrorRate', 'apiNote',
  'apiP50', 'apiP95', 'apiP99', 'liveOn', 'liveOff', 'slowQueries', 'slowSource',
  'slowEmpty', 'slowQuery', 'slowCalls', 'slowMean', 'slowMax', 'slowRows', 'redisSlow',
] as const;

function pick(obj: Record<string, unknown>, path: string): unknown {
  return path.split('.').reduce<unknown>((acc, key) => {
    if (acc && typeof acc === 'object') {
      return (acc as Record<string, unknown>)[key];
    }
    return undefined;
  }, obj);
}

describe('monitor i18n', () => {
  it('has matching zh-CN and en-US keys', () => {
    keys.forEach((key) => {
      expect(pick(zh.monitor, key), `zh missing monitor.${key}`).toBeTruthy();
      expect(pick(en.monitor, key), `en missing monitor.${key}`).toBeTruthy();
    });
    expect(zh.menu['/monitor']).toBeTruthy();
    expect(en.menu['/monitor']).toBeTruthy();
  });
});
