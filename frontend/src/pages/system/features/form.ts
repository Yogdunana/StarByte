import dayjs, { type Dayjs } from 'dayjs';
import type { FeatureFlag, FeatureRules, FeatureType, FeatureVariant } from '@/api/feature';

export const FEATURE_TYPES: FeatureType[] = [
  'boolean', 'user_allowlist', 'role_dept', 'percentage', 'ab_test',
];

export const FEATURE_ENVS = ['dev', 'test', 'prod'] as const;

export interface FlagForm {
  flag_key: string;
  name: string;
  description?: string;
  flag_type: FeatureType;
  enabled?: boolean;
  group_name?: string;
  priority?: number;
  user_ids?: string;
  role_codes?: string;
  department_ids?: string;
  percent?: number;
  salt?: string;
  environments?: string[];
  starts_at?: Dayjs | null;
  ends_at?: Dayjs | null;
  variants?: FeatureVariant[];
}

export function splitLines(raw?: string): string[] {
  return (raw || '').split(/[\n,]/).map((s) => s.trim()).filter(Boolean);
}

export function toRules(v: FlagForm): FeatureRules {
  const starts = v.starts_at && dayjs.isDayjs(v.starts_at) ? v.starts_at.toISOString() : undefined;
  const ends = v.ends_at && dayjs.isDayjs(v.ends_at) ? v.ends_at.toISOString() : undefined;
  return {
    user_ids: splitLines(v.user_ids),
    role_codes: splitLines(v.role_codes),
    department_ids: splitLines(v.department_ids),
    percent: v.percent ?? 0,
    salt: v.salt || '',
    environments: (v.environments || []).map((e) => e.trim()).filter(Boolean),
    starts_at: starts,
    ends_at: ends,
    variants: (v.variants || [])
      .filter((item) => item?.key?.trim())
      .map((item) => ({
        key: item.key.trim(),
        weight: Number(item.weight) || 0,
        enabled: item.enabled,
      })),
  };
}

export function flagToForm(row?: FeatureFlag): Partial<FlagForm> {
  if (!row) {
    return {
      flag_type: 'boolean',
      enabled: false,
      priority: 0,
      environments: [],
      variants: [
        { key: 'control', weight: 50, enabled: false },
        { key: 'treatment', weight: 50, enabled: true },
      ],
    };
  }
  return {
    flag_key: row.flag_key,
    name: row.name,
    description: row.description,
    flag_type: row.flag_type,
    enabled: row.enabled,
    group_name: row.group_name,
    priority: row.priority,
    user_ids: (row.rules.user_ids || []).join('\n'),
    role_codes: (row.rules.role_codes || []).join('\n'),
    department_ids: (row.rules.department_ids || []).join('\n'),
    percent: row.rules.percent,
    salt: row.rules.salt,
    environments: row.rules.environments || [],
    starts_at: row.rules.starts_at ? dayjs(row.rules.starts_at) : null,
    ends_at: row.rules.ends_at ? dayjs(row.rules.ends_at) : null,
    variants: row.rules.variants?.length
      ? row.rules.variants
      : [
        { key: 'control', weight: 50, enabled: false },
        { key: 'treatment', weight: 50, enabled: true },
      ],
  };
}

export function defaultVariants(): FeatureVariant[] {
  return [
    { key: 'control', weight: 50, enabled: false },
    { key: 'treatment', weight: 50, enabled: true },
  ];
}

export function analyticsRowKey(row: { variant?: string; enabled?: boolean }): string {
  return `${row.variant || 'default'}:${row.enabled ? 'on' : 'off'}`;
}
