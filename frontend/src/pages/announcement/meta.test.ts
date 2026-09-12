import { describe, expect, it, vi } from 'vitest';
import { collectPagedItems, expireUpdateFields, sanitizeAnnouncementHTML } from './meta';

describe('sanitizeAnnouncementHTML', () => {
  it('keeps safe markup and strips event handlers', () => {
    const dirty = '<p onclick="alert(1)">ok</p><script>alert(2)</script><iframe src="x"></iframe>';
    const clean = sanitizeAnnouncementHTML(dirty);
    expect(clean).toContain('<p>ok</p>');
    expect(clean.toLowerCase()).not.toContain('script');
    expect(clean.toLowerCase()).not.toContain('iframe');
    expect(clean.toLowerCase()).not.toContain('onclick');
  });

  it('drops svg onload payloads', () => {
    const dirty = '<p>safe</p><svg/onload=alert(1)><circle /></svg>';
    const clean = sanitizeAnnouncementHTML(dirty);
    expect(clean).toContain('safe');
    expect(clean.toLowerCase()).not.toContain('svg');
    expect(clean.toLowerCase()).not.toContain('onload');
  });

  it('keeps http images and drops javascript src', () => {
    const dirty = '<img src="https://cdn.example/a.png" alt="ok" /><img src="javascript:alert(1)" />';
    const clean = sanitizeAnnouncementHTML(dirty);
    expect(clean).toContain('https://cdn.example/a.png');
    expect(clean.toLowerCase()).not.toContain('javascript:');
  });

  it('collects every page until total is reached', async () => {
    const fetchPage = vi.fn(async (page: number) => {
      if (page === 1) return { list: [{ id: 'a' }, { id: 'b' }], total: 3 };
      return { list: [{ id: 'c' }], total: 3 };
    });
    await expect(collectPagedItems(fetchPage, 2)).resolves.toEqual([
      { id: 'a' }, { id: 'b' }, { id: 'c' },
    ]);
    expect(fetchPage).toHaveBeenCalledTimes(2);
  });

  it('omits unchanged expires_at so past unpublish does not block edits', () => {
    const prev = '2026-09-12T09:00:00.000Z';
    expect(expireUpdateFields(prev, '2026-09-12T09:00:00.000Z')).toEqual({
      expires_at: undefined,
      clear_expires_at: false,
    });
    expect(expireUpdateFields(prev, '2026-09-13T09:00:00.000Z')).toEqual({
      expires_at: '2026-09-13T09:00:00.000Z',
      clear_expires_at: false,
    });
    expect(expireUpdateFields(prev, undefined)).toEqual({
      expires_at: undefined,
      clear_expires_at: true,
    });
  });

  it('strips javascript: URLs', () => {
    const dirty = '<a href="javascript:alert(1)">click</a><a href="https://example.com">ok</a>';
    const clean = sanitizeAnnouncementHTML(dirty);
    expect(clean.toLowerCase()).not.toContain('javascript:');
    expect(clean).toContain('https://example.com');
    expect(clean).toContain('ok');
  });
});
