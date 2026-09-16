import { act, fireEvent, render, screen, waitFor, cleanup } from '@testing-library/react';
import { afterEach, beforeEach, expect, it, vi } from 'vitest';
import { Modal } from 'antd';
import ApplicationForm from './ApplicationForm';
import { getMemberDepartments } from '@/api/member';
import { getCurrentUser } from '@/api/auth';
vi.mock('@/api/auth', () => ({ getCurrentUser: vi.fn() }));
vi.mock('@/api/member', () => ({
  getMemberDepartments: vi.fn().mockResolvedValue([]),
  submitApplication: vi.fn(),
}));

afterEach(() => {
  cleanup();
  Modal.destroyAll();
  vi.restoreAllMocks();
});
beforeEach(() => {
  vi.mocked(getMemberDepartments).mockResolvedValue([]);
  vi.mocked(getCurrentUser).mockResolvedValue({
    real_name: '张三',
    student_no: '2026001',
    phone: '+86 138-0013-8000',
    email: 'zhang@smbu.edu.cn',
  } as Awaited<ReturnType<typeof getCurrentUser>>);
});
it('prefills identity, keeps it readonly on cancel, and unlocks only the confirmed field', async () => {
  const confirm = vi.spyOn(Modal, 'confirm').mockReturnValue({ destroy: vi.fn(), update: vi.fn() });
  render(<ApplicationForm />);
  const name = (await screen.findByDisplayValue('张三')) as HTMLInputElement;
  const student = screen.getByDisplayValue('2026001') as HTMLInputElement;
  expect(name.readOnly).toBe(true);
  expect(student.readOnly).toBe(true);
  fireEvent.click(name);
  expect(confirm.mock.calls[0][0].title).toBe('是否确定修改');
  expect(name.readOnly).toBe(true);
  fireEvent.click(name);
  act(() => {
    confirm.mock.calls[1][0].onOk?.();
  });
  await waitFor(() => expect(name.readOnly).toBe(false));
  expect(student.readOnly).toBe(true);
  fireEvent.change(name, { target: { value: '李四' } });
  expect(name.value).toBe('李四');
});
it('hides intended department for members and prefills phone plus email', async () => {
  render(<ApplicationForm />);
  await screen.findByDisplayValue('张三');
  expect(screen.queryByText('意向部门')).toBeNull();
  expect(screen.getByText('会员不隶属任何部门，无需选择意向部门')).toBeTruthy();
  expect(screen.getByDisplayValue('13800138000')).toBeTruthy();
  expect(screen.getByDisplayValue('zhang@smbu.edu.cn')).toBeTruthy();
  fireEvent.mouseDown(screen.getAllByRole('combobox')[0]);
  fireEvent.click(await screen.findByText('干事（需面试）'));
  expect(await screen.findByText('意向部门')).toBeTruthy();
  expect(screen.getByText('+86')).toBeTruthy();
  fireEvent.change(screen.getByDisplayValue('13800138000'), { target: { value: '+86 139-0013-9000' } });
  expect(screen.getByDisplayValue('13900139000')).toBeTruthy();
});

it('keeps missing identity editable', async () => {
  vi.mocked(getCurrentUser).mockResolvedValue({ real_name: '', student_no: '' } as Awaited<
    ReturnType<typeof getCurrentUser>
  >);
  render(<ApplicationForm />);
  await waitFor(() => expect(getCurrentUser).toHaveBeenCalled());
  expect((screen.getByPlaceholderText('真实姓名') as HTMLInputElement).readOnly).toBe(false);
});
