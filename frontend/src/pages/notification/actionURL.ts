import type { Notification } from '@/types/api';
import { isSafeInternalPath } from '@/utils/nextPath';

export function notificationActionURL(
  notification: Pick<Notification, 'action_url' | 'category' | 'title'>,
): string | undefined {
  const url = notification.action_url;
  // 走全站同一份形态判定，别在这里另写一套。
  // 这里不 decode：action_url 由后端生成（workflowActionURL() 拼的
  // /workflow/todo?task_id=<uuid> 或硬编码的 /member/application），
  // 不是 URL 参数，没有百分号编码需要还原。
  if (url && isSafeInternalPath(url)) return url;
  // Older workflow notifications had no task identifier. Open the inbox, never guess an applicant.
  if (!url && notification.category === 'task' && notification.title === '新通知')
    return '/workflow/todo';
  return undefined;
}
