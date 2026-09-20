import { expect, it } from 'vitest';
import i18n from '@/i18n';
import { roleDisplayName } from './roleDisplayName';

it('translates built-in office names but preserves custom role names', async () => {
  const original = i18n.language;
  try {
    await i18n.changeLanguage('zh-CN');
    expect(roleDisplayName({ code: 'super_admin', name: '超级管理员' })).toBe('系统管理员');
    expect(roleDisplayName({ code: 'president', name: '社长' })).toBe('会长');
    expect(roleDisplayName({ code: 'officer', name: '干事' })).toBe('正式干事');
    expect(roleDisplayName({ code: 'probationary', name: '候补成员' })).toBe('预备干事');
    expect(roleDisplayName({ code: 'vice_center_director', name: '副主任' })).toBe('副中心主任');
    expect(roleDisplayName({ code: 'honorary', name: '荣誉' })).toBe('荣誉会员');
    await i18n.changeLanguage('ru-RU');
    expect(roleDisplayName({ code: 'president', name: '社长' })).not.toMatch(/[\u3400-\u9fff]/);
    expect(roleDisplayName({ code: 'research_group', name: '研究组' })).toBe('研究组');
    await i18n.changeLanguage('en-US');
    expect(roleDisplayName({ code: 'center_director', name: '中心主任' })).toBe('Center director');
    expect(roleDisplayName({ code: 'super_admin', name: '超级管理员' })).toBe('System administrator');
    expect(roleDisplayName({ code: 'officer', name: '干事' })).toBe('Full officer');
  } finally {
    await i18n.changeLanguage(original);
  }
});
