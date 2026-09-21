import { describe, expect, it } from 'vitest';
import { loginPath, resolveRedirect, sanitizeNext } from './nextPath';

describe('sanitizeNext 防外站跳转', () => {
  it('浏览器会把 URL 里的控制字符删掉（这就是必须拒它们的理由）', () => {
    // WHATWG URL 解析会移除 tab/LF/CR，于是 "/<tab>/evil.test" 变成 "//evil.test"，
    // 即协议相对 URL —— 直接跳到外站。这个断言把该行为固定下来，
    // 免得以后有人觉得"多此一举"把这些检查删掉。
    expect(new URL('/\t/evil.test', 'https://starbyte.test/').href).toBe('https://evil.test/');
    expect(new URL('/\n/evil.test', 'https://starbyte.test/').href).toBe('https://evil.test/');
    expect(new URL('/\r/evil.test', 'https://starbyte.test/').href).toBe('https://evil.test/');
    expect(new URL('/\\evil.test', 'https://starbyte.test/').href).toBe('https://evil.test/');
  });

  it('拒绝控制字符与百分号编码的控制字符', () => {
    expect(sanitizeNext('/\t/evil.test')).toBe('');
    expect(sanitizeNext('/\n/evil.test')).toBe('');
    expect(sanitizeNext('/\r/evil.test')).toBe('');
    expect(sanitizeNext('/\u0001/evil.test')).toBe('');
    expect(sanitizeNext('/%09/evil.test')).toBe('');
    expect(sanitizeNext('/%0A/evil.test')).toBe('');
  });

  it('拒绝反斜杠与 ://', () => {
    expect(sanitizeNext('/\\evil.test')).toBe('');
    expect(sanitizeNext('/foo://bar')).toBe('');
  });

  it('空格不会被当成控制字符拒掉（它只会被编码成 %20 留在同源）', () => {
    expect(new URL('/dash board/x', 'https://starbyte.test/').host).toBe('starbyte.test');
    expect(sanitizeNext('/dash board/x')).toBe('/dash board/x');
  });

  it('正常站内路径不受影响', () => {
    expect(sanitizeNext('/tasks/123')).toBe('/tasks/123');
    expect(sanitizeNext('/docs/api?tab=go#sec')).toBe('/docs/api?tab=go#sec');
  });
});

describe('sanitizeNext', () => {
  it('accepts internal docs paths', () => {
    expect(sanitizeNext('/docs/user-manual')).toBe('/docs/user-manual');
    expect(sanitizeNext('/about-us')).toBe('/about-us');
  });

  it('rejects external or login loops', () => {
    expect(sanitizeNext('https://evil.test')).toBe('');
    expect(sanitizeNext('//evil.test')).toBe('');
    expect(sanitizeNext('/login')).toBe('');
    expect(sanitizeNext('/')).toBe('');
  });
});

describe('loginPath', () => {
  it('encodes next', () => {
    expect(loginPath('/docs/api-manual')).toBe('/?next=%2Fdocs%2Fapi-manual');
    expect(loginPath('/')).toBe('/');
  });
});

describe('resolveRedirect', () => {
  it('falls back to dashboard', () => {
    expect(resolveRedirect(null)).toBe('/dashboard');
    expect(resolveRedirect('/knowledge')).toBe('/knowledge');
  });
});
