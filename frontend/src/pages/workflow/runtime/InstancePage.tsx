import { tx, useLocale } from '@/i18n/text';
import { useCallback, useEffect, useRef, useState } from 'react';
import { Alert, Button, Empty, Grid, Pagination, Select, Skeleton, Table, Tag } from 'antd';
import { ReloadOutlined } from '@ant-design/icons';
import { useSearchParams } from 'react-router-dom';
import PageIntro from '@/components/PageIntro/PageIntro';
import { listWorkflowInstances, type WorkflowInstance } from '@/api/workflowRuntime';
import { dateLabel, instanceLabels } from './meta';
import InstanceDrawer from './InstanceDrawer';
import styles from './Runtime.module.css';
export default function InstancePage() {
  useLocale();
  const [status, setStatus] = useState<number | undefined>();
  const [page, setPage] = useState(1);
  const [items, setItems] = useState<WorkflowInstance[]>([]);
  const [total, setTotal] = useState<number | null>(null);
  const [loading, setLoading] = useState(true);
  const [failed, setFailed] = useState(false);
  const [params, setParams] = useSearchParams();
  const sequence = useRef(0);
  const invalidate = useCallback(() => {
    sequence.current += 1;
  }, []);
  const screens = Grid.useBreakpoint();
  const load = useCallback(async () => {
    const seq = ++sequence.current;
    setLoading(true);
    setFailed(false);
    try {
      const result = await listWorkflowInstances(page, status);
      if (sequence.current !== seq) return;
      setItems(result.list);
      setTotal(result.total);
    } catch {
      if (sequence.current === seq) {
        setItems([]);
        setTotal(null);
        setFailed(true);
      }
    } finally {
      if (sequence.current === seq) setLoading(false);
    }
  }, [page, status]);
  useEffect(() => {
    void load();
    return invalidate;
  }, [load, invalidate]);
  const open = (id: string) => setParams({ instance_id: id });
  return (
    <>
      <PageIntro
        eyebrow="WORKFLOW / PROGRESS"
        title={tx('流程进度')}
        description={tx('查看你发起、参与或负责范围内的流程，了解每一次处理的结果。')}
        actions={
          <Button icon={<ReloadOutlined />} onClick={() => void load()} loading={loading}>
            {tx('刷新')}
          </Button>
        }
      />
      <section className={styles.surface}>
        <div className={styles.toolbar}>
          <Select
            aria-label={tx('筛选流程状态')}
            placeholder={tx('全部状态')}
            allowClear
            value={status}
            onChange={(value) => {
              setStatus(value);
              setPage(1);
            }}
            style={{ width: 160 }}
            options={Object.entries(instanceLabels).map(([value, label]) => ({
              value: Number(value),
              label,
            }))}
          />
          <p>
            {total === null
              ? tx('正在读取流程清单')
              : tx('共 {{value0}} 个流程', { value0: total })}
          </p>
        </div>
        {failed ? (
          <Alert
            type="warning"
            showIcon
            message={tx('流程清单加载失败')}
            action={<Button onClick={() => void load()}>{tx('重试')}</Button>}
          />
        ) : loading ? (
          <div className={styles.cards}>
            <Skeleton active paragraph={{ rows: 6 }} />
          </div>
        ) : screens.md ? (
          <Table
            rowKey="id"
            dataSource={items}
            pagination={false}
            columns={[
              {
                title: tx('流程名称'),
                dataIndex: 'definition_name',
                render: (value: string, item: WorkflowInstance) => (
                  <Button type="link" onClick={() => open(item.id)}>
                    {value || tx('审批流程')}
                  </Button>
                ),
              },
              { title: tx('发起人'), dataIndex: 'initiator_name' },
              {
                title: tx('状态'),
                render: (_, item: WorkflowInstance) => (
                  <Tag
                    color={
                      item.status === 0 ? 'processing' : item.status === 1 ? 'success' : 'default'
                    }
                  >
                    {instanceLabels[item.status]}
                  </Tag>
                ),
              },
              { title: tx('发起时间'), dataIndex: 'started_at', render: dateLabel },
              { title: tx('结束时间'), dataIndex: 'ended_at', render: dateLabel },
              {
                title: '',
                render: (_, item: WorkflowInstance) => (
                  <Button size="small" onClick={() => open(item.id)}>
                    {tx('查看进度')}
                  </Button>
                ),
              },
            ]}
          />
        ) : (
          <div className={styles.cards}>
            {items.length ? (
              items.map((item) => (
                <article className={styles.card} key={item.id}>
                  <div className={styles.cardHead}>
                    <h3>{item.definition_name || tx('审批流程')}</h3>
                    <Tag>{instanceLabels[item.status]}</Tag>
                  </div>
                  <p>
                    {item.initiator_name || '—'} · {dateLabel(item.started_at)}
                  </p>
                  <Button onClick={() => open(item.id)}>{tx('查看进度')}</Button>
                </article>
              ))
            ) : (
              <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description={tx('暂无符合条件的流程')} />
            )}
          </div>
        )}
        {!!total && (
          <div className={styles.footer}>
            <Pagination
              simple={!screens.md}
              current={page}
              pageSize={12}
              total={total}
              showSizeChanger={false}
              onChange={setPage}
            />
          </div>
        )}
      </section>
      <InstanceDrawer
        id={params.get('instance_id')}
        onClose={() => setParams({})}
        onChanged={() => void load()}
      />
    </>
  );
}
