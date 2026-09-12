import { useCallback, useEffect, useRef, useState } from 'react';
import { getMyTodo } from '@/api/task';
import { getMyInterviews } from '@/api/interview';
import { getMyApplications } from '@/api/member';
import { getAnnouncementList, type Announcement } from '@/api/announcement';
import { getStatsOverview, type OverviewResponse } from '@/api/stats';
import type { Interview, MemberApplication, Task } from '@/types/api';
import { listWorkflowTasks, type WorkflowTask } from '@/api/workflowRuntime';

interface WorkspaceState {
  approvals: WorkflowTask[]; approvalTotal: number | null;
  tasks: Task[]; taskTotal: number | null; interviews: Interview[]; applications: MemberApplication[];
  announcements: Announcement[];
  overview: OverviewResponse | null; loading: boolean; failed: string[];
}
const initial: WorkspaceState = { approvals: [], approvalTotal: null, tasks: [], taskTotal: null, interviews: [], applications: [], announcements: [], overview: null, loading: true, failed: [] };

export function useWorkspace(canReadStats: boolean, includeAnnouncements = true, flagsReady = true) {
  const [state, setState] = useState<WorkspaceState>(initial);
  const sequence = useRef(0);
  const reload = useCallback(async () => {
    if (!flagsReady) {
      return;
    }
    const current = ++sequence.current;
    setState(previous => ({ ...previous, loading: true, failed: [], overview: canReadStats ? previous.overview : null }));
    const [tasks, interviews, applications, overview, approvals, announcements] = await Promise.allSettled([
      getMyTodo({ page: 1, page_size: 5 }), getMyInterviews(), getMyApplications(),
      canReadStats ? getStatsOverview() : Promise.resolve(null),
      listWorkflowTasks('todo'),
      includeAnnouncements ? getAnnouncementList({ page: 1, page_size: 5, status: 1 }) : Promise.resolve({ list: [], total: 0, page: 1, page_size: 5 }),
    ]);
    if (current !== sequence.current) return;
    const failed: string[] = [];
    if (tasks.status === 'rejected') failed.push('tasks');
    if (interviews.status === 'rejected') failed.push('interviews');
    if (applications.status === 'rejected') failed.push('applications');
    if (overview.status === 'rejected') failed.push('overview');
    if (approvals.status === 'rejected') failed.push('approvals');
    setState({
      approvals: approvals.status === 'fulfilled' ? approvals.value.list.slice(0, 3) : [],
      approvalTotal: approvals.status === 'fulfilled' ? approvals.value.total : null,
      tasks: tasks.status === 'fulfilled' ? tasks.value.list : [],
      taskTotal: tasks.status === 'fulfilled' ? tasks.value.total : null,
      interviews: interviews.status === 'fulfilled' ? interviews.value.filter(item => [0, 1, 2].includes(item.status)) : [],
      applications: applications.status === 'fulfilled' ? applications.value : [],
      announcements: announcements.status === 'fulfilled' ? announcements.value.list : [],
      overview: overview.status === 'fulfilled' ? overview.value : null,
      loading: false, failed,
    });
  }, [canReadStats, includeAnnouncements, flagsReady]);
  useEffect(() => { void reload(); return () => { sequence.current += 1; }; }, [reload]);
  return { ...state, reload };
}
