import { Empty, Grid, Pagination, Space, Table, Tag } from 'antd';
import type { ReactNode } from 'react';
import type { ColumnsType } from 'antd/es/table';
import dayjs from 'dayjs';
import StatusTag from '@/components/StatusTag/StatusTag';
import type { Task } from '@/types/api';
import { TaskPriorityMap, TaskStatusMap, TaskWorkflowStageMap } from './meta';
import styles from './TaskWorkspace.module.css';

export const taskDate = (value?: string) => value ? dayjs(value).format('MM-DD HH:mm') : '未设截止日期';
export const isOverdue = (task: Task) => !!task.due_date && ![2, 3].includes(task.status) && dayjs(task.due_date).isBefore(dayjs());
interface Props { rows: Task[]; loading: boolean; page: number; total: number; onPage: (page: number) => void; onOpen: (task: Task) => void; actions?: (task: Task) => ReactNode }
export function TaskCard({ task, onOpen, actions }: { task: Task; onOpen: () => void; actions?: ReactNode }) {
  return <article className={styles.card}>
    <div className={styles.cardTop}>{task.workflow_stage ? <Tag color="green">{TaskWorkflowStageMap[task.workflow_stage]}</Tag> : <StatusTag status={task.status} mapping={TaskStatusMap} />}<StatusTag status={task.priority} mapping={TaskPriorityMap} /></div>
    <h3><button className={styles.title} onClick={onOpen}>{task.title}</button></h3>
    <div className={styles.cardMeta}><span>{task.assignee?.name || '待分配负责人'}</span><span className={isOverdue(task) ? styles.late : ''}>{isOverdue(task) ? '已超期 · ' : ''}{taskDate(task.due_date)}</span></div>
    <div className={styles.cardFoot}><span className={styles.subtitle}>{task.department?.name || '个人协作'} · {task.comment_count} 条评论</span>{actions}</div>
  </article>;
}
export default function TaskCollection({ rows, loading, page, total, onPage, onOpen, actions }: Props) {
  const screens = Grid.useBreakpoint();
  const columns: ColumnsType<Task> = [
    { title: '任务', key: 'title', render: (_, t) => <div><button className={styles.title} onClick={() => onOpen(t)}>{t.title}</button><div className={styles.subtitle}>{t.department?.name || '个人协作'} · {t.creator.name} 创建</div></div> },
    { title: '状态', dataIndex: 'status', width: 100, render: (v, t) => t.workflow_stage ? <Tag color="green">{TaskWorkflowStageMap[t.workflow_stage]}</Tag> : <StatusTag status={v} mapping={TaskStatusMap} /> },
    { title: '优先级', dataIndex: 'priority', width: 90, render: v => <StatusTag status={v} mapping={TaskPriorityMap} /> },
    { title: '负责人', key: 'assignee', width: 130, render: (_, t) => t.assignee?.name || <span className={styles.subtitle}>待分配</span> },
    { title: '截止时间', key: 'due', width: 140, render: (_, t) => <span className={`${styles.deadline} ${isOverdue(t) ? styles.late : ''}`}>{taskDate(t.due_date)}{isOverdue(t) && <Tag color="volcano">超期</Tag>}</span> },
  ];
  if (actions) columns.push({ title: '操作', key: 'actions', width: 140, render: (_, t) => <Space>{actions(t)}</Space> });
  return screens.lg ? <Table rowKey="id" expandable={{ childrenColumnName: 'nestedRows' }} loading={loading} columns={columns} dataSource={rows} pagination={{ current: page, total, pageSize: 10, showSizeChanger: false, onChange: onPage }} locale={{ emptyText: '暂无符合条件的任务' }} /> : <>
    {!loading && rows.length === 0 ? <Empty description="暂无符合条件的任务" /> : <div className={styles.cards}>{rows.map(task => <TaskCard key={task.id} task={task} onOpen={() => onOpen(task)} actions={actions?.(task)} />)}</div>}
    <div className={styles.pagination}><Pagination simple current={page} total={total} pageSize={10} onChange={onPage} hideOnSinglePage /></div>
  </>;
}
