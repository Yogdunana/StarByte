/**
 * 名单粘贴/上传的解析。
 *
 * 测评系统导出来的东西可能是 xlsx 另存的 CSV，也可能是从表格里直接框选复制的一段
 * 文本（Excel 复制出来是 Tab 分隔）。所以这里同时认三种分隔符，并且支持带引号的
 * 单元格（引号里可以有逗号和换行）。没引第三方库是有意的：为这点格式处理，
 * 不值得再往产物里塞一个依赖。
 */

export type Delimited = ',' | '\t' | ';';

/** 去掉 Excel 写文件时常常留下的 BOM。 */
export function stripBOM(text: string): string {
  return text.charCodeAt(0) === 0xfeff ? text.slice(1) : text;
}

/**
 * 猜分隔符：在第一行里数谁出现得多。
 * 逗号看着更多是「身份证号里有逗号」这类误伤，所以 Tab 只要出现就优先。
 */
export function detectDelimiter(text: string): Delimited {
  const head = stripBOM(text).split(/\r?\n/).find((line) => line.trim() !== '') ?? '';
  // Excel 复制出来一定是 Tab，出现了就认它，别再用计数去比。
  if (head.includes('\t')) return '\t';
  const commas = (head.match(/,/g) ?? []).length;
  return commas > 0 || !head.includes(';') ? ',' : ';';
}

/**
 * 把一段文本解析成单元格二维数组。空行会被丢掉。
 */
export function parseDelimited(input: string, delimiter?: Delimited): string[][] {
  const delimiterToUse = delimiter ?? detectDelimiter(input);
  const text = stripBOM(input).replace(/\r\n/g, '\n').replace(/\r/g, '\n');
  const rows: string[][] = [];
  let row: string[] = [];
  let cell = '';
  let quoted = false;

  const endCell = () => {
    row.push(cell.trim());
    cell = '';
  };
  const endRow = () => {
    endCell();
    if (row.some((value) => value !== '')) rows.push(row);
    row = [];
  };

  for (let i = 0; i < text.length; i++) {
    const ch = text[i];
    if (quoted) {
      if (ch === '"') {
        if (text[i + 1] === '"') {
          cell += '"';
          i++;
        } else {
          quoted = false;
        }
      } else {
        cell += ch;
      }
      continue;
    }
    if (ch === '"' && cell === '') {
      quoted = true;
      continue;
    }
    if (ch === delimiterToUse) {
      endCell();
      continue;
    }
    if (ch === '\n') {
      endRow();
      continue;
    }
    cell += ch;
  }
  if (cell !== '' || row.length > 0) endRow();
  return rows;
}
