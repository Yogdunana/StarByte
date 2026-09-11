import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import { MemoryRouter } from 'react-router-dom';
import CalendarPage from './CalendarPage';

const navigate = vi.fn();
vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual<typeof import('react-router-dom')>('react-router-dom');
  return { ...actual, useNavigate: () => navigate };
});

vi.mock('@/hooks/usePermission', () => ({
  usePermission: () => true,
}));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => key,
  }),
}));

vi.mock('@/api/schedule', () => ({
  listCalendars: vi.fn().mockResolvedValue({
    list: [
      { id: 'cal-1', name: '我的日历', calendar_type: 1, source: 'personal', color: '#2563eb', can_edit: true },
      { id: 'aaaaaaaa-0000-4000-8000-0000000000a1', name: '活动', calendar_type: 1, source: 'activity', color: '#7c3aed', can_edit: false },
      { id: 'bbbbbbbb-0000-4000-8000-0000000000b2', name: '面试', calendar_type: 1, source: 'interview', color: '#db2777', can_edit: false },
    ],
    total: 3, page: 1, page_size: 50,
  }),
  rangeEvents: vi.fn().mockResolvedValue([
    {
      id: 'ev-1', calendar_id: 'cal-1', calendar_name: '我的日历', calendar_color: '#2563eb',
      title: '联调', start_at: '2026-09-11T10:00:00Z', end_at: '2026-09-11T11:00:00Z',
      location: 'A101', recurrence: 'none', can_edit: true, color: '#2563eb', source: 'personal',
    },
    {
      id: 'act-1', calendar_id: 'aaaaaaaa-0000-4000-8000-0000000000a1', calendar_name: '活动', calendar_color: '#7c3aed',
      title: '迎新晚会', start_at: '2026-09-11T14:00:00Z', end_at: '2026-09-11T16:00:00Z',
      location: '礼堂', recurrence: 'none', can_edit: false, color: '#7c3aed', source: 'activity', link: '/activity/act-1',
    },
  ]),
  createCalendar: vi.fn(),
  createEvent: vi.fn(),
  updateEvent: vi.fn(),
  deleteEvent: vi.fn(),
  importTimetable: vi.fn(),
  importICS: vi.fn(),
  googleStatus: vi.fn().mockResolvedValue({ configured: false, connected: false }),
  googleConnect: vi.fn(),
  googleDisconnect: vi.fn(),
  googleSync: vi.fn(),
}));

describe('CalendarPage', () => {
  it('renders schedule chrome and default month view', async () => {
    render(<MemoryRouter><CalendarPage /></MemoryRouter>);
    await waitFor(() => expect(screen.getByText('schedule.title')).toBeInTheDocument());
    expect(screen.getByText('schedule.desc')).toBeInTheDocument();
    expect(screen.getByText('schedule.newCalendar')).toBeInTheDocument();
    expect(screen.getByText('schedule.newEvent')).toBeInTheDocument();
    expect(screen.getByTestId('schedule-import')).toBeInTheDocument();
    expect(screen.getByTestId('schedule-layers')).toBeInTheDocument();
    expect(screen.getByText('schedule.layers')).toBeInTheDocument();
    expect(screen.getByText('schedule.source.activity')).toBeInTheDocument();
    expect(screen.getByText('schedule.source.interview')).toBeInTheDocument();
    expect(screen.getByRole('radio', { name: 'schedule.view.month' })).toBeInTheDocument();
    expect(document.querySelector('.ant-picker-calendar')).toBeTruthy();
  });

  it('deep-links read-only activity events instead of opening the editor', async () => {
    navigate.mockClear();
    const user = userEvent.setup();
    render(<MemoryRouter><CalendarPage /></MemoryRouter>);
    await waitFor(() => expect(screen.getByTestId('schedule-layers')).toBeInTheDocument());
    await user.click(screen.getByText('schedule.view.agenda'));
    await user.click(await screen.findByText('迎新晚会'));
    expect(navigate).toHaveBeenCalledWith('/activity/act-1');
    expect(screen.queryByText('common.edit')).not.toBeInTheDocument();
  });
});
