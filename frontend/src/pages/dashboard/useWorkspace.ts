import { useCallback, useEffect, useRef, useState } from 'react';
import { getMyTodo } from '@/api/task';
import { getMyInterviews } from '@/api/interview';
import { getMyApplications } from '@/api/member';
import { getStatsOverview, type OverviewResponse } from '@/api/stats';
import type { Interview, MemberApplication, Task } from '@/types/api';
import { listWorkflowTasks, type WorkflowTask } from '@/api/workflowRuntime';

interface WorkspaceState {
  approvals: WorkflowTask[]; approvalTotal: number | null;
  tasks: Task[]; taskTotal: number | null; interviews: Interview[]; applications: MemberApplication[];
  overview: OverviewResponse | null; loading: boolean; failed: string[];
}
const initial: WorkspaceState = { approvals: [], approvalTotal: null, tasks: [], taskTotal: null, interviews: [], applications: [], overview: null, loading: true, failed: [] };

export function useWorkspace(canReadStats: boolean) {
  const [state, setState] = useState<WorkspaceState>(initial);
  const sequence = useRef(0);
  const reload = useCallback(async () => {
    const current = ++sequence.current;
    setState(previous => ({ ...previous, loading: true, failed: [], overview: canReadStats ? previous.overview : null }));
    const [tasks, interviews, applications, overview, approvals] = await Promise.allSettled([
      getMyTodo({ page: 1, page_size: 5 }), getMyInterviews(), getMyApplications(),
      canReadStats ? getStatsOverview() : Promise.resolve(null),
      listWorkflowTasks('todo'),
    ]);
    if (current !== sequence.current) return;
    const failed: string[] = [];
    if (tasks.status === 'rejected') failed.push('我的任务');
    if (interviews.status === 'rejected') failed.push('面试安排');
    if (applications.status === 'rejected') failed.push('申请进度');
    if (overview.status === 'rejected') failed.push('协会概览');
    if (approvals.status === 'rejected') failed.push('审批待办');
    setState({
      approvals: approvals.status === 'fulfilled' ? approvals.value.list.slice(0, 3) : [],
      approvalTotal: approvals.status === 'fulfilled' ? approvals.value.total : null,
      tasks: tasks.status === 'fulfilled' ? tasks.value.list : [],
      taskTotal: tasks.status === 'fulfilled' ? tasks.value.total : null,
      interviews: interviews.status === 'fulfilled' ? interviews.value.filter(item => [0, 1, 2].includes(item.status)) : [],
      applications: applications.status === 'fulfilled' ? applications.value : [],
      overview: overview.status === 'fulfilled' ? overview.value : null,
      loading: false, failed,
    });
  }, [canReadStats]);
  useEffect(() => { void reload(); return () => { sequence.current += 1; }; }, [reload]);
  return { ...state, reload };
}
