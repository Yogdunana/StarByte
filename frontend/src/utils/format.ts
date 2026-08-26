/**
 * 日期/数字/金额格式化工具
 */

/**
 * 格式化日期时间
 * @param value 日期字符串、时间戳或 Date 对象
 * @param format 输出格式，默认 'YYYY-MM-DD HH:mm:ss'
 * @returns 格式化后的字符串，空值返回 '-'
 */
export function formatDateTime(
  value: string | number | Date | null | undefined,
  format: string = 'YYYY-MM-DD HH:mm:ss',
): string {
  if (!value && value !== 0) return '-';

  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return '-';

  const pad = (n: number) => String(n).padStart(2, '0');

  const map: Record<string, string> = {
    YYYY: String(date.getFullYear()),
    MM: pad(date.getMonth() + 1),
    DD: pad(date.getDate()),
    HH: pad(date.getHours()),
    mm: pad(date.getMinutes()),
    ss: pad(date.getSeconds()),
  };

  return format.replace(/YYYY|MM|DD|HH|mm|ss/g, (match) => map[match]);
}

/**
 * 格式化日期（不含时间）
 */
export function formatDate(value: string | number | Date | null | undefined): string {
  return formatDateTime(value, 'YYYY-MM-DD');
}

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
  if (diff < minute) return '刚刚';
  if (diff < hour) return `${Math.floor(diff / minute)}分钟前`;
  if (diff < day) return `${Math.floor(diff / hour)}小时前`;
  if (diff < week) return `${Math.floor(diff / day)}天前`;
  if (diff < month) return `${Math.floor(diff / week)}周前`;
  return formatDate(value);
}

/**
 * 格式化数字（千分位）
 */
export function formatNumber(value: number | string | null | undefined, decimals = 2): string {
  if (value === null || value === undefined || value === '') return '-';
  const num = typeof value === 'string' ? parseFloat(value) : value;
  if (Number.isNaN(num)) return '-';
  return num.toLocaleString('zh-CN', {
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
  return `¥${num.toLocaleString('zh-CN', {
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
