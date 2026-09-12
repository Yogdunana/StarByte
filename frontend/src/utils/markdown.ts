function escapeHtml(s: string): string {
  return s
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;');
}

function inline(s: string): string {
  let out = escapeHtml(s);
  out = out.replace(/`([^`]+)`/g, '<code>$1</code>');
  out = out.replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>');
  out = out.replace(/\*([^*]+)\*/g, '<em>$1</em>');
  out = out.replace(/\[([^\]]+)\]\(([^)]+)\)/g, (_, text: string, href: string) => {
    const safe = href.trim();
    if (!safe.startsWith('/') && !safe.startsWith('http://') && !safe.startsWith('https://') && !safe.startsWith('mailto:')) {
      return text;
    }
    return `<a href="${escapeHtml(safe)}">${text}</a>`;
  });
  return out;
}

export function renderMarkdown(src: string): string {
  if (!src) return '';
  const lines = src.replace(/\r\n/g, '\n').split('\n');
  const html: string[] = [];
  let i = 0;
  let list: string | null = null;

  const closeList = () => {
    if (list) {
      html.push(`</${list}>`);
      list = null;
    }
  };

  while (i < lines.length) {
    const line = lines[i];
    if (line.startsWith('```')) {
      closeList();
      const buf: string[] = [];
      i += 1;
      while (i < lines.length && !lines[i].startsWith('```')) {
        buf.push(escapeHtml(lines[i]));
        i += 1;
      }
      html.push(`<pre><code>${buf.join('\n')}</code></pre>`);
      i += 1;
      continue;
    }
    const heading = /^(#{1,6})\s+(.+)$/.exec(line);
    if (heading) {
      closeList();
      const n = heading[1].length;
      html.push(`<h${n}>${inline(heading[2])}</h${n}>`);
      i += 1;
      continue;
    }
    if (/^[-*]\s+/.test(line)) {
      if (list !== 'ul') {
        closeList();
        html.push('<ul>');
        list = 'ul';
      }
      html.push(`<li>${inline(line.replace(/^[-*]\s+/, ''))}</li>`);
      i += 1;
      continue;
    }
    if (/^\d+\.\s+/.test(line)) {
      if (list !== 'ol') {
        closeList();
        html.push('<ol>');
        list = 'ol';
      }
      html.push(`<li>${inline(line.replace(/^\d+\.\s+/, ''))}</li>`);
      i += 1;
      continue;
    }
    if (line.trim() === '') {
      closeList();
      i += 1;
      continue;
    }
    closeList();
    html.push(`<p>${inline(line)}</p>`);
    i += 1;
  }
  closeList();
  return html.join('');
}
