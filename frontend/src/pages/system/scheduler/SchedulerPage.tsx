import { tx, useLocale } from '@/i18n/text';
import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { Button, Card, Form, Input, Modal, Space, Statistic, Table, Tag, message } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { PlusOutlined, ReloadOutlined } from '@ant-design/icons';
import {
  createSchedulerTask,
  deleteSchedulerTask,
  getSchedulerHandlers,
  getSchedulerLogs,
  getSchedulerTasks,
  pauseSchedulerTask,
  resumeSchedulerTask,
  runSchedulerTask,
  updateSchedulerTask,
} from '@/api/scheduler';
import { usePermissions } from '@/hooks/usePermission';
import type { SchedulerHandlerInfo, SchedulerLogs, SchedulerTask } from '@/types/api';
import TaskEditor, { fillTaskForm, type TaskFormValues } from './TaskEditor';
import RunDrawer from './RunDrawer';
import './scheduler.css';

const statusMeta: Record<number, { color: string; text: string }> = {
  0: {
    color: 'green',
    get text() {
      return tx('运行中');
    },
  },
  1: {
    color: 'gold',
    get text() {
      return tx('已暂停');
    },
  },
  3: {
    color: 'default',
    get text() {
      return tx('已结束');
    },
  },
};

const SchedulerPage: React.FC = () => {
  const uiLanguage = useLocale();
  const [canCreate, canUpdate, canDelete, canRun, canManage] = usePermissions([
    'scheduler:create',
    'scheduler:update',
    'scheduler:delete',
    'scheduler:run',
    'scheduler:manage',
  ]);
  const [list, setList] = useState<SchedulerTask[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(false);
  const [keyword, setKeyword] = useState('');
  const [handlers, setHandlers] = useState<SchedulerHandlerInfo[]>([]);
  const [open, setOpen] = useState(false);
  const [editing, setEditing] = useState<SchedulerTask | null>(null);
  const [form] = Form.useForm<TaskFormValues>();
  const [logOpen, setLogOpen] = useState(false);
  const [logs, setLogs] = useState<SchedulerLogs | null>(null);
  const [logTaskId, setLogTaskId] = useState<string>();

  const load = useCallback(
    async (p = page) => {
      setLoading(true);
      try {
        const res = await getSchedulerTasks({
          page: p,
          page_size: 10,
          keyword: keyword || undefined,
        });
        setList(res?.list || []);
        setTotal(res?.total || 0);
      } finally {
        setLoading(false);
      }
    },
    [page, keyword],
  );

  useEffect(() => {
    void load();
  }, [load]);
  useEffect(() => {
    void getSchedulerHandlers()
      .then((h) => setHandlers(h || []))
      .catch(() => undefined);
  }, []);

  const stats = useMemo(() => {
    void uiLanguage;
    return {
      total,
      active: list.filter((t) => t.status === 0).length,
      paused: list.filter((t) => t.status === 1).length,
      dead: list.filter((t) => t.last_status === 'dead').length,
    };
  }, [list, total, uiLanguage]);

  const openCreate = () => {
    setEditing(null);
    form.resetFields();
    setOpen(true);
  };
  const openEdit = (row: SchedulerTask) => {
    setEditing(row);
    form.setFieldsValue(fillTaskForm(row));
    setOpen(true);
  };

  const onSave = async () => {
    const v = await form.validateFields();
    const body = {
      name: v.name,
      code: v.code,
      handler_key: v.handler_key,
      payload: v.payload || '',
      timezone: v.timezone,
      max_retries: v.max_retries,
      timeout_sec: v.timeout_sec,
      shard_key: v.shard_key || '',
      cron_expr: v.schedule_kind === 'cron' ? v.cron_expr : '',
      run_at: v.schedule_kind === 'once' && v.run_at ? v.run_at.toISOString() : undefined,
    };
    if (editing) {
      await updateSchedulerTask(editing.id, body);
      message.success(tx('已更新'));
    } else {
      await createSchedulerTask(body);
      message.success(tx('已创建'));
    }
    setOpen(false);
    void load();
  };

  const openLogs = async (id: string, runId?: string) => {
    setLogTaskId(id);
    setLogs(await getSchedulerLogs(id, runId));
    setLogOpen(true);
  };

  const columns: ColumnsType<SchedulerTask> = [
    { title: tx('名称'), dataIndex: 'name' },
    {
      title: tx('编码'),
      dataIndex: 'code',
      render: (v: string) => <span className="sched-mono">{v}</span>,
    },
    { title: 'Cron', dataIndex: 'cron_expr', render: (v: string) => v || tx('一次性') },
    { title: tx('处理器'), dataIndex: 'handler_key' },
    {
      title: tx('状态'),
      dataIndex: 'status',
      width: 90,
      render: (v: number) => <Tag color={statusMeta[v]?.color}>{statusMeta[v]?.text || v}</Tag>,
    },
    { title: tx('上次'), dataIndex: 'last_status', width: 100, render: (v: string) => v || '-' },
    {
      title: tx('操作'),
      width: 280,
      render: (_, row) => (
        <Space wrap size="small">
          <Button type="link" size="small" onClick={() => void openLogs(row.id)}>
            {tx('日志')}
          </Button>
          {canRun && (
            <Button
              type="link"
              size="small"
              onClick={() => {
                void runSchedulerTask(row.id).then(() => {
                  message.success(tx('已触发'));
                  void load();
                });
              }}
            >
              {tx('执行')}
            </Button>
          )}
          {canManage && row.status === 0 && (
            <Button
              type="link"
              size="small"
              onClick={() => {
                void pauseSchedulerTask(row.id).then(() => void load());
              }}
            >
              {tx('暂停')}
            </Button>
          )}
          {canManage && row.status === 1 && (
            <Button
              type="link"
              size="small"
              onClick={() => {
                void resumeSchedulerTask(row.id).then(() => void load());
              }}
            >
              {tx('恢复')}
            </Button>
          )}
          {canUpdate && (
            <Button type="link" size="small" onClick={() => openEdit(row)}>
              {tx('编辑')}
            </Button>
          )}
          {canDelete && (
            <Button
              type="link"
              size="small"
              danger
              onClick={() => {
                Modal.confirm({
                  title: tx('删除任务'),
                  onOk: async () => {
                    await deleteSchedulerTask(row.id);
                    void load();
                  },
                });
              }}
            >
              {tx('删除')}
            </Button>
          )}
        </Space>
      ),
    },
  ];

  return (
    <div>
      <div className="sched-hero">
        <div>
          <h2>{tx('定时任务')}</h2>
          <p>{tx('Cron / 一次性任务、失败重试与分布式锁。同一任务同一时刻只在一个节点执行。')}</p>
        </div>
        <Space>
          {canCreate && (
            <Button size="large" type="primary" icon={<PlusOutlined />} onClick={openCreate}>
              {tx('新建')}
            </Button>
          )}
          <Button size="large" icon={<ReloadOutlined />} onClick={() => void load()}>
            {tx('刷新')}
          </Button>
        </Space>
      </div>
      <div className="sched-stats">
        <Card>
          <Statistic title={tx('任务总数')} value={stats.total} />
        </Card>
        <Card>
          <Statistic title={tx('本页运行中')} value={stats.active} />
        </Card>
        <Card>
          <Statistic title={tx('本页已暂停')} value={stats.paused} />
        </Card>
        <Card>
          <Statistic title={tx('本页死信')} value={stats.dead} />
        </Card>
      </div>
      <Card className="sched-shell">
        <Input.Search
          allowClear
          placeholder={tx('搜索名称或编码')}
          onSearch={(v) => {
            setKeyword(v);
            setPage(1);
          }}
          style={{ width: 280, marginBottom: 16 }}
        />
        <Table
          rowKey="id"
          loading={loading}
          columns={columns}
          dataSource={list}
          pagination={{
            current: page,
            pageSize: 10,
            total,
            onChange: (p) => {
              setPage(p);
            },
          }}
        />
      </Card>
      <TaskEditor
        open={open}
        editing={editing}
        handlers={handlers}
        form={form}
        onCancel={() => setOpen(false)}
        onOk={() => void onSave()}
      />
      <RunDrawer
        open={logOpen}
        logs={logs}
        onClose={() => setLogOpen(false)}
        onSelectRun={(rid) => {
          if (logTaskId) void openLogs(logTaskId, rid);
        }}
      />
    </div>
  );
};

export default SchedulerPage;
