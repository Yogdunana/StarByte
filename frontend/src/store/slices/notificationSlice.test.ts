import { describe, expect, it } from 'vitest';
import reducer, { clearNotifications, fetchRecentNotifications, fetchUnreadCount } from './notificationSlice';

describe('notification request ownership', () => {
  it('ignores an older count response after a newer request', () => {
    let state = reducer(undefined, fetchUnreadCount.pending('old', undefined));
    state = reducer(state, fetchUnreadCount.pending('new', undefined));
    state = reducer(state, fetchUnreadCount.fulfilled(3, 'new', undefined));
    state = reducer(state, fetchUnreadCount.fulfilled(99, 'old', undefined));
    expect(state.unreadCount).toBe(3);
  });
  it('does not expose a prior session notification when logout races a response', () => {
    let state = reducer(undefined, fetchRecentNotifications.pending('old-session', undefined));
    state = reducer(state, clearNotifications());
    state = reducer(state, fetchRecentNotifications.fulfilled([{ id: 'private', title: 'Private', content: '', category: 'system', priority: 'normal', is_read: false, action_url: '', sender: { id: '', name: '' }, created_at: '' }], 'old-session', undefined));
    expect(state.recentNotifications).toEqual([]);
  });
});
