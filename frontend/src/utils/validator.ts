import { tx } from '@/i18n/text';
import type { Rule } from 'antd/es/form';
import { isCnMobile } from '@/utils/phone';

/**
 * 表单校验规则工具
 */

/** 必填规则 */
export const required = (message?: string): Rule => ({
  required: true,
  message: message || tx('此项为必填'),
});

/** 用户名校验（3-20位字母数字下划线） */
export const usernameRule: Rule[] = [
  {
    required: true,
    get message() {
      return tx('请输入用户名');
    },
  },
  {
    min: 3,
    max: 20,
    get message() {
      return tx('用户名长度为3-20个字符');
    },
  },
  {
    pattern: /^[a-zA-Z0-9_]+$/,
    get message() {
      return tx('只能包含字母、数字和下划线');
    },
  },
];

/**
 * 口令是否达标：至少 8 位，且数字、小写字母、大写字母、特殊字符四类里至少占三类。
 *
 * 与后端 pkg/utils.ValidatePasswordStrength 同一套口径，改一边就得改另一边。
 * 只统计可见 ASCII（0x21~0x7e）——中文不算任何一类，否则两段中文就能顶掉一整类。
 */
export const isStrongPassword = (value: string): boolean => {
  if (!value || value.length < 8) return false;
  let hasLower = false;
  let hasUpper = false;
  let hasDigit = false;
  let hasSpecial = false;
  for (let i = 0; i < value.length; i++) {
    const ch = value.charCodeAt(i);
    if (ch >= 0x61 && ch <= 0x7a) hasLower = true;
    else if (ch >= 0x41 && ch <= 0x5a) hasUpper = true;
    else if (ch >= 0x30 && ch <= 0x39) hasDigit = true;
    else if (ch >= 0x21 && ch <= 0x7e) hasSpecial = true;
  }
  return (hasLower ? 1 : 0) + (hasUpper ? 1 : 0) + (hasDigit ? 1 : 0) + (hasSpecial ? 1 : 0) >= 3;
};

/** 设置密码的通用规则（登录框不要用它，登录时不该校验策略，否则老用户进不去） */
export const passwordRule: Rule[] = [
  {
    required: true,
    get message() {
      return tx('请输入密码');
    },
  },
  {
    min: 8,
    get message() {
      return tx('密码至少8个字符');
    },
  },
  {
    validator: async (_, value) => {
      if (!value) return;
      if (!isStrongPassword(String(value))) {
        throw new Error(tx('密码需包含数字、小写字母、大写字母、特殊字符中的至少三种'));
      }
    },
  },
];

/** 手机号校验（中国大陆 11 位，可带 +86） */
export const phoneRule: Rule[] = [
  {
    validator: async (_, value) => {
      if (!value) return;
      if (!isCnMobile(value)) {
        throw new Error(tx('请输入有效的手机号'));
      }
    },
  },
];

/** 邮箱校验 */
export const emailRule: Rule[] = [
  {
    type: 'email',
    get message() {
      return tx('请输入有效的邮箱地址');
    },
  },
];

/** URL 校验 */
export const urlRule: Rule[] = [
  {
    pattern: /^https?:\/\/.+/,
    get message() {
      return tx('请输入有效的 URL（以 http:// 或 https:// 开头）');
    },
  },
];

/** 数量校验（正整数） */
export const positiveIntRule: Rule[] = [
  {
    pattern: /^[1-9]\d*$/,
    get message() {
      return tx('请输入正整数');
    },
  },
];

/** 金额校验（非负数，最多两位小数） */
export const amountRule: Rule[] = [
  {
    pattern: /^\d+(\.\d{1,2})?$/,
    get message() {
      return tx('请输入有效金额（最多两位小数）');
    },
  },
];

/** 身份证号校验（18位） */
export const idCardRule: Rule[] = [
  {
    pattern: /^[1-9]\d{5}(18|19|20)\d{2}(0[1-9]|1[0-2])(0[1-9]|[12]\d|3[01])\d{3}[\dXx]$/,
    get message() {
      return tx('请输入有效的身份证号');
    },
  },
];

/** 通用长度校验 */
export const lengthRule = (min: number, max: number, message?: string): Rule => ({
  min,
  max,
  message: message || tx('长度必须在{{value0}}-{{value1}}个字符之间', { value0: min, value1: max }),
});

/** 通用正则校验 */
export const patternRule = (regex: RegExp, message: string): Rule => ({
  pattern: regex,
  message,
});
