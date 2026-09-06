import type { FormField, VisibleWhen } from './types';

function toNumber(v: unknown): number | undefined {
  if (typeof v === 'number' && Number.isFinite(v)) return v;
  if (typeof v === 'string' && v.trim() !== '') {
    const n = Number(v);
    return Number.isFinite(n) ? n : undefined;
  }
  return undefined;
}

function stringify(v: unknown): string {
  if (v === null || v === undefined) return '';
  if (typeof v === 'string' || typeof v === 'number' || typeof v === 'boolean') return String(v);
  return JSON.stringify(v);
}

function equals(a: unknown, b: unknown): boolean {
  const na = toNumber(a);
  const nb = toNumber(b);
  if (na !== undefined && nb !== undefined) return na === nb;
  return stringify(a) === stringify(b);
}

function valueIn(actual: unknown, expected: unknown): boolean {
  if (Array.isArray(expected)) {
    return expected.some((item) => equals(actual, item));
  }
  return equals(actual, expected);
}

export function matchVisible(cond: VisibleWhen, values: Record<string, unknown>): boolean {
  const actual = values[cond.field];
  switch (cond.operator) {
    case '==':
      return equals(actual, cond.value);
    case '!=':
      return !equals(actual, cond.value);
    case '>': {
      const a = toNumber(actual);
      const b = toNumber(cond.value);
      return a !== undefined && b !== undefined && a > b;
    }
    case '<': {
      const a = toNumber(actual);
      const b = toNumber(cond.value);
      return a !== undefined && b !== undefined && a < b;
    }
    case 'in':
      return valueIn(actual, cond.value);
    default:
      return true;
  }
}

export function isFieldVisible(field: FormField, values: Record<string, unknown>): boolean {
  if (!field.visible_when) return true;
  return matchVisible(field.visible_when, values);
}
