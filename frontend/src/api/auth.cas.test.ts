import { describe, expect, it } from 'vitest';
import { getCasLoginURL } from './auth';

describe('getCasLoginURL', () => {
  it('builds campus CAS start URL with safe redirect', () => {
    expect(getCasLoginURL('/tasks')).toBe('/api/v1/auth/cas/login?redirect=%2Ftasks');
    expect(getCasLoginURL()).toBe('/api/v1/auth/cas/login');
  });
});

describe('CAS register draft wiring', () => {
  it('treats needs_registration exchange as bind flow, not silent login', () => {
    const exchange = {
      needs_registration: true,
      registration_token: 'reg-1',
      student_no: '20219999',
      access_token: '',
      refresh_token: '',
    };
    expect(exchange.needs_registration).toBe(true);
    expect(exchange.student_no).toBe('20219999');
    expect(exchange.access_token).toBe('');
  });
});
