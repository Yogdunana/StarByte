import { describe, expect, it } from 'vitest';
import {
  clearCASRegisterDraft,
  draftFromExchange,
  isStudentIdLikeUsername,
  loadCASRegisterDraft,
  saveCASRegisterDraft,
} from './casRegisterDraft';

describe('casRegisterDraft', () => {
  it('saves and loads a registration continuation', () => {
    saveCASRegisterDraft({
      token: 'tok-1',
      student_no: '20219999',
      real_name: '王五',
      email: 'wang@smbu.edu.cn',
      redirect: '/tasks',
    });
    expect(loadCASRegisterDraft()).toEqual({
      token: 'tok-1',
      student_no: '20219999',
      real_name: '王五',
      email: 'wang@smbu.edu.cn',
      redirect: '/tasks',
    });
    clearCASRegisterDraft();
    expect(loadCASRegisterDraft()).toBeNull();
  });

  it('rejects student-id-like usernames', () => {
    expect(isStudentIdLikeUsername('20219999', '20219999')).toBe(true);
    expect(isStudentIdLikeUsername('20212222', '20219999')).toBe(true);
    expect(isStudentIdLikeUsername('alice_wang', '20219999')).toBe(false);
    expect(isStudentIdLikeUsername('alice', 'alice')).toBe(true);
  });

  it('builds a draft from exchange needs_registration payload', () => {
    const draft = draftFromExchange({
      registration_token: 'abc',
      student_no: '20210001',
      real_name: '张三',
      redirect: '/dashboard',
    });
    expect(draft?.token).toBe('abc');
    expect(draft?.student_no).toBe('20210001');
    expect(draftFromExchange({})).toBeNull();
  });
});
