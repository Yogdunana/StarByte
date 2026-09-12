import { describe, expect, it } from 'vitest';
import { fadeRight, fadeUp } from './tokens';

function keysOf(variant: object): string[] {
  return Object.keys(variant).sort();
}

describe('motion tokens', () => {
  it('only animates opacity and transform axes', () => {
    expect(keysOf(fadeUp.hidden as object)).toEqual(['opacity', 'y']);
    expect(keysOf(fadeUp.show as object).filter((key) => key !== 'transition').sort()).toEqual(['opacity', 'y']);
    expect(keysOf(fadeRight.hidden as object)).toEqual(['opacity', 'x']);
    expect(keysOf(fadeRight.show as object).filter((key) => key !== 'transition').sort()).toEqual(['opacity', 'x']);
  });
});
