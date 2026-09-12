import { describe, expect, it } from 'vitest';
import en from '../../locales/en-US.json';
import zh from '../../locales/zh-CN.json';

function keysOf(value: unknown, prefix = ''): string[] {
  if (!value || typeof value !== 'object') return prefix ? [prefix] : [];
  return Object.entries(value as Record<string, unknown>).flatMap(([key, child]) => {
    const next = prefix ? `${prefix}.${key}` : key;
    return typeof child === 'object' && child !== null ? keysOf(child, next) : [next];
  });
}

describe('leave i18n', () => {
  it('has matching zh-CN and en-US keys', () => {
    const zhKeys = keysOf(zh.leave).sort();
    const enKeys = keysOf(en.leave).sort();
    expect(enKeys).toEqual(zhKeys);
    zhKeys.forEach((key) => {
      const zhValue = key.split('.').reduce<unknown>((acc, part) => (acc as Record<string, unknown>)[part], zh.leave);
      const enValue = key.split('.').reduce<unknown>((acc, part) => (acc as Record<string, unknown>)[part], en.leave);
      expect(String(zhValue).length, `zh.leave.${key}`).toBeGreaterThan(0);
      expect(String(enValue).length, `en.leave.${key}`).toBeGreaterThan(0);
    });
  });
});
