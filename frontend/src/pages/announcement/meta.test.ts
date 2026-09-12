import { describe, expect, it } from 'vitest';
import { sanitizeAnnouncementHTML } from './meta';

describe('sanitizeAnnouncementHTML', () => {
  it('strips script iframe and event handlers', () => {
    const dirty = '<p onclick="alert(1)">ok</p><script>alert(2)</script><iframe src="x"></iframe>';
    const clean = sanitizeAnnouncementHTML(dirty);
    expect(clean).toContain('<p>ok</p>');
    expect(clean).not.toContain('script');
    expect(clean).not.toContain('iframe');
    expect(clean).not.toContain('onclick');
  });
});
