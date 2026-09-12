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

export function isActiveStatus(status: number): boolean {
  return status === 0 || status === 1 || status === 4;
}

export function canRetryRestore(status: number): boolean {
  return status === 2 || status === 5 || status === 6;
}
