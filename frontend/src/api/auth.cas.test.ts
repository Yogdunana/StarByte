import { describe, expect, it } from 'vitest';
import { getCasLoginURL } from './auth';

describe('getCasLoginURL', () => {
  it('builds campus CAS start URL with safe redirect', () => {
    expect(getCasLoginURL('/tasks')).toBe('/api/v1/auth/cas/login?redirect=%2Ftasks');
    expect(getCasLoginURL()).toBe('/api/v1/auth/cas/login');
  });
});
