/** Strip +86 / 86 / separators and keep digits. */
export function nationalMobileDigits(raw?: string | null): string {
  const digits = String(raw ?? '').replace(/\D/g, '');
  if (digits.startsWith('0086') && digits.length >= 15) return digits.slice(4);
  if (digits.startsWith('86') && digits.length >= 13) return digits.slice(2);
  return digits;
}

/** Mainland China mobile: 1[3-9] + 9 digits. */
export function isCnMobile(raw?: string | null): boolean {
  return /^1[3-9]\d{9}$/.test(nationalMobileDigits(raw));
}

export function formatCnMobile(raw?: string | null): string {
  const digits = nationalMobileDigits(raw);
  if (!isCnMobile(digits)) return (raw ?? '').trim();
  return `+86 ${digits.slice(0, 3)} ${digits.slice(3, 7)} ${digits.slice(7)}`;
}
