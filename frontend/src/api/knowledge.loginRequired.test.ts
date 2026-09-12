import { AxiosError } from 'axios';
import { describe, expect, it } from 'vitest';
import { isKnowledgeLoginRequired } from './knowledge';

function axiosErr(status?: number, code?: number): AxiosError {
  return new AxiosError('x', undefined, undefined, undefined, {
    status,
    data: code == null ? {} : { code },
    statusText: '',
    headers: {},
    config: {} as never,
  });
}

describe('isKnowledgeLoginRequired', () => {
  it('treats 401 and 33004 as login-required', () => {
    expect(isKnowledgeLoginRequired(axiosErr(401))).toBe(true);
    expect(isKnowledgeLoginRequired(axiosErr(403, 33004))).toBe(true);
    expect(isKnowledgeLoginRequired(axiosErr(404, 33001))).toBe(false);
    expect(isKnowledgeLoginRequired(new Error('net'))).toBe(false);
  });
});
