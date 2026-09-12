import { afterEach, expect, it } from 'vitest';
import i18n from '@/i18n';
import { notificationText } from './localizedText';
import type { Notification } from '@/types/api';
afterEach(async () => {
  await i18n.changeLanguage('zh-CN');
});
it('translates workflow instructions and preserves applicant names and references', async () => {
  await i18n.changeLanguage('ru-RU');
  const record = {
    sender: { id: '', name: '工作流' },
    action_url: '/workflow/todo?task_id=123',
    title: '待办：入会申请 · 干事审批',
    content:
      '事项：入会申请\n当前环节：干事审批\n申请人：张三\n事项编号：ABC-123\n请打开待办详情查看资料并处理。',
  } as Notification;
  expect(notificationText(record, 'title')).toContain('Рассмотрение активом');
  const content = notificationText(record, 'content');
  expect(content).toContain('Заявитель: 张三');
  expect(content).toContain('Номер: ABC-123');
  expect(content).not.toContain('请打开');
  record.sender = { id: 'a-user', name: '工作流' };
  expect(notificationText(record, 'content')).toBe(record.content);
});
