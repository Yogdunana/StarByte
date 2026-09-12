import { describe, expect, it } from 'vitest';
import { motionDataset, resolveMotionMode } from './preference';

describe('motion preference', () => {
  it('forces Motion to skip when the user toggle is on', () => {
    expect(resolveMotionMode(true)).toBe('always');
    expect(motionDataset(true)).toBe('reduced');
  });

  it('defers to prefers-reduced-motion when the user toggle is off', () => {
    expect(resolveMotionMode(false)).toBe('user');
    expect(motionDataset(false)).toBe('system');
  });
});
