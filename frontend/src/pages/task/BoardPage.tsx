import { tx, useLocale } from '@/i18n/text';
import { useCallback, useEffect, useState } from 'react';
import { Alert, Button, Card, Empty, Grid, Modal, Pagination, Select, Spin, message } from 'antd';
import PageIntro from '@/components/PageIntro/PageIntro';
import { getTaskList, updateTaskStatus } from '@/api/task';
import type { Task, TaskStatus } from '@/types/api';
import { BOARD_COLUMNS } from './meta';
import DetailDrawer from './DetailDrawer';
import { TaskCard } from './TaskCollection';
import styles from './TaskWorkspace.module.css';
const transitions: Record<number, number[]> = { 0: [1, 3], 1: [2, 4, 3], 4: [1] };
interface LaneProps {
  status: TaskStatus;
  title: string;
  revision: number;
  onOpen: (id: string) => void;
  onDrag: (task: Task) => void;
  onDrop: (status: TaskStatus) => void;
}
function Lane({ status, title, revision, onOpen, onDrag, onDrop }: LaneProps) {
  useLocale();
  const [rows, setRows] = useState<Task[]>([]);
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(false);
  const [failed, setFailed] = useState(false);
  const [retry, setRetry] = useState(0);
  useEffect(() => {
    let active = true;
    setLoading(true);
    void getTaskList({ page, page_size: 10, status, sort_by: 'priority', sort_order: 'desc' })
      .then((res) => {
        if (!active) return;
        if (!res.list.length && res.total > 0 && page > 1) {
          setPage(page - 1);
          return;
        }
        setRows(res.list || []);
        setTotal(res.total);
        setFailed(false);
      })
      .catch(() => {
        if (active) setFailed(true);
      })
      .finally(() => {
        if (active) setLoading(false);
      });
    return () => {
      active = false;
    };
  }, [page, status, revision, retry]);
  return (
    <section
      className={styles.lane}
      onDragOver={(e) => e.preventDefault()}
      onDrop={(e) => {
        e.preventDefault();
        onDrop(status);
      }}
    >
      <h2>
        {title}
        <span>{total}</span>
      </h2>
      {failed && (
        <Alert
          type="error"
          message={tx('加载失败')}
          action={
            <Button size="small" onClick={() => setRetry((v) => v + 1)}>
              {tx('重试')}
            </Button>
          }
        />
      )}
      <Spin spinning={loading}>
        {!rows.length && !loading && (
          <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description={tx('暂无任务')} />
        )}
        {rows.map((task) => (
          <div
            key={task.id}
            draggable={!!task.can_update && !task.workflow_stage}
            onDragStart={() => onDrag(task)}
          >
            <TaskCard task={task} onOpen={() => onOpen(task.id)} />
          </div>
        ))}
      </Spin>
      <Pagination
        simple
        size="small"
        current={page}
        total={total}
        pageSize={10}
        onChange={setPage}
        hideOnSinglePage
      />
    </section>
  );
}
export default function BoardPage() {
  useLocale();
  const screens = Grid.useBreakpoint();
  const [modal, holder] = Modal.useModal();
  const [drag, setDrag] = useState<Task | null>(null);
  const [detailId, setDetailId] = useState<string | null>(null);
  const [revision, setRevision] = useState(0);
  const [mobileStatus, setMobileStatus] = useState<TaskStatus>(0);
  const reload = useCallback(() => setRevision((v) => v + 1), []);
  const drop = (status: TaskStatus) => {
    const task = drag;
    setDrag(null);
    if (!task || task.status === status) return;
    if (task.workflow_stage || !task.can_update || !transitions[task.status]?.includes(status)) {
      message.warning(tx('该任务不能直接进入这个状态，请打开详情查看可用操作'));
      return;
    }
    modal.confirm({
      title: tx('将“{{value0}}”移到{{value1}}？', {
        value0: task.title,
        value1: BOARD_COLUMNS.find((c) => c.status === status)?.title,
      }),
      content: [2, 3].includes(status)
        ? tx('关闭后将保留历史，无法继续编辑或转办。')
        : tx('状态变更会记录到流转历史。'),
      okText: tx('确认变更'),
      cancelText: tx('保留原状态'),
      onOk: async () => {
        await updateTaskStatus(task.id, status);
        reload();
      },
    });
  };
  const columns = screens.lg
    ? BOARD_COLUMNS
    : BOARD_COLUMNS.filter((c) => c.status === mobileStatus);
  return (
    <>
      {holder}
      <PageIntro
        eyebrow="WORKSPACE / BOARD"
        title={tx('任务看板')}
        description={tx('按阶段查看进展。打开任务即可处理，桌面也支持拖动卡片变更状态。')}
        actions={<Button onClick={reload}>{tx('刷新看板')}</Button>}
      />
      <Card>
        {!screens.lg && (
          <Select
            aria-label={tx('看板阶段')}
            value={mobileStatus}
            onChange={setMobileStatus}
            style={{ width: '100%', marginBottom: 20 }}
            options={BOARD_COLUMNS.map((c) => ({ value: c.status, label: c.title }))}
          />
        )}
        <div className={styles.board}>
          {columns.map((col) => (
            <Lane
              key={col.status}
              {...col}
              revision={revision}
              onOpen={setDetailId}
              onDrag={setDrag}
              onDrop={drop}
            />
          ))}
        </div>
      </Card>
      <DetailDrawer
        taskId={detailId}
        open={!!detailId}
        onClose={() => setDetailId(null)}
        onChanged={reload}
      />
    </>
  );
}
