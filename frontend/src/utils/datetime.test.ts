import { describe, expect, it } from 'vitest';
import dayjs from 'dayjs';
import {
  APP_TIMEZONE,
  formatDate,
  formatDateTime,
  formatMinute,
  shanghaiWallClock,
  toAppISO,
} from './datetime';

// 学校里北京时间（UTC+8）和莫斯科时间（UTC+3）都在用。
// 这些用例钉死一条规则：界面上出现的时间一律是北京时间，
// 跟打开页面的电脑在哪个时区无关。
describe('datetime renders in Asia/Shanghai regardless of browser zone', () => {
  // 2026-09-26 12:00 北京时间 == 2026-09-26 04:00 UTC
  const noonShanghai = '2026-09-26T04:00:00.000Z';

  it('formats an instant as Shanghai wall clock', () => {
    expect(APP_TIMEZONE).toBe('Asia/Shanghai');
    expect(formatDateTime(noonShanghai)).toBe('2026-09-26 12:00:00');
    expect(formatDateTime(noonShanghai, 'YYYY-MM-DD HH:mm')).toBe('2026-09-26 12:00');
    expect(formatMinute(noonShanghai)).toBe('2026-09-26 12:00');
    expect(formatDate(noonShanghai)).toBe('2026-09-26');
  });

  it('keeps the date boundary at Shanghai midnight, not UTC midnight', () => {
    // 北京时间 09-26 07:00 == UTC 09-25 23:00。按 UTC 取日界会退一天。
    expect(formatDate('2026-09-25T23:00:00.000Z')).toBe('2026-09-26');
    // 反过来：北京时间 09-26 08:00 == UTC 09-26 00:00
    expect(formatDate('2026-09-26T00:00:00.000Z')).toBe('2026-09-26');
  });

  it('normalizes midnight to 00 instead of 24', () => {
    // 北京时间 2026-09-27 00:30 == UTC 2026-09-26 16:30
    expect(formatDateTime('2026-09-26T16:30:00.000Z', 'HH:mm')).toBe('00:30');
  });

  it('returns a dash for empty or invalid input', () => {
    expect(formatDateTime(undefined)).toBe('-');
    expect(formatDateTime(null)).toBe('-');
    expect(formatDateTime('')).toBe('-');
    expect(formatDateTime('not-a-date')).toBe('-');
  });

  it('exposes the Shanghai wall clock for grouping', () => {
    expect(shanghaiWallClock(noonShanghai)).toBe('2026-09-26T12:00:00');
  });
});

describe('toAppISO treats picker input as Shanghai wall clock', () => {
  it('does not shift the picked time by the browser zone', () => {
    // 无论选择器是在哪个时区创建的 dayjs，挑的 12:00 都应该是北京时间 12:00，
    // 也就是 04:00Z。旧的 dayjs.toISOString() 会按本地时区解释，
    // 莫斯科（UTC+3）的同学挑 12:00 会得到 09:00Z —— 存进库再按北京显示就是 17:00。
    const picked = dayjs('2026-09-26T12:00:00');
    expect(toAppISO(picked)).toBe('2026-09-26T04:00:00.000Z');
  });

  it('round-trips back to the same wall clock', () => {
    const iso = toAppISO(dayjs('2026-09-26T12:00:00'));
    expect(formatDateTime(iso)).toBe('2026-09-26 12:00:00');
  });

  it('ignores non-dayjs values', () => {
    expect(toAppISO(undefined)).toBeUndefined();
    expect(toAppISO(null)).toBeUndefined();
    expect(toAppISO('2026-09-26')).toBeUndefined();
  });
});
