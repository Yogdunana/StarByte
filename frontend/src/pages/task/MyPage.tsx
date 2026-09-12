import { tx, useLocale } from '@/i18n/text';
import { useCallback, useEffect, useRef, useState } from 'react';
import { Alert, Button, Card, Input, Tabs } from 'antd';
import PageIntro from '@/components/PageIntro/PageIntro';
import { getMyCreated, getMyDone, getMyOverdue, getMyTodo } from '@/api/task';
import type { PageResponse, Task } from '@/types/api';
import DetailDrawer from './DetailDrawer';
import TaskCollection from './TaskCollection';
import styles from './TaskWorkspace.module.css';
type TabKey = 'todo' | 'done' | 'created' | 'overdue';
const fetchers: Record<
  TabKey,
  (p: { page: number; page_size: number; keyword?: string }) => Promise<PageResponse<Task>>
> = { todo: getMyTodo, done: getMyDone, created: getMyCreated, overdue: getMyOverdue };
export default function MyPage() {
  useLocale();
  const [tab, setTab] = useState<TabKey>('todo');
  const [list, setList] = useState<Task[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [keyword, setKeyword] = useState('');
  const [loading, setLoading] = useState(false);
  const [failed, setFailed] = useState(false);
  const [detailId, setDetailId] = useState<string | null>(null);
  const seq = useRef(0);
  const invalidate = useCallback(() => {
    seq.current++;
  }, []);
  const load = useCallback(async () => {
    const current = ++seq.current;
    setLoading(true);
    try {
      const res = await fetchers[tab]({ page, page_size: 10, keyword });
      if (current === seq.current) {
        setList(res.list || []);
        setTotal(res.total);
        setFailed(false);
      }
    } catch {
      if (current === seq.current) setFailed(true);
    } finally {
      if (current === seq.current) setLoading(false);
    }
  }, [tab, page, keyword]);
  useEffect(() => {
    void load();
    return invalidate;
  }, [load, invalidate]);
  return (
    <>
      <PageIntro
        eyebrow="MY WORK"
        title={tx('我的任务')}
        description={tx('先处理需要你推进的事，再回看已经完成的工作。')}
        actions={
          <Button loading={loading} onClick={() => void load()}>
            {tx('刷新任务')}
          </Button>
        }
      />
      <Card>
        <div className={styles.filters}>
          <Input.Search
            allowClear
            placeholder={tx('搜索我的任务')}
            onSearch={(v) => {
              setKeyword(v);
              setPage(1);
            }}
          />
        </div>
        <Tabs
          activeKey={tab}
          onChange={(k) => {
            setTab(k as TabKey);
            setPage(1);
          }}
          items={[
            { key: 'todo', label: tx('待我处理') },
            { key: 'overdue', label: tx('已超期') },
            { key: 'created', label: tx('我创建的') },
            { key: 'done', label: tx('已完成') },
          ]}
        />
        {failed && (
          <Alert
            type="error"
            message={tx('任务加载失败')}
            action={<Button onClick={() => void load()}>{tx('重试')}</Button>}
            style={{ marginBottom: 16 }}
          />
        )}
        <TaskCollection
          rows={list}
          loading={loading}
          page={page}
          total={total}
          onPage={setPage}
          onOpen={(t) => setDetailId(t.id)}
        />
      </Card>
      <DetailDrawer
        taskId={detailId}
        open={!!detailId}
        onClose={() => setDetailId(null)}
        onChanged={() => void load()}
      />
    </>
  );
}
