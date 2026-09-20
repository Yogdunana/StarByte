import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';
import { expect, it } from 'vitest';

it('publishes charter titles that match the association role vocabulary', () => {
  const html = readFileSync(resolve(process.cwd(), 'public/charter/smbu-ca-charter.html'), 'utf8');
  expect(html).toContain('会长不是系统管理员');
  expect(html).toContain('面试 → 正式签字 → 预备干事 → 正式干事');
  expect(html).toContain('副中心主任');
  expect(html).toContain('副部长');
  expect(html).toContain('荣誉会员');
  expect(html).toContain('队长 / 队员');
  expect(html).toContain('指导老师');
  expect(html).not.toContain('社长');
  expect(html).not.toContain('候补成员');
  expect(html).not.toContain('指导教师');
});
