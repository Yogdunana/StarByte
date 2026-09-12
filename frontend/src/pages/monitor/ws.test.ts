import { describe, expect, it } from 'vitest';
import { buildMonitorWSUrl, parseMonitorWSFrame } from './ws';

describe('monitor websocket helpers', () => {
  it('builds a same-origin ws URL', () => {
    const url = buildMonitorWSUrl('tok en', { protocol: 'https:', host: 'app.example' }, '/ws');
    expect(url).toBe('wss://app.example/ws/monitor?token=tok%20en');
  });

  it('keeps an absolute ws base', () => {
    const url = buildMonitorWSUrl('abc', { protocol: 'http:', host: 'localhost:5173' }, 'ws://localhost:8080/ws');
    expect(url).toBe('ws://localhost:8080/ws/monitor?token=abc');
  });

  it('parses snapshot frames and rejects junk', () => {
    expect(parseMonitorWSFrame('{"type":"snapshot","data":{"app":{}}}')?.type).toBe('snapshot');
    expect(parseMonitorWSFrame('{')).toBeNull();
    expect(parseMonitorWSFrame('{"no":"type"}')).toBeNull();
  });
});
