import { useCallback, useEffect, useState } from 'react';
import { getAnnouncementUnreadCount } from '@/api/announcement';
import { useFeature } from '@/hooks/useFeature';
import { usePermission } from '@/hooks/usePermission';

/** 顶栏 / 列表页未读公告数。有 announcement:read 且命中灰度时轮询。 */
export function useAnnouncementUnread(pollMs = 60000) {
  const canRead = usePermission('announcement:read');
  const feed = useFeature('announcement.feed');
  const allowed = canRead && feed.enabled;
  const [count, setCount] = useState(0);

  const refresh = useCallback(async () => {
    if (!allowed) {
      setCount(0);
      return;
    }
    try {
      const res = await getAnnouncementUnreadCount();
      setCount(res.count);
    } catch {
      // 列表页/顶栏失败时保持上次计数，避免闪 0
    }
  }, [allowed]);

  useEffect(() => {
    void refresh();
    if (!allowed || pollMs <= 0) return undefined;
    const timer = window.setInterval(() => {
      void refresh();
    }, pollMs);
    return () => window.clearInterval(timer);
  }, [allowed, pollMs, refresh]);

  return { count, refresh, canRead: allowed };
}
