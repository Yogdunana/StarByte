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
  // 反斜杠：WHATWG URL 解析会把 "\evil.test" 规范成 "//evil.test"（协议相对），
  // 于是 "/\evil.test" 就变成了外站跳转。
  if (path.includes('\\')) return '';
  if (path.includes('://')) return '';
  // 控制字符：URL 解析器会把制表符(0x09)、换行(0x0A)、回车(0x0D)从字符串里**删掉**，
  // 所以 "/\t/evil.test" 会被规范成 "//evil.test" —— 同样跳去外站。
  // 这类字符只能在字符串层面先拒掉，指望浏览器/路由库是拦不住的。
  //
  // 刻意不拒 0x20（空格）：它不会被删掉，而是被编码成 %20 留在同源路径里，
  // 实测 new URL('/ dash/x', base) 仍是同源。拒了反而会误伤带空格的正常路径。
  // 后端 internal/auth/service/cas.go 的 sanitizeRedirect 拒的是 <= 0x20，
  // 比这里更严 —— 那是安全方向的差异，不影响正确性。
  for (let i = 0; i < path.length; i += 1) {
    const code = path.charCodeAt(i);
    if (code <= 0x1f || code === 0x7f) return '';
  }
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
