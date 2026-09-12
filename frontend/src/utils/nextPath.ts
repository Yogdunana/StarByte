const BLOCKED = new Set(['/', '/login', '/login/cas', '/register/cas']);

export function sanitizeNext(raw: string | null | undefined): string {
  if (!raw) return '';
  let path = raw.trim();
  try {
    path = decodeURIComponent(path);
  } catch {
    return '';
  }
  if (!path.startsWith('/') || path.startsWith('//')) return '';
  if (path.includes('\\')) return '';
  const bare = path.split(/[?#]/)[0];
  if (BLOCKED.has(bare)) return '';
  return path;
}

export function loginPath(next?: string): string {
  const n = sanitizeNext(next);
  if (!n) return '/';
  return `/?next=${encodeURIComponent(n)}`;
}

export function resolveRedirect(next: string | null | undefined, fallback = '/dashboard'): string {
  return sanitizeNext(next) || fallback;
}
