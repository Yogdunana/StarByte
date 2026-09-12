import { fireEvent, render, screen, waitFor, cleanup } from '@testing-library/react';
import { afterEach, beforeEach, expect, it, vi } from 'vitest';
import RolePage from './RolePage';
import * as api from '@/api/role';
import { usePermissions } from '@/hooks/usePermission';
vi.mock('@/api/role', () => ({
  getRoleList: vi.fn(),
  getRoleDetail: vi.fn(),
  getPermissionTree: vi.fn(),
  createRole: vi.fn(),
  updateRole: vi.fn(),
  deleteRole: vi.fn(),
  assignRolePermissions: vi.fn(),
}));
vi.mock('@/hooks/usePermission', () => ({ usePermissions: vi.fn() }));
vi.mock('react-redux', () => ({ useDispatch: () => vi.fn() }));
vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string, fallback: string) => fallback || key }),
}));
afterEach(cleanup);
beforeEach(() => {
  vi.clearAllMocks();
  vi.mocked(usePermissions).mockReturnValue([true, true, true, true, true]);
  vi.mocked(api.getRoleList).mockResolvedValue({
    list: [{ id: '1', name: '自定义审核员', code: 'custom', status: 0, is_system: false }],
    total: 1,
  } as Awaited<ReturnType<typeof api.getRoleList>>);
});
it('creates a role through the actual form', async () => {
  render(<RolePage />);
  await screen.findByText('自定义审核员');
  fireEvent.click(screen.getByRole('button', { name: /新\s*增/ }));
  fireEvent.change(screen.getByLabelText('名称'), { target: { value: '编辑员' } });
  fireEvent.change(screen.getByLabelText('编码'), { target: { value: 'editor' } });
  fireEvent.click(screen.getByText('OK'));
  await waitFor(() =>
    expect(api.createRole).toHaveBeenCalledWith(
      expect.objectContaining({ name: '编辑员', code: 'editor' }),
    ),
  );
});
it('hides mutation actions without their permissions', async () => {
  vi.mocked(usePermissions).mockReturnValue([false, false, false, false, false]);
  render(<RolePage />);
  await screen.findByText('自定义审核员');
  expect(screen.queryByText('新增')).toBeNull();
  expect(screen.queryByText('分配权限')).toBeNull();
  expect(screen.queryByText('编辑')).toBeNull();
});
