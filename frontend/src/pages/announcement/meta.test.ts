import { describe, expect, it } from 'vitest';
import { sanitizeAnnouncementHTML } from './meta';

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

  it('strips javascript: URLs', () => {
    const dirty = '<a href="javascript:alert(1)">click</a><a href="https://example.com">ok</a>';
    const clean = sanitizeAnnouncementHTML(dirty);
    expect(clean.toLowerCase()).not.toContain('javascript:');
    expect(clean).toContain('https://example.com');
    expect(clean).toContain('ok');
  });
});
