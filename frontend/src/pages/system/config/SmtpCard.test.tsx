import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { afterEach, expect, it, vi } from 'vitest';
import { getSMTPSettings, updateSMTPSettings } from '@/api/config';
import SmtpCard from './SmtpCard';
vi.mock('@/api/config', () => ({
  getSMTPSettings: vi.fn(),
  updateSMTPSettings: vi.fn(),
  testSMTPSettings: vi.fn(),
}));
vi.mock('@/hooks/usePermission', () => ({ usePermission: () => true }));
vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string, fallback?: string) => fallback || key }),
}));
afterEach(cleanup);
it('sends a replacement password once then clears it and preserves it on subsequent saves', async () => {
  const settings = {
    host: 'smtp.example.test',
    port: 465,
    ssl_mode: 'implicit' as const,
    from: 'a@example.test',
    from_name: 'StarByte',
    username: 'a@example.test',
    password_configured: true,
    password_source: 'web',
  };
  vi.mocked(getSMTPSettings).mockResolvedValue(settings);
  vi.mocked(updateSMTPSettings).mockResolvedValue(settings);
  render(<SmtpCard />);
  const field = await screen.findByPlaceholderText('输入新密码；留空保留当前密码');
  expect(field).toHaveValue('');
  fireEvent.change(field, { target: { value: 'replacement-secret' } });
  fireEvent.click(screen.getByRole('button', { name: 'smtp.save' }));
  await waitFor(() =>
    expect(updateSMTPSettings).toHaveBeenCalledWith(
      expect.objectContaining({ password: 'replacement-secret' }),
    ),
  );
  await waitFor(() => expect(field).toHaveValue(''));
  fireEvent.click(screen.getByRole('button', { name: 'smtp.save' }));
  await waitFor(() =>
    expect(updateSMTPSettings).toHaveBeenLastCalledWith(
      expect.objectContaining({ password: undefined }),
    ),
  );
});
