import { describe, expect, it } from 'vitest';
import { engineReviewClosed } from './engineChain';

describe('engineReviewClosed', () => {
  it('hides actions for approved, rejected, and supplement', () => {
    expect(engineReviewClosed(0)).toBe(false);
    expect(engineReviewClosed(1)).toBe(false);
    expect(engineReviewClosed(3)).toBe(true);
    expect(engineReviewClosed(4)).toBe(true);
    expect(engineReviewClosed(5)).toBe(true);
  });
});
