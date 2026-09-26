import dayjs from 'dayjs';

/**
 * 系统业务时区。
 *
 * 深圳北理莫斯科大学是中俄合办学校，同学和老师的电脑时区并不统一：
 * 北京时间 UTC+8、莫斯科时间 UTC+3 都常见。之前前端用 dayjs.toISOString() 提交
 * （把用户选的墙钟按浏览器本地时区解释），又用 new Date().getHours() 渲染
 * （同样按浏览器本地时区），结果是**同一条记录在两个人屏幕上显示两个时间**，
 * 跨时区的人还会看到自己填的 12:00 变成 17:00。
 *
 * 收口成一条规则：**界面上出现的所有日期时间，一律是北京时间**。
 * - 渲染：不管浏览器在哪，都按 Asia/Shanghai 取年月日时分
 * - 输入：用户在选择器里挑的墙钟，按 Asia/Shanghai 解释成绝对时刻再提交
 * 这样「选 12:00 就是北京时间 12:00」，所见即所得。
 *
 * 后端 leave 模块早就是这么定的（internal/leave/service/duration.go 的
 * bizTimezone = "Asia/Shanghai"），这里只是把前端对齐到同一口径。
 */
export const APP_TIMEZONE = 'Asia/Shanghai';

/** Asia/Shanghai 没有夏令时，全年恒为 UTC+8，所以可以写死。 */
export const APP_UTC_OFFSET = '+08:00';

/**
 * 取某个时刻在北京时间下的年月日时分秒。
 * 用 Intl 而不是 getHours()，避免跟着浏览器时区跑。
 */
function shanghaiParts(value: string | number | Date): Record<string, string> | null {
  const date = value instanceof Date ? value : new Date(value);
  if (Number.isNaN(date.getTime())) return null;
  const parts = new Intl.DateTimeFormat('en-CA', {
    timeZone: APP_TIMEZONE,
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    hour12: false,
  }).formatToParts(date);
  const get = (type: string) => parts.find((p) => p.type === type)?.value ?? '00';
  return {
    YYYY: get('year'),
    MM: get('month'),
    DD: get('day'),
    // hour12:false 在部分 ICU 上会把午夜给成 24，归一成 00。
    HH: get('hour') === '24' ? '00' : get('hour'),
    mm: get('minute'),
    ss: get('second'),
  };
}

/**
 * 格式化日期时间，固定按北京时间输出。
 * @param value 日期字符串、时间戳或 Date 对象
 * @param format 输出格式，默认 'YYYY-MM-DD HH:mm:ss'
 */
export function formatDateTime(
  value: string | number | Date | null | undefined,
  format: string = 'YYYY-MM-DD HH:mm:ss',
): string {
  if (!value && value !== 0) return '-';
  const map = shanghaiParts(value);
  if (!map) return '-';
  return format.replace(/YYYY|MM|DD|HH|mm|ss/g, (m) => map[m]);
}

/** 格式化日期（不含时间），按北京时间取日界。 */
export function formatDate(value: string | number | Date | null | undefined): string {
  return formatDateTime(value, 'YYYY-MM-DD');
}

/** 只取时分，按北京时间。 */
export function formatTime(value: string | number | Date | null | undefined): string {
  return formatDateTime(value, 'HH:mm');
}

/** 精确到分，按北京时间。 */
export function formatMinute(value: string | number | Date | null | undefined): string {
  return formatDateTime(value, 'YYYY-MM-DD HH:mm');
}

/**
 * 返回某个时刻在东八区的「墙上时间」YYYY-MM-DDTHH:mm:ss（不带时区标记）。
 * 供需要按北京时间做日期比较 / 分组的地方用。
 */
export function shanghaiWallClock(value: string | number | Date): string | undefined {
  const map = shanghaiParts(value);
  if (!map) return undefined;
  return `${map.YYYY}-${map.MM}-${map.DD}T${map.HH}:${map.mm}:${map.ss}`;
}

/**
 * 把 dayjs 选择器选出来的墙钟按北京时间解释，转成给后端的 ISO 字符串。
 *
 * 用来替换 dayjs.toISOString()：那个按浏览器本地时区解释，
 * 莫斯科的同学选 12:00 会变成 09:00Z，存进库再按北京显示就是 17:00。
 *
 * 这里只取年月日时分秒（丢掉浏览器时区），拼上 +08:00 再转绝对时刻。
 */
export function toAppISO(value: unknown): string | undefined {
  if (!value || !dayjs.isDayjs(value)) return undefined;
  const d = value as ReturnType<typeof dayjs>;
  const wall = `${d.format('YYYY-MM-DD')}T${d.format('HH:mm:ss')}${APP_UTC_OFFSET}`;
  const date = new Date(wall);
  if (Number.isNaN(date.getTime())) return undefined;
  return date.toISOString();
}
