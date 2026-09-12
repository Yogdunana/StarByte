import { describe, expect, it } from 'vitest';
import { loginPath, resolveRedirect, sanitizeNext } from './nextPath';

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
