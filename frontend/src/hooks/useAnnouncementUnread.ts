import { useCallback, useEffect, useState } from 'react';
import { getAnnouncementUnreadCount } from '@/api/announcement';
import { usePermission } from '@/hooks/usePermission';

/** 顶栏 / 列表页未读公告数。有 announcement:read 时轮询。 */
export function useAnnouncementUnread(pollMs = 60000) {
  const canRead = usePermission('announcement:read');
  const [count, setCount] = useState(0);

  const refresh = useCallback(async () => {
    if (!canRead) {
      setCount(0);
      return;
    }
    try {
      const res = await getAnnouncementUnreadCount();
      setCount(res.count);
    } catch {
      // 列表页/顶栏失败时保持上次计数，避免闪 0
    }
  }, [canRead]);

  useEffect(() => {
    void refresh();
    if (!canRead || pollMs <= 0) return undefined;
    const timer = window.setInterval(() => {
      void refresh();
    }, pollMs);
    return () => window.clearInterval(timer);
  }, [canRead, pollMs, refresh]);

  return { count, refresh, canRead };
}
