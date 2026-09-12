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
it('keeps missing identity editable', async () => {
  vi.mocked(getCurrentUser).mockResolvedValue({ real_name: '', student_no: '' } as Awaited<
    ReturnType<typeof getCurrentUser>
  >);
  render(<ApplicationForm />);
  await waitFor(() => expect(getCurrentUser).toHaveBeenCalled());
  expect((screen.getByPlaceholderText('真实姓名') as HTMLInputElement).readOnly).toBe(false);
});
