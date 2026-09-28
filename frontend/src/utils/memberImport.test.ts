import { describe, expect, it } from 'vitest';
import { detectDelimiter, parseDelimited, stripBOM } from './delimited';
import { MEMBER_IMPORT_TEMPLATE_HEADER, parseMemberImportDraft } from './memberImport';

describe('parseDelimited', () => {
  it('splits simple CSV rows and drops empty lines', () => {
    const rows = parseDelimited('a,b,c\n1,2,3\n\n4,5,6');
    expect(rows).toEqual([
      ['a', 'b', 'c'],
      ['1', '2', '3'],
      ['4', '5', '6'],
    ]);
  });

  it('keeps commas and newlines inside quoted cells', () => {
    const rows = parseDelimited('name,note\n张三,"换行\n还有,逗号"');
    expect(rows[1]).toEqual(['张三', '换行\n还有,逗号']);
  });

  it('unescapes doubled quotes', () => {
    // 单元格 他说"你好" 在 CSV 里写成 "他说""你好"""
    expect(parseDelimited('a\n"他说""你好"""')[1]).toEqual(['他说"你好"']);
  });

  it('reads Windows line endings and strips BOM', () => {
    expect(stripBOM('\ufeffabc')).toBe('abc');
    expect(parseDelimited('\ufeffa,b\r\n1,2\r\n')).toEqual([
      ['a', 'b'],
      ['1', '2'],
    ]);
  });

  it('picks the delimiter from the first non-empty line', () => {
    expect(detectDelimiter('a,b\n1,2')).toBe(',');
    expect(detectDelimiter('a\tb\n1\t2')).toBe('\t');
    expect(detectDelimiter('a;b\n1;2')).toBe(';');
    // Excel 复制出来的 Tab 要压过中文姓名里可能出现的逗号
    expect(detectDelimiter('张三,李四\t备注')).toBe('\t');
  });
});

describe('parseMemberImportDraft', () => {
  it('maps Chinese headers and keeps the spreadsheet row number', () => {
    const rows = parseMemberImportDraft(
      ['学号,姓名,性别,部门,角色', '20240001,张三,男,技术部,部长', '20240002,李四,女,技术部,'].join('\n'),
    );
    expect(rows).toHaveLength(2);
    expect(rows[0].row).toBe(2);
    expect(rows[0]).toMatchObject({
      student_no: '20240001',
      real_name: '张三',
      gender: 1,
      department: '技术部',
      role: '部长',
    });
    expect(rows[1].gender).toBe(2);
    expect(rows[1].role).toBe('');
  });

  it('understands English headers too', () => {
    const rows = parseMemberImportDraft('student_no,real_name\n20240003,王五');
    expect(rows[0]).toMatchObject({ student_no: '20240003', real_name: '王五' });
  });

  it('falls back to positional order when there is no header row', () => {
    const rows = parseMemberImportDraft('20240004,赵六,1,2024级,计算机系');
    expect(rows[0]).toMatchObject({
      student_no: '20240004',
      real_name: '赵六',
      gender: 1,
      grade: '2024级',
      major: '计算机系',
    });
  });

  it('ignores unknown columns instead of failing', () => {
    const rows = parseMemberImportDraft('学号,备注,姓名\n20240005,ok,孙七');
    expect(rows[0]).toMatchObject({ student_no: '20240005', real_name: '孙七' });
  });

  it('returns nothing for empty content', () => {
    expect(parseMemberImportDraft('  \n \n')).toEqual([]);
    expect(parseMemberImportDraft('')).toEqual([]);
  });

  it('ships a template header that itself round-trips', () => {
    const rows = parseMemberImportDraft(MEMBER_IMPORT_TEMPLATE_HEADER.join(','));
    expect(rows).toEqual([]);
  });
});
