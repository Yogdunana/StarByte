import { render, screen } from '@testing-library/react';
import { createElement, type ReactNode } from 'react';
import { MemoryRouter } from 'react-router-dom';
import { describe, expect, it, vi } from 'vitest';
import Dashboard from './Dashboard';

vi.mock('react-redux', () => ({
  useSelector: (selector: (state: unknown) => unknown) =>
    selector({
      user: { currentUser: { real_name: 'Admin' } },
      notification: { unreadCount: 3 },
    }),
}));

vi.mock('@/hooks/usePermission', () => ({
  usePermission: () => true,
}));

vi.mock('@/hooks/useFeature', () => ({
  useFeature: () => ({ enabled: false, loading: false, reason: '', variant: '' }),
}));

vi.mock('./useWorkspace', () => ({
  useWorkspace: () => ({
    tasks: [],
    taskTotal: 0,
    approvals: [],
    approvalTotal: 0,
    interviews: [],
    applications: [],
    announcements: [],
    overview: null,
    loading: false,
    failed: [],
    reload: vi.fn(),
  }),
}));

vi.mock('react-i18next', async (importOriginal) => {
  const actual = await importOriginal<typeof import('react-i18next')>();
  return {
    ...actual,
    useTranslation: () => ({
      t: (key: string) => key,
      i18n: { language: 'zh-CN' },
    }),
  };
});

vi.mock('motion/react', () => {
  const passthrough = (tag: string) =>
    ({ children, className }: { children?: ReactNode; className?: string }) =>
      createElement(tag, { className }, children);
  return {
    motion: new Proxy(
      {},
      {
        get: (_target, tag: string) => passthrough(typeof tag === 'string' ? tag : 'div'),
      },
    ),
  };
});

describe('Dashboard metrics', () => {
  it('renders four aligned metric cards in a single list', () => {
    render(
      <MemoryRouter>
        <Dashboard />
      </MemoryRouter>,
    );
    const list = screen.getByRole('list', { name: 'dashboard.metricsAria' });
    expect(list.querySelectorAll(':scope > li')).toHaveLength(4);
    expect(list.querySelectorAll('a')).toHaveLength(4);
    expect(screen.getByText('dashboard.metricApprovals')).toBeTruthy();
    expect(screen.getByText('dashboard.metricTasks')).toBeTruthy();
    expect(screen.getByText('dashboard.metricInterviews')).toBeTruthy();
    expect(screen.getByText('dashboard.metricUnread')).toBeTruthy();
    expect(screen.getByText('01')).toBeTruthy();
    expect(screen.getByText('02')).toBeTruthy();
    expect(screen.getByText('03')).toBeTruthy();
    expect(screen.getByText('04')).toBeTruthy();
  });
});
