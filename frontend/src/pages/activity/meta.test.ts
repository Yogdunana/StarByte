import { describe, expect, it } from 'vitest';
import { registerSuccessText } from './meta';

describe('registerSuccessText', () => {
  it('shows waitlist copy when status is waitlist', () => {
    expect(registerSuccessText(3)).toBe('已加入候补');
  });

  it('shows success copy when approved', () => {
    expect(registerSuccessText(1)).toBe('报名成功');
  });
});
