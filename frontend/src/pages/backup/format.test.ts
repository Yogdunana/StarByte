import { describe, expect, it } from 'vitest';
import { canRetryRestore, formatBytes, isActiveStatus } from './format';

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
    expect(isActiveStatus(6)).toBe(false);
  });

  it('allows restore retry on success / restored / restore-failed', () => {
    expect(canRetryRestore(2)).toBe(true);
    expect(canRetryRestore(5)).toBe(true);
    expect(canRetryRestore(6)).toBe(true);
    expect(canRetryRestore(3)).toBe(false);
    expect(canRetryRestore(4)).toBe(false);
  });
});
