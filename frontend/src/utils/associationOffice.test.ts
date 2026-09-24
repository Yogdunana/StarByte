import { describe, expect, it } from 'vitest';
import { associationOffice, isScopedOffice } from './associationOffice';

// 这份映射是后端 associationOffice（internal/rbac/handler/role_membership.go）的镜像。
// 两边不同步就会出现「后端放行、前端把入口藏了」，所以这里把整套取值钉死。
describe('associationOffice', () => {
  it('classifies unscoped charter offices', () => {
    for (const code of ['president', 'advisor', 'honorary', 'captain', 'teammate']) {
      expect(associationOffice(code)).toBe('unscoped');
    }
  });

  it('classifies center offices', () => {
    for (const code of ['vice_president', 'center_director', 'vice_center_director']) {
      expect(associationOffice(code)).toBe('center');
    }
  });

  it('classifies ministers as department offices', () => {
    expect(associationOffice('minister')).toBe('department');
  });

  it('treats everything else as a plain role', () => {
    for (const code of ['member', 'officer', 'super_admin', 'custom_role', undefined]) {
      expect(associationOffice(code)).toBe('');
    }
  });

  it('requires a scope only for center and department offices', () => {
    expect(isScopedOffice('minister')).toBe(true);
    expect(isScopedOffice('center_director')).toBe(true);
    // 副中心主任漏进 leadership 列表时，前端会静默隐藏任职范围，这里堵回去。
    expect(isScopedOffice('vice_center_director')).toBe(true);
    expect(isScopedOffice('president')).toBe(false);
    expect(isScopedOffice('advisor')).toBe(false);
    expect(isScopedOffice('honorary')).toBe(false);
  });
});
