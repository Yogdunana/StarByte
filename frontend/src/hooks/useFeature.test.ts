import { describe, expect, it } from 'vitest';
import { mapEvaluate } from './useFeature';

describe('mapEvaluate', () => {
  it('maps enabled flags and reasons', () => {
    const got = mapEvaluate({
      'cms.public': { key: 'cms.public', enabled: true, reason: 'allowlist_hit' },
      'membership.portal': { key: 'membership.portal', enabled: false, reason: 'disabled' },
    });
    expect(got.flags['cms.public']).toBe(true);
    expect(got.flags['membership.portal']).toBe(false);
    expect(got.reasons['cms.public']).toBe('allowlist_hit');
  });

  it('fails closed on empty payload', () => {
    expect(mapEvaluate(null).flags).toEqual({});
    expect(mapEvaluate(undefined).flags).toEqual({});
  });
});
