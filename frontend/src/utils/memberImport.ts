import aliases from './memberImportAliases.json';
import { parseDelimited } from './delimited';

/**
 * 名单列怎么对应到后端字段。
 *
 * 测评系统的导出表头长什么样没法控制，所以每个字段都给了一串别名（见同目录的
 * memberImportAliases.json —— 放在 json 里是因为这些中文是全球唯一要从源头
 * 表头里匹配的字面量，不能随界面语言变，也不该被当成 UI 文案）。
 * 认识不了的列直接丢掉，不在这里报错：缺什么由后端统一判定，前端不做第二套校验。
 * 第一行如果看着就是数据（一个已知列名都没有），按固定列顺序读。
 */

export interface MemberImportDraftRow {
  row: number;
  student_no: string;
  real_name: string;
  gender?: number;
  grade: string;
  major: string;
  contact_phone: string;
  contact_email: string;
  department: string;
  role: string;
}

type ImportField = Exclude<keyof MemberImportDraftRow, 'row'>;

const POSITIONAL_ORDER: ImportField[] = [
  'student_no',
  'real_name',
  'gender',
  'grade',
  'major',
  'contact_phone',
  'contact_email',
  'department',
  'role',
];

const FIELD_ALIASES: Record<ImportField, string[]> = aliases.fields;
const GENDER_VALUES: Record<string, number> = aliases.genders;

/** 模板文件的表头，用来给管理员一个「照着填」的样板。 */
export const MEMBER_IMPORT_TEMPLATE_HEADER: string[] = aliases.templateHeader;

function normalizeHeader(value: string): string {
  return value.trim().toLowerCase().replace(/[\s_-]/g, '');
}

function resolveGender(raw: string): number | undefined {
  const value = raw.trim();
  if (value === '') return undefined;
  const mapped = GENDER_VALUES[value.toLowerCase()];
  if (mapped !== undefined) return mapped;
  const numeric = Number(value);
  return Number.isFinite(numeric) && [0, 1, 2].includes(numeric) ? numeric : undefined;
}

/**
 * 把粘贴或上传的文本变成一批待录入的行。
 * 返回空数组表示这份东西一列都对不上，调用方据此提示。
 */
export function parseMemberImportDraft(content: string): MemberImportDraftRow[] {
  const table = parseDelimited(content);
  if (table.length === 0) return [];

  const columns = new Map<number, ImportField>();
  const header = table[0].map(normalizeHeader);
  header.forEach((name, index) => {
    for (const field of Object.keys(FIELD_ALIASES) as ImportField[]) {
      if (FIELD_ALIASES[field].some((alias) => normalizeHeader(alias) === name)) {
        columns.set(index, field);
        break;
      }
    }
  });

  const hasHeader = columns.size > 0;
  const body = hasHeader ? table.slice(1) : table;
  if (!hasHeader) {
    POSITIONAL_ORDER.forEach((field, index) => columns.set(index, field));
  }

  const rows: MemberImportDraftRow[] = [];
  body.forEach((cells, offset) => {
    const draft: MemberImportDraftRow = {
      row: offset + (hasHeader ? 2 : 1),
      student_no: '',
      real_name: '',
      grade: '',
      major: '',
      contact_phone: '',
      contact_email: '',
      department: '',
      role: '',
    };
    columns.forEach((field, index) => {
      const raw = cells[index] ?? '';
      if (field === 'gender') {
        const gender = resolveGender(raw);
        if (gender !== undefined) draft.gender = gender;
        return;
      }
      draft[field] = raw.trim();
    });
    rows.push(draft);
  });
  return rows;
}
