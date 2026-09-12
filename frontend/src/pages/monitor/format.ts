export function formatBytes(n: number | undefined): string {
  const value = Number(n || 0);
  if (value <= 0) return '0 B';
  const units = ['B', 'KB', 'MB', 'GB', 'TB'];
  let v = value;
  let i = 0;
  while (v >= 1024 && i < units.length - 1) {
    v /= 1024;
    i += 1;
  }
  return `${v.toFixed(i === 0 ? 0 : 1)} ${units[i]}`;
}

export function formatDuration(seconds: number | undefined): string {
  const s = Math.max(0, Math.floor(Number(seconds || 0)));
  const d = Math.floor(s / 86400);
  const h = Math.floor((s % 86400) / 3600);
  const m = Math.floor((s % 3600) / 60);
  const rest = s % 60;
  if (d > 0) return `${d}d ${h}h ${m}m`;
  if (h > 0) return `${h}h ${m}m ${rest}s`;
  if (m > 0) return `${m}m ${rest}s`;
  return `${rest}s`;
}

export function formatPercent(n: number | undefined): number {
  const v = Number(n || 0);
  if (Number.isNaN(v)) return 0;
  return Math.round(v * 10) / 10;
}

export function pushSample<T>(list: T[], item: T, max = 20): T[] {
  const next = [...list, item];
  return next.length > max ? next.slice(next.length - max) : next;
}
