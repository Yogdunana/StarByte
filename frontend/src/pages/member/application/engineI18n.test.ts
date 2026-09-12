import { describe, expect, it } from 'vitest';
import en from '../../../locales/en-US.json';
import zh from '../../../locales/zh-CN.json';

const keys = [
  'title', 'approve', 'reject', 'comment', 'commentRequired', 'approved', 'rejected', 'hint', 'viewFlow',
  'steps.apply', 'steps.minister', 'steps.president', 'steps.end',
] as const;

function pick(obj: Record<string, unknown>, path: string): unknown {
  return path.split('.').reduce<unknown>((acc, key) => {
    if (acc && typeof acc === 'object') {
      return (acc as Record<string, unknown>)[key];
    }
    return undefined;
  }, obj);
}

describe('member engine i18n', () => {
  it('has matching zh-CN and en-US keys', () => {
    keys.forEach((key) => {
      expect(pick(zh.member.engine, key), `zh missing member.engine.${key}`).toBeTruthy();
      expect(pick(en.member.engine, key), `en missing member.engine.${key}`).toBeTruthy();
    });
  });
});
