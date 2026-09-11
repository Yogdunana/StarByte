import { describe, expect, it } from 'vitest';
import { toCreateParams, toUpdateParams } from './formPayload';

const base = {
  title: '讲座',
  start_time: '2026-09-11T12:00:00.000Z',
  end_time: '2026-09-11T14:00:00.000Z',
};

describe('toUpdateParams', () => {
  it('sends clear_geo when all fence fields are empty', () => {
    const payload = toUpdateParams({
      ...base,
      latitude: null,
      longitude: null,
      checkin_radius_m: null,
    });
    expect(payload.clear_geo).toBe(true);
    expect(payload.latitude).toBeUndefined();
    expect(payload.longitude).toBeUndefined();
    expect(payload.checkin_radius_m).toBeUndefined();
  });

  it('keeps fence values when they are present', () => {
    const payload = toUpdateParams({
      ...base,
      latitude: 31.23,
      longitude: 121.47,
      checkin_radius_m: 80,
    });
    expect(payload.clear_geo).toBe(false);
    expect(payload.latitude).toBe(31.23);
    expect(payload.longitude).toBe(121.47);
    expect(payload.checkin_radius_m).toBe(80);
  });
});

describe('toCreateParams', () => {
  it('omits empty geo fields', () => {
    const payload = toCreateParams({ ...base, latitude: null });
    expect(payload.latitude).toBeUndefined();
  });
});
