import { expect, it } from 'vitest';
import { formatCnMobile, isCnMobile, nationalMobileDigits } from './phone';

it('normalizes +86 and separators to 11 digits', () => {
  expect(nationalMobileDigits('+86 138-0013-8000')).toBe('13800138000');
  expect(nationalMobileDigits('8613800138000')).toBe('13800138000');
  expect(isCnMobile('13800138000')).toBe(true);
  expect(isCnMobile('12800138000')).toBe(false);
  expect(isCnMobile('1380013800')).toBe(false);
  expect(formatCnMobile('13800138000')).toBe('+86 138 0013 8000');
});
