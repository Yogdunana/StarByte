import { tx } from '@/i18n/text';
import type { Notification } from '@/types/api';

/** Localize built-in system copy; never rewrite names or user-authored messages. */
export function notificationText(record: Notification, part: 'title' | 'content'): string {
  const source = record[part];
  if (record.sender?.id) return source;
  if (record.sender?.name !== '工作流' || !record.action_url?.startsWith('/workflow/todo')) {
    return tx(source);
  }
  if (part === 'title') {
    const match = /^待办：(.*?) · (.*)$/.exec(source);
    return match
      ? tx('待办：{{business}} · {{step}}', { business: tx(match[1]), step: tx(match[2]) })
      : tx(source);
  }
  return source
    .split('\n')
    .map((line) => {
      const match = /^(事项|当前环节|申请人|事项编号)：(.*)$/.exec(line);
      if (!match) return tx(line);
      const value = match[1] === '事项' || match[1] === '当前环节' ? tx(match[2]) : match[2];
      return tx('{{label}}：{{value}}', { label: tx(match[1]), value });
    })
    .join('\n');
}

export function notificationSender(record: Notification): string {
  const name = record.sender?.name || '';
  return record.sender?.id ? name : tx(name);
}
