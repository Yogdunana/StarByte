import type { Notification } from '@/types/api';

export function notificationActionURL(
  notification: Pick<Notification, 'action_url' | 'category' | 'title'>,
): string | undefined {
  const url = notification.action_url;
  if (
    url?.startsWith('/') &&
    !url.startsWith('//') &&
    !url.includes('\\') &&
    !Array.from(url).some((char) => char.charCodeAt(0) < 32)
  )
    return url;
  // Older workflow notifications had no task identifier. Open the inbox, never guess an applicant.
  if (!url && notification.category === 'task' && notification.title === '新通知')
    return '/workflow/todo';
  return undefined;
}
