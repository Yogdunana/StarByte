import { render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { getMembershipPortal } from '@/api/feature';
import { useFeature } from '@/hooks/useFeature';
import PortalPage from './PortalPage';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock('@/hooks/useFeature', () => ({
  useFeature: vi.fn(),
}));

vi.mock('@/api/feature', () => ({
  getMembershipPortal: vi.fn(),
}));

beforeEach(() => {
  vi.mocked(getMembershipPortal).mockReset();
  vi.mocked(getMembershipPortal).mockResolvedValue({
    applications: [{ id: 'a1', real_name: '张三', status: 1, submitted_at: '2026-09-12' }],
    hint: '',
  });
});

describe('PortalPage', () => {
  it('does not fetch while the membership.portal flag is loading', async () => {
    vi.mocked(useFeature).mockReturnValue({ enabled: false, loading: true, reason: '' });
    render(<MemoryRouter><PortalPage /></MemoryRouter>);
    await waitFor(() => expect(getMembershipPortal).not.toHaveBeenCalled());
  });

  it('does not fetch when membership.portal is closed', async () => {
    vi.mocked(useFeature).mockReturnValue({ enabled: false, loading: false, reason: 'disabled' });
    render(<MemoryRouter><PortalPage /></MemoryRouter>);
    await waitFor(() => expect(screen.getByText('feature.gated')).toBeInTheDocument());
    expect(getMembershipPortal).not.toHaveBeenCalled();
  });

  it('fetches portal data only after the flag is on', async () => {
    vi.mocked(useFeature).mockReturnValue({ enabled: true, loading: false, reason: 'boolean' });
    render(<MemoryRouter><PortalPage /></MemoryRouter>);
    await waitFor(() => expect(getMembershipPortal).toHaveBeenCalledTimes(1));
    expect(await screen.findByText('张三')).toBeInTheDocument();
  });
});
