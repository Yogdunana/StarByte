import { fireEvent, render, screen, waitFor, cleanup } from '@testing-library/react';
import { afterEach, expect, it, vi } from 'vitest';
import { configureStore } from '@reduxjs/toolkit';
import { Provider } from 'react-redux';
import reducer from '@/store/slices/userSlice';
import { getCurrentUser } from '@/api/auth';
import { updateMyProfile } from '@/api/user';
import ProfileMePage from './ProfileMePage';
vi.mock('@/api/auth', () => ({ getCurrentUser: vi.fn() }));
vi.mock('@/api/user', () => ({ updateMyProfile: vi.fn().mockResolvedValue(undefined) }));
vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string, fallback: string) => fallback || key }),
}));
afterEach(cleanup);
it('saves only editable fields then refreshes the current user', async () => {
  const user = {
    id: 'me',
    username: 'stable-login',
    real_name: '原姓名',
    roles: ['member'],
    permissions: [],
    email: 'old@example.test',
    phone: '123',
    gender: 0,
  } as Awaited<ReturnType<typeof getCurrentUser>>;
  vi.mocked(getCurrentUser).mockResolvedValue(user);
  const store = configureStore({ reducer: { user: reducer } });
  render(
    <Provider store={store}>
      <ProfileMePage />
    </Provider>,
  );
  await screen.findByText('原姓名');
  fireEvent.click(screen.getByRole('button', { name: '编辑资料' }));
  fireEvent.change(screen.getByLabelText('姓名'), { target: { value: '新姓名' } });
  vi.mocked(getCurrentUser).mockResolvedValue({ ...user, real_name: '新姓名' });
  fireEvent.click(screen.getByRole('button', { name: 'OK' }));
  await waitFor(() =>
    expect(updateMyProfile).toHaveBeenCalledWith({
      real_name: '新姓名',
      email: 'old@example.test',
      phone: '123',
      gender: 0,
    }),
  );
  await screen.findByText('新姓名');
  expect(store.getState().user.currentUser?.real_name).toBe('新姓名');
});
