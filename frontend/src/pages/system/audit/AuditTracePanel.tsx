import React, { useState } from 'react';
import { Button, Descriptions, Empty, Input, Select, Space, Table, Tag, Typography, message } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import dayjs from 'dayjs';
import { getAuditTrace, type AuditTraceItem } from '@/api/audit';
import { actionColorMap, formatJSON } from './auditColumns';

const { Text, Paragraph } = Typography;

const entityOptions = [
  { value: 'user', label: '用户 user' },
  { value: 'role', label: '角色 role' },
  { value: 'department', label: '部门 department' },
];

const AuditTracePanel: React.FC = () => {
  const [entityType, setEntityType] = useState('user');
  const [entityId, setEntityId] = useState('');
  const [loading, setLoading] = useState(false);
  const [list, setList] = useState<AuditTraceItem[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [expanded, setExpanded] = useState<string | null>(null);

  const load = async (p: number, ps: number) => {
    const id = entityId.trim();
    if (!id) {
      message.warning('请填写实体 ID');
      return;
    }
    setLoading(true);
    try {
      const res = await getAuditTrace(entityType, id, { page: p, page_size: ps });
      setList(res.list);
      setTotal(res.total);
    } catch {
      message.error('加载变更历史失败');
    } finally {
      setLoading(false);
    }
  };

  const columns: ColumnsType<AuditTraceItem> = [
    {
      title: '时间',
      dataIndex: 'timestamp',
      width: 170,
      render: (t: string) => (t ? dayjs(t).format('YYYY-MM-DD HH:mm:ss') : '-'),
    },
    {
      title: '操作人',
      width: 120,
      render: (_, row) => row.user?.username || '未认证',
    },
    {
      title: '动作',
      dataIndex: 'action',
      width: 90,
      render: (action: string) => <Tag color={actionColorMap[action] || 'default'}>{action}</Tag>,
    },
    { title: '路径', dataIndex: 'path', ellipsis: true },
    {
      title: '合规',
      dataIndex: 'compliance_flags',
      width: 160,
      render: (flags: string[] | undefined) =>
        flags && flags.length
          ? flags.map((f) => (
              <Tag key={f} color={f === 'delete' ? 'red' : f === 'export' ? 'orange' : 'purple'}>
                {f}
              </Tag>
            ))
          : '-',
    },
    {
      title: 'Diff',
      width: 80,
      render: (_, row) => (
        <Button type="link" size="small" onClick={() => setExpanded(expanded === row.id ? null : row.id)}>
          {expanded === row.id ? '收起' : '查看'}
        </Button>
      ),
    },
  ];

  return (
    <div>
      <Space wrap style={{ marginBottom: 16 }}>
        <Select value={entityType} onChange={setEntityType} options={entityOptions} style={{ width: 180 }} />
        <Input
          placeholder="实体 UUID"
          value={entityId}
          onChange={(e) => setEntityId(e.target.value)}
          style={{ width: 360 }}
          allowClear
        />
        <Button
          type="primary"
          onClick={() => {
            setPage(1);
            load(1, pageSize);
          }}
        >
          查询历史
        </Button>
      </Space>
      <Table
        rowKey="id"
        loading={loading}
        columns={columns}
        dataSource={list}
        size="middle"
        pagination={{
          current: page,
          pageSize,
          total,
          onChange: (p, ps) => {
            setPage(p);
            setPageSize(ps);
            load(p, ps);
          },
        }}
        expandable={{
          expandedRowKeys: expanded ? [expanded] : [],
          showExpandColumn: false,
          expandedRowRender: (row) => (
            <div>
              <Descriptions size="small" column={1} bordered>
                <Descriptions.Item label="变更字段">
                  {row.diff?.length
                    ? row.diff.map((d) => (
                        <Paragraph key={d.path} style={{ marginBottom: 4 }}>
                          <Text code>{d.path}</Text>：{JSON.stringify(d.before)} → {JSON.stringify(d.after)}
                        </Paragraph>
                      ))
                    : '无字段级 Diff（可能缺少 before 快照）'}
                </Descriptions.Item>
              </Descriptions>
              <div style={{ display: 'flex', gap: 12, marginTop: 12 }}>
                <div style={{ flex: 1 }}>
                  <Text strong>Before</Text>
                  <pre style={preStyle}>{formatJSON(row.before_json)}</pre>
                </div>
                <div style={{ flex: 1 }}>
                  <Text strong>After</Text>
                  <pre style={preStyle}>{formatJSON(row.after_json)}</pre>
                </div>
              </div>
            </div>
          ),
        }}
        locale={{ emptyText: <Empty description="输入实体类型与 ID 查询变更历史" /> }}
      />
    </div>
  );
};

const preStyle: React.CSSProperties = {
  background: '#f5f5f5',
  padding: 12,
  borderRadius: 4,
  maxHeight: 220,
  overflow: 'auto',
  fontSize: 12,
  whiteSpace: 'pre-wrap',
  wordBreak: 'break-all',
};

export default AuditTracePanel;
