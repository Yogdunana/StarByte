import { describe, expect, it } from 'vitest';
import { renderMarkdown } from './markdown';

describe('renderMarkdown', () => {
  it('renders headings lists and links', () => {
    const html = renderMarkdown('# 标题\n\n- 一项\n- [章程](/docs/association-charter)\n\n**粗体**');
    expect(html).toContain('<h1>标题</h1>');
    expect(html).toContain('<ul>');
    expect(html).toContain('href="/docs/association-charter"');
    expect(html).toContain('<strong>粗体</strong>');
  });

  it('escapes html and rejects unsafe links', () => {
    const html = renderMarkdown('<script>x</script>\n[x](javascript:alert(1))');
    expect(html).not.toContain('<script>');
    expect(html).toContain('&lt;script&gt;');
    expect(html).not.toContain('javascript:');
  });
});
