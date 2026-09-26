import i18n from '@/i18n';
import { tx } from '@/i18n/text';
import { formatDate, formatDateTime } from './datetime';
/**
 * 日期/数字/金额格式化工具
 *
 * 日期时间一律按北京时间渲染（@/utils/datetime），不跟浏览器时区走：
 * 学校里北京时间与莫斯科时间都在用，跟浏览器走会出现同一条记录
 * 在不同电脑上显示不同时间。
 */

export { formatDateTime, formatDate };

/**
 * 格式化相对时间（"3分钟前"、"2小时前"等）
 */
export function formatRelativeTime(value: string | number | Date | null | undefined): string {
  if (!value && value !== 0) return '-';

  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return '-';

  const now = Date.now();
  const diff = now - date.getTime();
  const minute = 60 * 1000;
  const hour = 60 * minute;
  const day = 24 * hour;
  const week = 7 * day;
  const month = 30 * day;

  if (diff < 0) return formatDateTime(value);
  if (diff < minute) return tx('刚刚');
  if (diff < hour) return tx('{{value0}}分钟前', { value0: Math.floor(diff / minute) });
  if (diff < day) return tx('{{value0}}小时前', { value0: Math.floor(diff / hour) });
  if (diff < week) return tx('{{value0}}天前', { value0: Math.floor(diff / day) });
  if (diff < month) return tx('{{value0}}周前', { value0: Math.floor(diff / week) });
  return formatDate(value);
}

/**
 * 格式化数字（千分位）
 */
export function formatNumber(value: number | string | null | undefined, decimals = 2): string {
  if (value === null || value === undefined || value === '') return '-';
  const num = typeof value === 'string' ? parseFloat(value) : value;
  if (Number.isNaN(num)) return '-';
  return num.toLocaleString(i18n.language, {
    minimumFractionDigits: 0,
    maximumFractionDigits: decimals,
  });
}

/**
 * 格式化金额（人民币）
 */
export function formatCurrency(value: number | string | null | undefined, decimals = 2): string {
  if (value === null || value === undefined || value === '') return '¥-';
  const num = typeof value === 'string' ? parseFloat(value) : value;
  if (Number.isNaN(num)) return '¥-';
  return `¥${num.toLocaleString(i18n.language, {
    minimumFractionDigits: decimals,
    maximumFractionDigits: decimals,
  })}`;
}

/**
 * 格式化百分比
 */
export function formatPercent(value: number | null | undefined, decimals = 1): string {
  if (value === null || value === undefined || Number.isNaN(value)) return '-';
  return `${(value * 100).toFixed(decimals)}%`;
}

/**
 * 格式化文件大小
 */
export function formatFileSize(bytes: number | null | undefined): string {
  if (!bytes || bytes === 0) return '0 B';
  const units = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(1024));
  return `${(bytes / Math.pow(1024, i)).toFixed(i > 0 ? 1 : 0)} ${units[i]}`;
}
