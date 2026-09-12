import { describe, expect, it } from 'vitest';
import en from '../../locales/en-US.json';
import zh from '../../locales/zh-CN.json';

const keys = [
  'title', 'hint', 'viewFlow', 'refresh', 'comment', 'commentRequired', 'updated',
  'approve', 'reject', 'return', 'claim', 'submit', 'start', 'pause', 'resume',
  'approved', 'rejected', 'claimed', 'closed',
  'steps.assignment', 'steps.execution', 'steps.review', 'steps.acceptance', 'steps.completed',
  'waitingAssignee', 'assignmentMode', 'assignmentManual', 'submission', 'submissionEmpty',
] as const;

function pick(obj: Record<string, unknown>, path: string): unknown {
  return path.split('.').reduce<unknown>((acc, key) => {
    if (acc && typeof acc === 'object') {
      return (acc as Record<string, unknown>)[key];
    }
    return undefined;
  }, obj);
}

describe('task engine i18n', () => {
  it('has matching zh-CN and en-US keys', () => {
    keys.forEach((key) => {
      expect(pick(zh.task.engine, key), `zh missing task.engine.${key}`).toBeTruthy();
      expect(pick(en.task.engine, key), `en missing task.engine.${key}`).toBeTruthy();
    });
  });
});
