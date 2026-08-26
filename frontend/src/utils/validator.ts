import type { Rule } from 'antd/es/form';

/**
 * 表单校验规则工具
 */

/** 必填规则 */
export const required = (message?: string): Rule => ({
  required: true,
  message: message || '此项为必填',
});

/** 用户名校验（3-20位字母数字下划线） */
export const usernameRule: Rule[] = [
  { required: true, message: '请输入用户名' },
  { min: 3, max: 20, message: '用户名长度为3-20个字符' },
  { pattern: /^[a-zA-Z0-9_]+$/, message: '只能包含字母、数字和下划线' },
];

/** 密码校验（至少6位） */
export const passwordRule: Rule[] = [
  { required: true, message: '请输入密码' },
  { min: 6, message: '密码至少6个字符' },
  {
    pattern: /^(?=.*[a-zA-Z])(?=.*\d).+$/,
    message: '密码必须包含字母和数字',
  },
];

/** 手机号校验 */
export const phoneRule: Rule[] = [
  { pattern: /^1[3-9]\d{9}$/, message: '请输入有效的手机号' },
];

/** 邮箱校验 */
export const emailRule: Rule[] = [
  { type: 'email', message: '请输入有效的邮箱地址' },
];

/** URL 校验 */
export const urlRule: Rule[] = [
  {
    pattern: /^https?:\/\/.+/,
    message: '请输入有效的 URL（以 http:// 或 https:// 开头）',
  },
];

/** 数量校验（正整数） */
export const positiveIntRule: Rule[] = [
  { pattern: /^[1-9]\d*$/, message: '请输入正整数' },
];

/** 金额校验（非负数，最多两位小数） */
export const amountRule: Rule[] = [
  { pattern: /^\d+(\.\d{1,2})?$/, message: '请输入有效金额（最多两位小数）' },
];

/** 身份证号校验（18位） */
export const idCardRule: Rule[] = [
  {
    pattern: /^[1-9]\d{5}(18|19|20)\d{2}(0[1-9]|1[0-2])(0[1-9]|[12]\d|3[01])\d{3}[\dXx]$/,
    message: '请输入有效的身份证号',
  },
];

/** 通用长度校验 */
export const lengthRule = (min: number, max: number, message?: string): Rule => ({
  min,
  max,
  message: message || `长度必须在${min}-${max}个字符之间`,
});

/** 通用正则校验 */
export const patternRule = (regex: RegExp, message: string): Rule => ({
  pattern: regex,
  message,
});
