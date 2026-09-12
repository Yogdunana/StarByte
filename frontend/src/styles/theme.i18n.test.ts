import { describe, expect, it } from 'vitest';
import en from '../locales/en-US.json';
import zh from '../locales/zh-CN.json';

const groups = ['login', 'topbar', 'shell', 'empty', 'dashboard'] as const;

function keysOf(value: unknown, prefix = ''): string[] {
  if (!value || typeof value !== 'object') return prefix ? [prefix] : [];
  return Object.entries(value as Record<string, unknown>).flatMap(([key, child]) => {
    const next = prefix ? `${prefix}.${key}` : key;
    return typeof child === 'object' && child !== null ? keysOf(child, next) : [next];
  });
}

describe('shell theme i18n', () => {
  it('keeps zh-CN and en-US keys aligned for shell surfaces', () => {
    groups.forEach((group) => {
      const zhKeys = keysOf(zh[group]).sort();
      const enKeys = keysOf(en[group]).sort();
      expect(enKeys, `en missing or extra keys in ${group}`).toEqual(zhKeys);
      zhKeys.forEach((key) => {
        const zhValue = key.split('.').reduce<unknown>((acc, part) => (acc as Record<string, unknown>)[part], zh[group]);
        const enValue = key.split('.').reduce<unknown>((acc, part) => (acc as Record<string, unknown>)[part], en[group]);
        expect(typeof zhValue, `zh.${group}.${key}`).toBe('string');
        expect(typeof enValue, `en.${group}.${key}`).toBe('string');
        expect(String(zhValue).length, `zh.${group}.${key} empty`).toBeGreaterThan(0);
        expect(String(enValue).length, `en.${group}.${key} empty`).toBeGreaterThan(0);
      });
    });
    expect(zh.common.loadingUser).toBeTruthy();
    expect(en.common.loadingUser).toBeTruthy();
  });
});
