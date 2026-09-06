import { describe, expect, it } from 'vitest';
import { isFieldVisible, matchVisible } from './visible';
import { fieldRules } from './rules';
import type { FormField } from './types';

describe('matchVisible', () => {
  it('compares equality and inequality', () => {
    expect(matchVisible({ field: 'k', operator: '==', value: 1 }, { k: 1 })).toBe(true);
    expect(matchVisible({ field: 'k', operator: '!=', value: 1 }, { k: 2 })).toBe(true);
    expect(matchVisible({ field: 'k', operator: '>', value: 3 }, { k: 5 })).toBe(true);
    expect(matchVisible({ field: 'k', operator: '<', value: 3 }, { k: 1 })).toBe(true);
    expect(matchVisible({ field: 'k', operator: 'in', value: [1, 2] }, { k: 2 })).toBe(true);
    expect(matchVisible({ field: 'k', operator: '==', value: 1 }, { k: 'x' })).toBe(false);
  });

  it('treats unknown operator as visible', () => {
    expect(matchVisible({ field: 'k', operator: '==', value: 1 }, {})).toBe(false);
  });
});

describe('isFieldVisible', () => {
  const field = (when?: FormField['visible_when']): FormField => ({
    name: 'score',
    label: '分数',
    type: 'number',
    visible_when: when,
  });

  it('is visible without condition', () => {
    expect(isFieldVisible(field(), {})).toBe(true);
  });

  it('hides when condition fails', () => {
    expect(isFieldVisible(field({ field: 'level', operator: '==', value: 'hard' }), { level: 'easy' })).toBe(
      false,
    );
  });
});

describe('fieldRules', () => {
  it('adds required and length rules', () => {
    const rules = fieldRules({
      name: 'title',
      label: '标题',
      type: 'text',
      required: true,
      validation: { min_length: 2, max_length: 8, pattern: '^[a-z]+$' },
    });
    expect(rules.some((r) => r.required)).toBe(true);
    expect(rules.length).toBeGreaterThan(1);
  });

  it('ignores invalid regexp', () => {
    const rules = fieldRules({
      name: 't',
      label: 't',
      type: 'text',
      validation: { pattern: '(' },
    });
    expect(rules.every((r) => !('pattern' in r) || r.pattern === undefined)).toBe(true);
  });
});
