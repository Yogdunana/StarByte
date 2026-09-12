import { describe, expect, it } from 'vitest';
import { formatBytes, isActiveStatus } from './format';

describe('backup format', () => {
  it('formats bytes', () => {
    expect(formatBytes(0)).toBe('0 B');
    expect(formatBytes(2048)).toBe('2.0 KB');
  });

  it('detects in-flight statuses', () => {
    expect(isActiveStatus(0)).toBe(true);
    expect(isActiveStatus(1)).toBe(true);
    expect(isActiveStatus(4)).toBe(true);
    expect(isActiveStatus(2)).toBe(false);
    expect(isActiveStatus(3)).toBe(false);
  });
});
