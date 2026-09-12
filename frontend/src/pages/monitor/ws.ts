export function buildMonitorWSUrl(
  token: string,
  loc: Pick<Location, 'protocol' | 'host'> = window.location,
  wsBase = import.meta.env.VITE_WS_URL || '/ws',
): string {
  const path = `${wsBase}/monitor`;
  const origin = `${loc.protocol === 'https:' ? 'wss:' : 'ws:'}//${loc.host}`;
  const url = /^wss?:\/\//.test(path) ? path : origin + path;
  return `${url}?token=${encodeURIComponent(token)}`;
}

export interface MonitorWSFrame {
  type: string;
  data?: unknown;
}

export function parseMonitorWSFrame(raw: string): MonitorWSFrame | null {
  try {
    const parsed: unknown = JSON.parse(raw);
    if (typeof parsed === 'object' && parsed !== null && 'type' in parsed && typeof (parsed as MonitorWSFrame).type === 'string') {
      return parsed as MonitorWSFrame;
    }
    return null;
  } catch {
    return null;
  }
}
