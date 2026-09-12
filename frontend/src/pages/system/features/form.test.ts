import { describe, expect, it } from 'vitest';
import dayjs from 'dayjs';
import { analyticsRowKey, flagToForm, splitLines, toRules } from './form';
import type { FeatureFlag } from '@/api/feature';

describe('feature form helpers', () => {
  it('splits lines and commas', () => {
    expect(splitLines(' a\nb, c ')).toEqual(['a', 'b', 'c']);
    expect(splitLines('')).toEqual([]);
  });

  it('serializes schedule, env and variants', () => {
    const start = dayjs('2026-09-15T00:00:00Z');
    const rules = toRules({
      flag_key: 'exp.hero',
      name: 'Hero',
      flag_type: 'ab_test',
      percent: 10,
      salt: 'v1',
      user_ids: '11111111-1111-4111-8111-111111111111',
      environments: ['prod', ''],
      starts_at: start,
      ends_at: null,
      variants: [
        { key: 'control', weight: 50, enabled: false },
        { key: ' treatment ', weight: 50, enabled: true },
        { key: '   ', weight: 1 },
      ],
    });
    expect(rules.environments).toEqual(['prod']);
    expect(rules.starts_at).toBe(start.toISOString());
    expect(rules.ends_at).toBeUndefined();
    expect(rules.variants).toEqual([
      { key: 'control', weight: 50, enabled: false },
      { key: 'treatment', weight: 50, enabled: true },
    ]);
    expect(rules.percent).toBe(10);
  });

  it('maps a flag back into the form', () => {
    const row = {
      id: '1',
      flag_key: 'cms.public',
      name: 'CMS',
      description: 'd',
      flag_type: 'percentage',
      enabled: true,
      group_name: 'cms',
      priority: 2,
      rules: {
        percent: 10,
        salt: 's',
        user_ids: ['u1'],
        environments: ['dev'],
        starts_at: '2026-09-15T00:00:00Z',
      },
      is_system: true,
      created_at: '',
      updated_at: '',
    } as FeatureFlag;
    const form = flagToForm(row);
    expect(form.flag_key).toBe('cms.public');
    expect(form.percent).toBe(10);
    expect(form.environments).toEqual(['dev']);
    expect(form.starts_at?.toISOString()).toBe(dayjs('2026-09-15T00:00:00Z').toISOString());
    expect(flagToForm().flag_type).toBe('boolean');
    expect(flagToForm().variants?.length).toBe(2);
  });

  it('keeps boolean analytics buckets unique when variant is empty', () => {
    expect(analyticsRowKey({ variant: '', enabled: true })).toBe('default:on');
    expect(analyticsRowKey({ variant: '', enabled: false })).toBe('default:off');
    expect(analyticsRowKey({ variant: 'treatment', enabled: true })).toBe('treatment:on');
  });
});
