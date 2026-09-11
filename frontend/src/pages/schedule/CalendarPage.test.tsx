import { render, screen, waitFor } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import CalendarPage from './CalendarPage';

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
    list: [{ id: 'cal-1', name: '我的日历', calendar_type: 1, source: 'personal', color: '#2563eb', can_edit: true }],
    total: 1, page: 1, page_size: 50,
  }),
  rangeEvents: vi.fn().mockResolvedValue([
    {
      id: 'ev-1', calendar_id: 'cal-1', calendar_name: '我的日历', calendar_color: '#2563eb',
      title: '联调', start_at: '2026-09-11T10:00:00Z', end_at: '2026-09-11T11:00:00Z',
      location: 'A101', recurrence: 'none', can_edit: true, color: '#2563eb',
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
    render(<CalendarPage />);
    await waitFor(() => expect(screen.getByText('schedule.title')).toBeInTheDocument());
    expect(screen.getByText('schedule.desc')).toBeInTheDocument();
    expect(screen.getByText('schedule.newCalendar')).toBeInTheDocument();
    expect(screen.getByText('schedule.newEvent')).toBeInTheDocument();
    expect(screen.getByTestId('schedule-import')).toBeInTheDocument();
    expect(screen.getByTestId('schedule-layers')).toBeInTheDocument();
    expect(screen.getByText('schedule.layers')).toBeInTheDocument();
    expect(screen.getByTitle('schedule.view.month')).toBeInTheDocument();
    expect(document.querySelector('.ant-picker-calendar')).toBeTruthy();
  });
});
