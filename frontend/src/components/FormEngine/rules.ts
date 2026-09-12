import { tx } from '@/i18n/text';
import type { Rule } from 'antd/es/form';
import type { FormField } from './types';

export function fieldRules(field: FormField): Rule[] {
  const rules: Rule[] = [];
  if (field.required) {
    rules.push({
      required: true,
      message: field.validation?.message || tx('请填写{{value0}}', { value0: field.label }),
    });
  }
  const v = field.validation;
  if (!v) return rules;
  if (field.type === 'text' || field.type === 'textarea') {
    if (v.min_length !== undefined || v.max_length !== undefined) {
      rules.push({
        min: v.min_length,
        max: v.max_length,
        message: v.message || tx('{{value0}}长度不符合要求', { value0: field.label }),
      });
    }
    if (v.pattern) {
      try {
        rules.push({
          pattern: new RegExp(v.pattern),
          message: v.message || tx('{{value0}}格式不正确', { value0: field.label }),
        });
      } catch {
        /* ignore invalid pattern from schema */
      }
    }
  }
  if (field.type === 'number' || field.type === 'rating') {
    if (v.min_value !== undefined) {
      rules.push({
        type: 'number',
        min: v.min_value,
        message: v.message || tx('{{value0}}过小', { value0: field.label }),
      });
    }
    if (v.max_value !== undefined) {
      rules.push({
        type: 'number',
        max: v.max_value,
        message: v.message || tx('{{value0}}过大', { value0: field.label }),
      });
    }
  }
  return rules;
}
