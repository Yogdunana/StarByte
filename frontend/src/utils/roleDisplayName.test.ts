import { expect, it } from 'vitest';
import i18n from '@/i18n';
import { roleDisplayName } from './roleDisplayName';

it('translates built-in office names but preserves custom role names', async () => {
  const original = i18n.language;
  try {
    await i18n.changeLanguage('ru-RU');
    expect(roleDisplayName({ code: 'president', name: '社长' })).not.toMatch(/[\u3400-\u9fff]/);
    expect(roleDisplayName({ code: 'research_group', name: '研究组' })).toBe('研究组');
    await i18n.changeLanguage('en-US');
    expect(roleDisplayName({ code: 'center_director', name: '中心主任' })).toBe('Center director');
  } finally {
    await i18n.changeLanguage(original);
  }
});
