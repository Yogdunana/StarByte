import request from './request';

export interface MonitorServer {
  cpu_percent: number;
  mem_used: number;
  mem_total: number;
  mem_percent: number;
  disk_used: number;
  disk_total: number;
  disk_percent: number;
  disk_path: string;
  load1: number;
  load5: number;
  load15: number;
  goos: string;
  goarch: string;
  collected_at: string;
}

export interface MonitorApp {
  started_at: string;
  uptime_seconds: number;
  version: string;
  go_version: string;
  goroutines: number;
  mem_alloc: number;
  mem_sys: number;
  heap_alloc: number;
  heap_sys: number;
  heap_inuse: number;
  stack_inuse: number;
  num_gc: number;
  next_gc: number;
  last_gc_unix_ms: number;
  pause_total_ns: number;
  gc_cpu_fraction: number;
  collected_at: string;
}

export interface MonitorDatabase {
  available: boolean;
  open_connections: number;
  in_use: number;
  idle: number;
  wait_count: number;
  wait_duration_ms: number;
  max_open_connections: number;
  max_idle_closed: number;
  max_lifetime_closed: number;
  collected_at: string;
}

export interface MonitorRedis {
  available: boolean;
  connected_clients: number;
  used_memory: number;
  max_memory: number;
  keyspace_hits: number;
  keyspace_misses: number;
  hit_rate: number;
  keys: number;
  version: string;
  collected_at: string;
}

export interface MonitorAPIStats {
  available: boolean;
  source: string;
  request_total: number;
  error_total: number;
  error_rate: number;
  p50_seconds: number | null;
  p95_seconds: number | null;
  p99_seconds: number | null;
  percentiles_note: string;
  collected_at: string;
}

export function getMonitorServer(): Promise<MonitorServer> {
  return request.get('/monitor/server');
}

export function getMonitorApp(): Promise<MonitorApp> {
  return request.get('/monitor/app');
}

export function getMonitorDatabase(): Promise<MonitorDatabase> {
  return request.get('/monitor/database');
}

export function getMonitorRedis(): Promise<MonitorRedis> {
  return request.get('/monitor/redis');
}

export function getMonitorAPIStats(): Promise<MonitorAPIStats> {
  return request.get('/monitor/api-stats');
}
