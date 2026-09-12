import { act, renderHook, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { getMyTodo } from '@/api/task';
import { getMyInterviews } from '@/api/interview';
import { getMyApplications } from '@/api/member';
import { getStatsOverview } from '@/api/stats';
import { useWorkspace } from './useWorkspace';
import { listWorkflowTasks } from '@/api/workflowRuntime';
import { getAnnouncementList } from '@/api/announcement';

vi.mock('@/api/task', () => ({ getMyTodo: vi.fn() }));
vi.mock('@/api/interview', () => ({ getMyInterviews: vi.fn() }));
vi.mock('@/api/member', () => ({ getMyApplications: vi.fn() }));
vi.mock('@/api/stats', () => ({ getStatsOverview: vi.fn() }));
vi.mock('@/api/workflowRuntime', () => ({ listWorkflowTasks: vi.fn() }));
vi.mock('@/api/announcement', () => ({ getAnnouncementList: vi.fn() }));

beforeEach(() => {
  vi.resetAllMocks();
  vi.mocked(getMyTodo).mockResolvedValue({ list: [], total: 0, page: 1, page_size: 5 });
  vi.mocked(getMyInterviews).mockResolvedValue([]);
  vi.mocked(listWorkflowTasks).mockResolvedValue({ list: [], total: 0, page: 1, page_size: 20 });
  vi.mocked(getMyApplications).mockResolvedValue([]);
  vi.mocked(getAnnouncementList).mockResolvedValue({ list: [], total: 0, page: 1, page_size: 5 });
});

describe('workspace data', () => {
  it('does not request association statistics without permission', async () => {
    const { result } = renderHook(() => useWorkspace(false));
    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(getStatsOverview).not.toHaveBeenCalled();
    expect(result.current.overview).toBeNull();
  });
  it('reports task failures as unavailable instead of showing a false zero', async () => {
    vi.mocked(getMyTodo).mockRejectedValueOnce(new Error('unavailable'));
    const { result } = renderHook(() => useWorkspace(false));
    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.taskTotal).toBeNull();
    expect(result.current.failed).toEqual(['tasks']);
    await act(async () => { await result.current.reload(); });
    expect(result.current.taskTotal).toBe(0);
    expect(result.current.failed).toEqual([]);
  });
  it('reports approval failures as unavailable and recovers on retry', async () => {
    vi.mocked(listWorkflowTasks).mockRejectedValueOnce(new Error('unavailable'));
    const { result } = renderHook(() => useWorkspace(false));
    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.approvalTotal).toBeNull();
    expect(result.current.failed).toEqual(['approvals']);
    await act(async () => { await result.current.reload(); });
    expect(result.current.approvalTotal).toBe(0);
    expect(result.current.failed).toEqual([]);
  });
  it('loads homepage announcement feed without blocking workspace', async () => {
    vi.mocked(getAnnouncementList).mockResolvedValueOnce({
      list: [{ id: 'a1', title: '周报', pinned: true } as never],
      total: 1, page: 1, page_size: 5,
    });
    const { result } = renderHook(() => useWorkspace(false));
    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.announcements).toHaveLength(1);
    expect(getAnnouncementList).toHaveBeenCalled();
  });
  it('does not replace a refreshed result with an older late response', async () => {
    let finish!: (value: Awaited<ReturnType<typeof getMyTodo>>) => void;
    vi.mocked(getMyTodo).mockImplementationOnce(() => new Promise(resolve => { finish = resolve; }));
    const { result } = renderHook(() => useWorkspace(false));
    await act(async () => { await result.current.reload(); });
    await act(async () => { finish({ list: [], total: 99, page: 1, page_size: 5 }); });
    expect(result.current.taskTotal).toBe(0);
  });
});
