import { expect, it } from 'vitest';
import { readFileSync, readdirSync } from 'node:fs';
import { resolve } from 'node:path';
import ts from 'typescript';
import zh from '@/locales/zh-CN.json';
import en from '@/locales/en-US.json';
import ru from '@/locales/ru-RU.json';
import zhText from '@/locales/zh-CN.text.json';
import enText from '@/locales/en-US.text.json';
import ruText from '@/locales/ru-RU.text.json';
import i18n, { persistLang, readLang } from './index';
import { priorityLabelMap } from '@/pages/notification/notificationMeta';

function flatten(data: object, prefix = ''): Record<string, string> {
  return Object.fromEntries(
    Object.entries(data).flatMap(([key, value]) =>
      typeof value === 'string'
        ? [[prefix + key, value]]
        : Object.entries(flatten(value, prefix + key + '.')),
    ),
  );
}
function placeholders(text: string): string[] {
  return (text.match(/{{[^}]+}}/g) || []).sort();
}

it('has complete Chinese, English and Russian keys with matching interpolation variables', () => {
  for (const [source, languages] of [
    [flatten(zh), [flatten(en), flatten(ru)]],
    [zhText, [enText, ruText]],
  ] as const) {
    for (const target of languages) {
      expect(Object.keys(target).sort()).toEqual(Object.keys(source).sort());
      for (const [key, value] of Object.entries(source)) {
        const translated = (target as Record<string, string>)[key];
        expect(translated, key).toBeTruthy();
        expect(placeholders(translated), key).toEqual(placeholders(value));
        if (key !== '简体中文')
          expect(translated.replace(/{{[^}]+}}/g, ''), key).not.toMatch(/[\u3400-\u9fff]/);
      }
    }
  }
});

it('covers static translation calls and rejects newly hardcoded Chinese UI strings', () => {
  const keys = flatten(zh);
  // These are persisted business identifiers, compared against historical server data.
  const identifiers: Record<string, string[]> = {
    'pages/dashboard/DashboardPanels.tsx': ['已完成'],
    'pages/member/application/engineChain.ts': ['干事审批'],
    'pages/notification/actionURL.ts': ['新通知'],
    'pages/notification/localizedText.ts': ['工作流', '事项', '当前环节'],
    'pages/stats/ProviderCharts.tsx': ['任务状态'],
  };
  const root = resolve(process.cwd(), 'src');
  const errors: string[] = [];
  function scan(dir: string): void {
    for (const item of readdirSync(dir, { withFileTypes: true })) {
      const file = resolve(dir, item.name);
      if (item.isDirectory()) {
        if (item.name !== 'locales') scan(file);
        continue;
      }
      if (!/\.tsx?$/.test(file) || /\.(test|spec)\./.test(file)) continue;
      const relative = file.slice(root.length + 1);
      const ast = ts.createSourceFile(
        file,
        readFileSync(file, 'utf8'),
        ts.ScriptTarget.Latest,
        true,
      );
      const translatedCall = (node: ts.Node): boolean => {
        return ts.isCallExpression(node) && /^(t|tx|i18n\.t)$/.test(node.expression.getText(ast));
      };
      const visit = (node: ts.Node): void => {
        if (ts.isCallExpression(node) && translatedCall(node)) {
          const first = node.arguments[0];
          if (first && ts.isStringLiteral(first)) {
            const catalog = node.expression.getText(ast) === 'tx' ? zhText : keys;
            if (!Object.prototype.hasOwnProperty.call(catalog, first.text))
              errors.push(`${relative}: missing ${first.text}`);
          }
        }
        if (
          (ts.isStringLiteral(node) ||
            ts.isJsxText(node) ||
            ts.isNoSubstitutionTemplateLiteral(node) ||
            ts.isTemplateHead(node) ||
            ts.isTemplateMiddle(node) ||
            ts.isTemplateTail(node)) &&
          /[\u3400-\u9fff]/.test(node.text)
        ) {
          let parent: ts.Node | undefined = node.parent;
          let translated = false;
          while (parent) {
            if (translatedCall(parent)) translated = true;
            parent = parent.parent;
          }
          if (!translated && !identifiers[relative]?.includes(node.text))
            errors.push(`${relative}: hardcoded ${node.text}`);
        }
        ts.forEachChild(node, visit);
      };
      visit(ast);
    }
  }
  scan(root);
  expect(errors).toEqual([]);
});

it('persists Russian and updates labels created before a language switch', async () => {
  persistLang('ru-RU');
  expect(readLang()).toBe('ru-RU');
  expect(document.documentElement.lang).toBe('ru-RU');
  await i18n.changeLanguage('ru-RU');
  expect(priorityLabelMap.high).toBe('Высокий');
  await i18n.changeLanguage('en-US');
  expect(priorityLabelMap.high).toBe('High');
  persistLang('zh-CN');
  await i18n.changeLanguage('zh-CN');
});
