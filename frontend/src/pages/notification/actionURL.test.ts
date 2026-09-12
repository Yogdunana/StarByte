import { describe, expect, it } from 'vitest';
import { notificationActionURL } from './actionURL';
describe('notification navigation', () => {
  it('opens the exact workflow task', () => {
    expect(
      notificationActionURL({
        action_url: '/workflow/todo?task_id=123',
        category: 'member',
        title: '入会审批',
      }),
    ).toBe('/workflow/todo?task_id=123');
  });
  it('routes an old generic workflow notification to the inbox without guessing an applicant', () => {
    expect(notificationActionURL({ action_url: '', category: 'task', title: '新通知' })).toBe(
      '/workflow/todo',
    );
    expect(
      notificationActionURL({ action_url: '', category: 'meeting', title: '会议' }),
    ).toBeUndefined();
  });
  it.each(['https://example.com', '//example.com', '/\\example.com', '/\n/example.com'])(
    'rejects unsafe action URL %s',
    (action_url) => {
      expect(
        notificationActionURL({ action_url, category: 'task', title: '通知' }),
      ).toBeUndefined();
    },
  );
});
