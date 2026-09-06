import React, { useCallback, useEffect, useMemo, useState } from 'react';
import {
  Button, Card, Descriptions, Drawer, Input, Modal, Select, Space, Table, Tag, message,
} from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { ReloadOutlined } from '@ant-design/icons';
import { getSessions, getUserSessions, kickSession, kickUserSessions } from '@/api/session';
import { usePermission } from '@/hooks/usePermission';
import type { AuthSession, UserAuthSessions } from '@/types/api';
import './session.css';

const deviceOptions = [
  { value: 'Desktop', label: '桌面' },
  { value: 'Mobile', label: '手机' },
  { value: 'Tablet', label: '平板' },
  { value: 'Unknown', label: '未知' },
];

const SessionPage: React.FC = () => {
  const canKick = usePermission('session:delete');
  const [list, setList] = useState<AuthSession[]>([]);
  const [loading, setLoading] = useState(false);
  const [keyword, setKeyword] = useState('');
  const [device, setDevice] = useState<string>();
  const [detail, setDetail] = useState<UserAuthSessions | null>(null);
  const [detailOpen, setDetailOpen] = useState(false);

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const res = await getSessions({ keyword: keyword || undefined });
      setList(res?.list || []);
    } finally {
      setLoading(false);
    }
  }, [keyword]);

  useEffect(() => { void load(); }, [load]);

  useEffect(() => {
    const timer = window.setInterval(() => { void load(); }, 10000);
    return () => window.clearInterval(timer);
  }, [load]);

  const rows = useMemo(
    () => (device ? list.filter((s) => s.device === device) : list),
    [list, device],
  );

  const openDetail = async (row: AuthSession) => {
    const data = await getUserSessions(row.user_id);
    setDetail(data);
    setDetailOpen(true);
  };

  const onKickOne = (row: AuthSession) => {
    Modal.confirm({
      title: '强制下线该会话',
      content: `${row.username || row.user_id} · ${row.ip} · ${row.browser}/${row.os}`,
      okType: 'danger',
      onOk: async () => {
        await kickSession(row.token_id);
        message.success('已强制下线');
        void load();
      },
    });
  };

  const onKickUser = (row: AuthSession) => {
    Modal.confirm({
      title: '下线该用户全部会话',
      content: `将立即失效 ${row.username || row.user_id} 的所有在线设备。`,
      okType: 'danger',
      onOk: async () => {
        await kickUserSessions(row.user_id);
        message.success('已下线该用户全部会话');
        void load();
        setDetailOpen(false);
      },
    });
  };

  const columns: ColumnsType<AuthSession> = useMemo(() => [
    { title: '用户', dataIndex: 'username', width: 120, render: (v: string, row) => v || row.user_id.slice(0, 8) },
    { title: '姓名', dataIndex: 'real_name', width: 100 },
    { title: 'IP', dataIndex: 'ip', width: 130 },
    {
      title: '设备',
      width: 200,
      render: (_, row) => (
        <Space size={4} wrap>
          <Tag>{row.device}</Tag>
          <span>{row.browser} / {row.os}</span>
        </Space>
      ),
    },
    {
      title: '异常',
      width: 140,
      render: (_, row) => (
        <Space size={4}>
          {row.multi_device && <Tag color="orange">多设备</Tag>}
          {row.multi_ip && <Tag color="red">异地</Tag>}
          {!row.multi_device && !row.multi_ip && <Tag>正常</Tag>}
        </Space>
      ),
    },
    { title: '登录时间', dataIndex: 'login_at', width: 180, render: (v: string) => v?.replace('T', ' ').slice(0, 19) },
    {
      title: '操作',
      width: 200,
      render: (_, row) => (
        <Space>
          <Button type="link" size="small" onClick={() => void openDetail(row)}>详情</Button>
          {canKick && <Button type="link" size="small" danger onClick={() => onKickOne(row)}>下线</Button>}
          {canKick && <Button type="link" size="small" danger onClick={() => onKickUser(row)}>下线全部</Button>}
        </Space>
      ),
    },
  ], [canKick]);

  return (
    <div>
      <div className="sess-hero">
        <div>
          <h2>在线会话</h2>
          <p>查看当前有效 Access Token，强制下线立即拉黑并清除 Refresh Token。列表每 10 秒刷新。</p>
        </div>
        <Button size="large" icon={<ReloadOutlined />} onClick={() => void load()}>立即刷新</Button>
      </div>
      <Card className="sess-shell">
        <Space style={{ marginBottom: 16 }} wrap>
          <Input.Search
            allowClear
            placeholder="用户 / IP / 浏览器"
            onSearch={setKeyword}
            style={{ width: 240 }}
          />
          <Select
            allowClear
            placeholder="设备类型"
            style={{ width: 140 }}
            value={device}
            options={deviceOptions}
            onChange={(v) => setDevice(v)}
          />
        </Space>
        <Table
          rowKey="token_id"
          loading={loading}
          columns={columns}
          dataSource={rows}
          pagination={{ pageSize: 20 }}
        />
      </Card>
      <Drawer
        title="会话详情"
        width={520}
        open={detailOpen}
        onClose={() => setDetailOpen(false)}
      >
        {detail && (
          <>
            <Descriptions column={1} size="small" style={{ marginBottom: 16 }}>
              <Descriptions.Item label="用户">{detail.username || detail.user_id}</Descriptions.Item>
              <Descriptions.Item label="姓名">{detail.real_name || '-'}</Descriptions.Item>
              <Descriptions.Item label="异常">
                <Space>
                  {detail.multi_device && <Tag color="orange">多设备</Tag>}
                  {detail.multi_ip && <Tag color="red">异地</Tag>}
                  {!detail.multi_device && !detail.multi_ip && <Tag>正常</Tag>}
                </Space>
              </Descriptions.Item>
            </Descriptions>
            {detail.sessions?.map((s) => (
              <Card key={s.token_id} size="small" style={{ marginBottom: 12 }}>
                <p>{s.ip} · {s.browser} / {s.os} · {s.device}</p>
                <p className="sess-ua">{s.user_agent || '-'}</p>
                <p>登录 {s.login_at?.replace('T', ' ').slice(0, 19)}</p>
                {canKick && (
                  <Button danger size="small" onClick={() => onKickOne(s)}>强制下线</Button>
                )}
              </Card>
            ))}
          </>
        )}
      </Drawer>
    </div>
  );
};

export default SessionPage;
