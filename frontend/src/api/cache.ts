import request from './request';
import type { CacheStats, CacheDeleteResult, CacheWarmupRequest, CacheWarmupResult } from '@/types/api';

export function getCacheStats(pattern?: string): Promise<CacheStats> {
  return request.get('/system/cache/stats', { params: { pattern } });
}

export function deleteCacheKey(key: string): Promise<CacheDeleteResult> {
  return request.delete(`/system/cache/${encodeURIComponent(key)}`);
}

export function deleteCachePattern(pattern: string): Promise<CacheDeleteResult> {
  return request.delete(`/system/cache/pattern/${encodeURIComponent(pattern)}`);
}

export function warmupCache(body: CacheWarmupRequest): Promise<CacheWarmupResult> {
  return request.post('/system/cache/warmup', body);
}
