import { describe, expect, it } from 'vitest';
import { resolveTheme } from './preference';

describe('resolveTheme', () => {
  it('follows system when preference is system', () => {
    expect(resolveTheme('system', true)).toBe('dark');
    expect(resolveTheme('system', false)).toBe('light');
  });

  it('honors explicit light and dark', () => {
    expect(resolveTheme('dark', false)).toBe('dark');
    expect(resolveTheme('light', true)).toBe('light');
  });
});
