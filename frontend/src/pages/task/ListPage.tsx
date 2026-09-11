import { useCallback, useEffect, useRef, useState } from 'react';
import { Alert, Button, Card, Input, Modal, Select, message } from 'antd';
import { PlusOutlined, ReloadOutlined } from '@ant-design/icons';
import PageIntro from '@/components/PageIntro/PageIntro';
import { usePermission } from '@/hooks/usePermission';
import { createTask, deleteTask, getTaskList, getTaskStats, updateTask } from '@/api/task';
import type { Task, TaskPriority, TaskStats, TaskStatus } from '@/types/api';
import { TaskPriorityMap, TaskStatusMap } from './meta';
import FormModal from './FormModal';
import DetailDrawer from './DetailDrawer';
import TaskCollection from './TaskCollection';
import styles from './TaskWorkspace.module.css';
export default function ListPage() {
  const canCreate = usePermission('task:create');
  const [modal, holder] = Modal.useModal();
  const [list, setList] = useState<Task[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [status, setStatus] = useState<TaskStatus>();
  const [priority, setPriority] = useState<TaskPriority>();
  const [keyword, setKeyword] = useState('');
  const [sortBy, setSortBy] = useState('created_at');
  const [loading, setLoading] = useState(false);
  const [failed, setFailed] = useState(false);
  const [open, setOpen] = useState(false);
  const [editing, setEditing] = useState<Task | null>(null);
  const [detailId, setDetailId] = useState<string | null>(null);
  const [stats, setStats] = useState<TaskStats | null>(null);
  const seq = useRef(0);
  const invalidate = useCallback(() => { seq.current++; }, []);
  const load = useCallback(async () => {
    const current = ++seq.current; setLoading(true);
    try {
      const [res, summary] = await Promise.all([getTaskList({ page, page_size: 10, status, priority, keyword, sort_by: sortBy, sort_order: sortBy === 'due_date' ? 'asc' : 'desc' }), getTaskStats()]);
      if (seq.current !== current) return;
      if (!res.list.length && res.total > 0 && page > 1) { setPage(page - 1); return; }
      setList(res.list || []); setTotal(res.total); setStats(summary); setFailed(false);
    } catch { if (seq.current === current) setFailed(true); } finally { if (seq.current === current) setLoading(false); }
  }, [page, status, priority, keyword, sortBy]);
  useEffect(() => { void load(); return invalidate; }, [load, invalidate]);
  const remove = (task: Task) => modal.confirm({ title: '删除这项任务？', content: '任务将移出日常列表，已有评论和流转记录会保留。存在子任务时，请先处理子任务。', okText: '删除任务', okButtonProps: { danger: true }, cancelText: '保留任务', onOk: async () => { await deleteTask(task.id); message.success('任务已删除'); await load(); } });
  return <>{holder}<PageIntro eyebrow="WORKSPACE / TASKS" title="任务协作" description="分配负责人，跟进进度，保留每一次交接。" actions={<>{canCreate && <Button type="primary" icon={<PlusOutlined />} onClick={() => { setEditing(null); setOpen(true); }}>新建任务</Button>}<Button icon={<ReloadOutlined />} onClick={() => void load()} loading={loading}>刷新</Button></>} />
    {stats && <div className={styles.metrics}><div><span>可见任务</span><strong>{stats.total}</strong></div><div><span>进行中</span><strong>{stats.by_status.doing || 0}</strong></div><div><span>已完成</span><strong>{stats.by_status.done || 0}</strong></div><div><span>需要关注 · 超期</span><strong className={styles.overdue}>{stats.overdue}</strong></div></div>}
    <Card>
      <div className={styles.filters}><Input.Search allowClear placeholder="搜索任务标题或说明" onSearch={v => { setKeyword(v); setPage(1); }} />
        <Select aria-label="任务状态" allowClear placeholder="全部状态" value={status} onChange={v => { setStatus(v); setPage(1); }} options={Object.entries(TaskStatusMap).map(([k, v]) => ({ value: Number(k), label: v.text }))} />
        <Select aria-label="任务优先级" allowClear placeholder="全部优先级" value={priority} onChange={v => { setPriority(v); setPage(1); }} options={Object.entries(TaskPriorityMap).map(([k, v]) => ({ value: Number(k), label: v.text }))} />
        <Select aria-label="排序方式" value={sortBy} onChange={v => { setSortBy(v); setPage(1); }} options={[{ value: 'created_at', label: '最近创建' }, { value: 'due_date', label: '按截止时间' }, { value: 'priority', label: '优先级从高到低' }]} />
      </div>
      {failed && <Alert type="error" showIcon message="任务加载失败，当前内容可能不是最新状态" action={<Button onClick={() => void load()}>重试</Button>} style={{ marginBottom: 16 }} />}
      <TaskCollection rows={list} loading={loading} page={page} total={total} onPage={setPage} onOpen={t => setDetailId(t.id)} actions={t => <>{t.can_update && <Button size="small" type="text" onClick={() => { setEditing(t); setOpen(true); }}>编辑</Button>}{t.can_delete && <Button size="small" type="text" danger onClick={() => remove(t)}>删除</Button>}</>} />
    </Card>
    <FormModal open={open} editing={editing} onCancel={() => setOpen(false)} onSubmit={async values => { if (editing) await updateTask(editing.id, values); else await createTask(values); message.success(editing ? '修改已保存' : '任务已创建'); setOpen(false); await load(); }} />
    <DetailDrawer taskId={detailId} open={!!detailId} onClose={() => setDetailId(null)} onChanged={() => void load()} />
  </>;
}
