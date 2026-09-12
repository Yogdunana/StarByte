import type { TFunction } from 'i18next';
import type { StatusMap } from '@/types/common';
import type { AnnouncementCategory } from '@/api/announcement';

export function announcementStatusMap(t: TFunction): StatusMap {
  return {
    0: { color: 'default', text: t('announcement.status.0') },
    1: { color: 'success', text: t('announcement.status.1') },
    2: { color: 'warning', text: t('announcement.status.2') },
  };
}

export const AnnouncementCategories: { value: AnnouncementCategory; labelKey: string }[] = [
  { value: 'association', labelKey: 'announcement.category.association' },
  { value: 'activity', labelKey: 'announcement.category.activity' },
  { value: 'system', labelKey: 'announcement.category.system' },
  { value: 'personnel', labelKey: 'announcement.category.personnel' },
];

const ALLOWED_TAGS = new Set([
  'P', 'BR', 'STRONG', 'EM', 'B', 'I', 'U', 'UL', 'OL', 'LI',
  'H1', 'H2', 'H3', 'H4', 'H5', 'H6', 'BLOCKQUOTE', 'CODE', 'PRE',
  'A', 'SPAN', 'DIV', 'IMG',
]);

const DROP_WITH_CHILDREN = new Set([
  'SCRIPT', 'IFRAME', 'OBJECT', 'EMBED', 'LINK', 'META', 'STYLE',
  'SVG', 'MATH', 'FORM', 'TEMPLATE', 'NOSCRIPT', 'BASE', 'FRAME', 'FRAMESET',
]);

const ALLOWED_ATTRS: Record<string, Set<string>> = {
  A: new Set(['href', 'title']),
  IMG: new Set(['src', 'alt', 'title']),
};

function isSafeUrl(raw: string): boolean {
  const value = raw.trim();
  if (!value || value.startsWith('#')) return true;
  try {
    const parsed = new URL(value, 'https://starbyte.invalid');
    return parsed.protocol === 'http:' || parsed.protocol === 'https:' || parsed.protocol === 'mailto:';
  } catch {
    return false;
  }
}

function stripNode(el: Element): void {
  if (DROP_WITH_CHILDREN.has(el.tagName)) {
    el.remove();
    return;
  }
  const parent = el.parentNode;
  if (!parent) {
    el.remove();
    return;
  }
  while (el.firstChild) {
    parent.insertBefore(el.firstChild, el);
  }
  el.remove();
}

function sanitizeElement(el: Element): void {
  if (DROP_WITH_CHILDREN.has(el.tagName) || !ALLOWED_TAGS.has(el.tagName)) {
    stripNode(el);
    return;
  }
  const allowed = ALLOWED_ATTRS[el.tagName] ?? new Set<string>();
  for (const attr of Array.from(el.attributes)) {
    const name = attr.name.toLowerCase();
    if (name.startsWith('on') || name === 'style' || name.startsWith('xlink:')) {
      el.removeAttribute(attr.name);
      continue;
    }
    if (!allowed.has(attr.name) && !allowed.has(name)) {
      el.removeAttribute(attr.name);
      continue;
    }
    if ((name === 'href' || name === 'src') && !isSafeUrl(attr.value)) {
      el.removeAttribute(attr.name);
    }
  }
}

/**
 * Allowlist sanitizer for stored announcement HTML.
 * v1 detail page renders markdown/plain text (React-escaped). This remains
 * the only safe path if HTML is ever injected later.
 */
export function sanitizeAnnouncementHTML(html: string): string {
  if (!html) return '';
  const doc = new DOMParser().parseFromString(html, 'text/html');
  const walker = doc.createTreeWalker(doc.body, NodeFilter.SHOW_ELEMENT);
  const seen: Element[] = [];
  let current = walker.nextNode();
  while (current) {
    seen.push(current as Element);
    current = walker.nextNode();
  }
  for (let i = seen.length - 1; i >= 0; i -= 1) {
    sanitizeElement(seen[i]);
  }
  return doc.body.innerHTML;
}
