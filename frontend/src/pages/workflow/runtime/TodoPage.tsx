import { useCallback, useEffect, useRef, useState } from 'react';
import { Alert, Button, Empty, Grid, Pagination, Segmented, Skeleton, Table, Tag } from 'antd';
import { ReloadOutlined } from '@ant-design/icons';
import { useSearchParams } from 'react-router-dom';
import PageIntro from '@/components/PageIntro/PageIntro';
import { listWorkflowTasks, type WorkflowTask } from '@/api/workflowRuntime';
import { dateLabel, taskLabels } from './meta';
import TaskDrawer from './TaskDrawer';
import styles from './Runtime.module.css';
export default function TodoPage() {
  const [kind, setKind] = useState<'todo' | 'done'>('todo');
  const [page, setPage] = useState(1);
  const [items, setItems] = useState<WorkflowTask[]>([]);
  const [total, setTotal] = useState<number | null>(null);
  const [loading, setLoading] = useState(true);
  const [failed, setFailed] = useState(false);
  const [params, setParams] = useSearchParams();
  const selected = params.get('task_id');
  const sequence = useRef(0);
  const invalidate = useCallback(() => { sequence.current += 1; }, []);
  const screens = Grid.useBreakpoint();
  const load = useCallback(async () => {
    const seq = ++sequence.current;
    setLoading(true); setFailed(false);
    try { const result = await listWorkflowTasks(kind, page); if (sequence.current !== seq) return; setItems(result.list); setTotal(result.total); }
    catch { if (sequence.current === seq) { setFailed(true); setItems([]); setTotal(null); } }
    finally { if (sequence.current === seq) setLoading(false); }
  }, [kind, page]);
  useEffect(() => { void load(); return invalidate; }, [load, invalidate]);
  const open = (id: string) => setParams({ task_id: id });
  return <>
    <PageIntro eyebrow="WORKFLOW / INBOX" title="我的审批" description="需要你确认的事项，在这里逐一处理。每次决定都会留下记录。" actions={<Button icon={<ReloadOutlined />} onClick={() => void load()} loading={loading}>刷新</Button>} />
    <section className={styles.surface}>
      <div className={styles.toolbar}><Segmented value={kind} options={[{ value: 'todo', label: '待我处理' }, { value: 'done', label: '已处理记录' }]} onChange={value => { setKind(value as 'todo' | 'done'); setPage(1); }} /><p>{failed ? '清单暂不可用' : total === null ? '正在读取审批清单' : `共 ${total} 项${kind === 'todo' ? '待办' : '记录'}`}</p></div>
      {failed ? <Alert type="warning" showIcon message="审批清单加载失败" action={<Button onClick={() => void load()}>重试</Button>} /> : loading ? <div className={styles.cards}><Skeleton active paragraph={{ rows: 6 }} /></div> : screens.md ? <Table rowKey="id" dataSource={items} pagination={false} columns={[
        { title: '当前环节', dataIndex: 'node_name', render: (value: string, row: WorkflowTask) => <Button type="link" onClick={() => open(row.id)}>{value || '审批待办'}</Button> },
        { title: '所属流程', dataIndex: 'definition_name' },
        { title: '状态', render: (_, row: WorkflowTask) => <Tag color={row.status === 0 ? 'gold' : 'default'}>{row.instance_status === 3 && row.status === 0 ? '流程已挂起' : taskLabels[row.status]}</Tag> },
        { title: kind === 'todo' ? '到达时间' : '处理时间', render: (_, row: WorkflowTask) => dateLabel(kind === 'todo' ? row.created_at : row.completed_at) },
        { title: '截止时间', dataIndex: 'due_date', render: dateLabel },
        { title: '', render: (_, row: WorkflowTask) => <Button size="small" onClick={() => open(row.id)}>{kind === 'todo' ? '查看并处理' : '查看记录'}</Button> },
      ]} /> : <div className={styles.cards}>{items.length ? items.map(item => <article className={styles.card} key={item.id}>
        <div className={styles.cardHead}><h3>{item.node_name || '审批待办'}</h3><Tag>{taskLabels[item.status]}</Tag></div><p>{item.definition_name || '审批流程'}</p>
        <p>{kind === 'todo' ? '到达' : '处理'}于 {dateLabel(kind === 'todo' ? item.created_at : item.completed_at)}</p>
        {item.due_date && <p>截止 {dateLabel(item.due_date)}</p>}<Button onClick={() => open(item.id)}>{kind === 'todo' ? '查看并处理' : '查看记录'}</Button>
      </article>) : <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description={kind === 'todo' ? '暂无待办，可以继续其他工作' : '还没有处理记录'} />}</div>}
      {!!total && <div className={styles.footer}><Pagination simple={!screens.md} current={page} pageSize={12} total={total} showSizeChanger={false} onChange={setPage} /></div>}
    </section>
    <TaskDrawer id={selected} onClose={() => setParams({})} onChanged={() => void load()} />
  </>;
}
