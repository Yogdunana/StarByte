export const CAS_REGISTER_STORAGE_KEY = 'starbyte_cas_register';

export interface CASRegisterDraft {
  token: string;
  student_no: string;
  real_name?: string;
  email?: string;
  redirect?: string;
}

export function saveCASRegisterDraft(draft: CASRegisterDraft): void {
  if (typeof sessionStorage === 'undefined') return;
  sessionStorage.setItem(CAS_REGISTER_STORAGE_KEY, JSON.stringify(draft));
}

export function loadCASRegisterDraft(): CASRegisterDraft | null {
  if (typeof sessionStorage === 'undefined') return null;
  const raw = sessionStorage.getItem(CAS_REGISTER_STORAGE_KEY);
  if (!raw) return null;
  try {
    const parsed = JSON.parse(raw) as Partial<CASRegisterDraft>;
    if (!parsed || typeof parsed.token !== 'string' || !parsed.token.trim()) {
      return null;
    }
    return {
      token: parsed.token.trim(),
      student_no: typeof parsed.student_no === 'string' ? parsed.student_no : '',
      real_name: typeof parsed.real_name === 'string' ? parsed.real_name : '',
      email: typeof parsed.email === 'string' ? parsed.email : '',
      redirect: typeof parsed.redirect === 'string' ? parsed.redirect : '/dashboard',
    };
  } catch {
    return null;
  }
}

export function clearCASRegisterDraft(): void {
  if (typeof sessionStorage === 'undefined') return;
  sessionStorage.removeItem(CAS_REGISTER_STORAGE_KEY);
}

export function isStudentIdLikeUsername(username: string, studentNo?: string): boolean {
  const value = username.trim();
  if (!value) return false;
  const ownStudentNo = studentNo?.trim() ?? '';
  if (ownStudentNo && value === ownStudentNo) return true;
  let digits = 0;
  for (const ch of value) {
    if (ch >= '0' && ch <= '9') digits += 1;
  }
  return digits >= Math.floor((value.length * 3) / 4);
}

export function draftFromExchange(result: {
  registration_token?: string;
  student_no?: string;
  real_name?: string;
  email?: string;
  redirect?: string;
}): CASRegisterDraft | null {
  const token = result.registration_token?.trim() ?? '';
  if (!token) return null;
  return {
    token,
    student_no: result.student_no?.trim() ?? '',
    real_name: result.real_name?.trim() ?? '',
    email: result.email?.trim() ?? '',
    redirect: result.redirect ?? '/dashboard',
  };
}
