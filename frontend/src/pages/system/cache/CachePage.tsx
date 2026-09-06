import React, { useCallback, useEffect, useMemo, useState } from 'react';
import {
  Button, Card, Form, Input, InputNumber, Modal, Space, Statistic, Table, Tag, message,
} from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { DeleteOutlined, FireOutlined, ReloadOutlined } from '@ant-design/icons';
import { deleteCacheKey, deleteCachePattern, getCacheStats, warmupCache } from '@/api/cache';
import { usePermissions } from '@/hooks/usePermission';
import type { CacheKeyInfo, CacheStats } from '@/types/api';
import './cache.css';

const emptyStats: CacheStats = {
  healthy: false,
  pool: {
    ping_ok: false, hits: 0, misses: 0, timeouts: 0,
    total_conns: 0, idle_conns: 0, stale_conns: 0,
  },
  l1_hits: 0, l1_misses: 0, l1_size: 0, keys: [], key_count: 0, pattern: '*',
};

interface WarmForm {
  key: string;
  value: string;
  ttl_seconds?: number;
  scan_prefix?: string;
}

function ttlLabel(sec: number): string {
  if (sec < 0) return '永不过期';
  if (sec < 60) return `${sec} 秒`;
  if (sec < 3600) return `${Math.floor(sec / 60)} 分`;
  return `${Math.floor(sec / 3600)} 小时`;
}

const CachePage: React.FC = () => {
  const [canDelete, canManage] = usePermissions(['cache:delete', 'cache:manage']);
  const [stats, setStats] = useState<CacheStats>(emptyStats);
  const [loading, setLoading] = useState(false);
  const [pattern, setPattern] = useState('demo:*');
  const [warmOpen, setWarmOpen] = useState(false);
  const [form] = Form.useForm<WarmForm>();

  const load = useCallback(async (p?: string) => {
    setLoading(true);
    try {
      setStats(await getCacheStats(p || pattern || undefined) || emptyStats);
    } finally {
      setLoading(false);
    }
  }, [pattern]);

  useEffect(() => { void load(); }, [load]);

  const hitRatio = useMemo(() => {
    const total = stats.l1_hits + stats.l1_misses;
    if (!total) return 0;
    return Math.round((stats.l1_hits / total) * 100);
  }, [stats.l1_hits, stats.l1_misses]);

  const onDeleteKey = (row: CacheKeyInfo) => {
    Modal.confirm({
      title: '清除该缓存键',
      content: row.key,
      okType: 'danger',
      onOk: async () => {
        await deleteCacheKey(row.key);
        message.success('已删除');
        void load();
      },
    });
  };

  const onDeletePattern = () => {
    Modal.confirm({
      title: '按模式批量清除',
      content: `将删除匹配 ${pattern} 的键（拒绝过于宽泛的 *）。`,
      okType: 'danger',
      onOk: async () => {
        const res = await deleteCachePattern(pattern);
        message.success(`已删除 ${res?.deleted ?? 0} 个键`);
        void load();
      },
    });
  };

  const onWarmup = async () => {
    const v = await form.validateFields();
    await warmupCache({
      entries: v.key ? [{ key: v.key, value: v.value || '', ttl_seconds: v.ttl_seconds ?? 60 }] : [],
      scan_prefix: v.scan_prefix || undefined,
    });
    message.success('预热完成');
    setWarmOpen(false);
    form.resetFields();
    void load(v.key ? `${v.key.split(':')[0]}:*` : pattern);
  };

  const columns: ColumnsType<CacheKeyInfo> = [
    {
      title: '键',
      dataIndex: 'key',
      render: (v: string) => <span className="cache-mono">{v}</span>,
    },
    {
      title: 'TTL',
      width: 140,
      dataIndex: 'ttl_seconds',
      render: (v: number) => <Tag color={v < 0 ? 'purple' : 'blue'}>{ttlLabel(v)}</Tag>,
    },
    {
      title: '操作',
      width: 100,
      render: (_, row) => canDelete
        ? <Button type="link" size="small" danger onClick={() => onDeleteKey(row)}>删除</Button>
        : null,
    },
  ];

  return (
    <div>
      <div className="cache-hero">
        <div>
          <h2>缓存管理</h2>
          <p>查看 Redis 连接池与 L1 命中，按键或模式清除，预热热点数据。值不在列表中展示。</p>
        </div>
        <Space>
          {canManage && (
            <Button size="large" icon={<FireOutlined />} onClick={() => setWarmOpen(true)}>预热</Button>
          )}
          <Button size="large" icon={<ReloadOutlined />} onClick={() => void load()}>刷新</Button>
        </Space>
      </div>
      <div className="cache-stats">
        <Card>
          <Statistic title="Redis" value={stats.healthy ? '健康' : '异常'} />
        </Card>
        <Card>
          <Statistic title="连接" value={stats.pool.total_conns} suffix={`空闲 ${stats.pool.idle_conns}`} />
        </Card>
        <Card>
          <Statistic title="L1 条目" value={stats.l1_size} />
        </Card>
        <Card>
          <Statistic title="L1 命中率" value={hitRatio} suffix="%" />
        </Card>
      </div>
      <Card className="cache-shell">
        <Space style={{ marginBottom: 16 }} wrap>
          <Input.Search
            allowClear
            placeholder="SCAN 模式，如 dict:*"
            defaultValue={pattern}
            onSearch={(v) => { setPattern(v || '*'); void load(v || '*'); }}
            style={{ width: 280 }}
          />
          {canDelete && (
            <Button danger icon={<DeleteOutlined />} onClick={onDeletePattern}>按模式清除</Button>
          )}
        </Space>
        <Table
          rowKey="key"
          loading={loading}
          columns={columns}
          dataSource={stats.keys || []}
          pagination={{ pageSize: 20 }}
        />
      </Card>
      <Modal
        title="缓存预热"
        open={warmOpen}
        onCancel={() => setWarmOpen(false)}
        onOk={() => void onWarmup()}
        destroyOnClose
      >
        <Form form={form} layout="vertical" initialValues={{ ttl_seconds: 300 }}>
          <Form.Item name="key" label="键">
            <Input placeholder="demo:hot" />
          </Form.Item>
          <Form.Item name="value" label="值">
            <Input.TextArea rows={3} placeholder="可选，写入 L1+L2" />
          </Form.Item>
          <Form.Item name="ttl_seconds" label="TTL（秒，0 表示永不过期）">
            <InputNumber min={0} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="scan_prefix" label="或从 Redis 扫描前缀加载到 L1">
            <Input placeholder="dict:*" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default CachePage;
