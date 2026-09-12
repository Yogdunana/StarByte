import { describe, expect, it } from 'vitest';
import { formatBytes, formatDuration, formatPercent, formatSeconds, pushSample } from './format';

describe('monitor format', () => {
  it('formats bytes', () => {
    expect(formatBytes(0)).toBe('0 B');
    expect(formatBytes(512)).toBe('512 B');
    expect(formatBytes(1024)).toBe('1.0 KB');
    expect(formatBytes(1536)).toBe('1.5 KB');
    expect(formatBytes(1048576)).toBe('1.0 MB');
  });

  it('formats duration', () => {
    expect(formatDuration(12)).toBe('12s');
    expect(formatDuration(80)).toBe('1m 20s');
    expect(formatDuration(3700)).toBe('1h 1m 40s');
    expect(formatDuration(90000)).toBe('1d 1h 0m');
  });

  it('rounds percent and trims samples', () => {
    expect(formatPercent(12.34)).toBe(12.3);
    expect(formatPercent(undefined)).toBe(0);
    expect(pushSample([1, 2], 3, 2)).toEqual([2, 3]);
    expect(formatSeconds(null)).toBe('—');
    expect(formatSeconds(0.0123)).toBe('12.3 ms');
    expect(formatSeconds(1.5)).toBe('1.50 s');
  });
});
