import { tx, useLocale } from '@/i18n/text';
import React, { useCallback, useEffect, useMemo, useState } from 'react';
import {
  Button,
  Card,
  Descriptions,
  Drawer,
  Input,
  Modal,
  Select,
  Space,
  Table,
  Tag,
  message,
} from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { ReloadOutlined } from '@ant-design/icons';
import { getSessions, getUserSessions, kickSession, kickUserSessions } from '@/api/session';
import { usePermission } from '@/hooks/usePermission';
import type { AuthSession, UserAuthSessions } from '@/types/api';
import './session.css';

const deviceOptions = [
  {
    value: 'Desktop',
    get label() {
      return tx('桌面');
    },
  },
  {
    value: 'Mobile',
    get label() {
      return tx('手机');
    },
  },
  {
    value: 'Tablet',
    get label() {
      return tx('平板');
    },
  },
  {
    value: 'Unknown',
    get label() {
      return tx('未知');
    },
  },
];

const SessionPage: React.FC = () => {
  const uiLanguage = useLocale();
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

  useEffect(() => {
    void load();
  }, [load]);

  useEffect(() => {
    const timer = window.setInterval(() => {
      void load();
    }, 10000);
    return () => window.clearInterval(timer);
  }, [load]);

  const rows = useMemo(() => {
    void uiLanguage;
    return device ? list.filter((s) => s.device === device) : list;
  }, [list, device, uiLanguage]);

  const openDetail = async (row: AuthSession) => {
    const data = await getUserSessions(row.user_id);
    setDetail(data);
    setDetailOpen(true);
  };

  const onKickOne = useCallback(
    (row: AuthSession) => {
      Modal.confirm({
        title: tx('强制下线该会话'),
        content: `${row.username || row.user_id} · ${row.ip} · ${row.browser}/${row.os}`,
        okType: 'danger',
        onOk: async () => {
          await kickSession(row.token_id);
          message.success(tx('已强制下线'));
          void load();
        },
      });
    },
    [load],
  );

  const onKickUser = useCallback(
    (row: AuthSession) => {
      Modal.confirm({
        title: tx('下线该用户全部会话'),
        content: tx('将立即失效 {{value0}} 的所有在线设备。', {
          value0: row.username || row.user_id,
        }),
        okType: 'danger',
        onOk: async () => {
          await kickUserSessions(row.user_id);
          message.success(tx('已下线该用户全部会话'));
          void load();
          setDetailOpen(false);
        },
      });
    },
    [load],
  );

  const columns: ColumnsType<AuthSession> = useMemo(() => {
    void uiLanguage;
    return [
      {
        title: tx('用户'),
        dataIndex: 'username',
        width: 120,
        render: (v: string, row) => v || row.user_id.slice(0, 8),
      },
      { title: tx('姓名'), dataIndex: 'real_name', width: 100 },
      { title: 'IP', dataIndex: 'ip', width: 130 },
      {
        title: tx('设备'),
        width: 200,
        render: (_, row) => (
          <Space size={4} wrap>
            <Tag>{row.device}</Tag>
            <span>
              {row.browser} / {row.os}
            </span>
          </Space>
        ),
      },
      {
        title: tx('异常'),
        width: 140,
        render: (_, row) => (
          <Space size={4}>
            {row.multi_device && <Tag color="orange">{tx('多设备')}</Tag>}
            {row.multi_ip && <Tag color="red">{tx('异地')}</Tag>}
            {!row.multi_device && !row.multi_ip && <Tag>{tx('正常')}</Tag>}
          </Space>
        ),
      },
      {
        title: tx('登录时间'),
        dataIndex: 'login_at',
        width: 180,
        render: (v: string) => v?.replace('T', ' ').slice(0, 19),
      },
      {
        title: tx('操作'),
        width: 200,
        render: (_, row) => (
          <Space>
            <Button type="link" size="small" onClick={() => void openDetail(row)}>
              {tx('详情')}
            </Button>
            {canKick && (
              <Button type="link" size="small" danger onClick={() => onKickOne(row)}>
                {tx('下线')}
              </Button>
            )}
            {canKick && (
              <Button type="link" size="small" danger onClick={() => onKickUser(row)}>
                {tx('下线全部')}
              </Button>
            )}
          </Space>
        ),
      },
    ];
  }, [canKick, uiLanguage, onKickOne, onKickUser]);

  return (
    <div>
      <div className="sess-hero">
        <div>
          <h2>{tx('在线会话')}</h2>
          <p>
            {tx(
              '查看当前有效 Access Token，强制下线立即拉黑并清除 Refresh Token。列表每 10 秒刷新。',
            )}
          </p>
        </div>
        <Button size="large" icon={<ReloadOutlined />} onClick={() => void load()}>
          {tx('立即刷新')}
        </Button>
      </div>
      <Card className="sess-shell">
        <Space style={{ marginBottom: 16 }} wrap>
          <Input.Search
            allowClear
            placeholder={tx('用户 / IP / 浏览器')}
            onSearch={setKeyword}
            style={{ width: 240 }}
          />
          <Select
            allowClear
            placeholder={tx('设备类型')}
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
        title={tx('会话详情')}
        width={520}
        open={detailOpen}
        onClose={() => setDetailOpen(false)}
      >
        {detail && (
          <>
            <Descriptions column={1} size="small" style={{ marginBottom: 16 }}>
              <Descriptions.Item label={tx('用户')}>
                {detail.username || detail.user_id}
              </Descriptions.Item>
              <Descriptions.Item label={tx('姓名')}>{detail.real_name || '-'}</Descriptions.Item>
              <Descriptions.Item label={tx('异常')}>
                <Space>
                  {detail.multi_device && <Tag color="orange">{tx('多设备')}</Tag>}
                  {detail.multi_ip && <Tag color="red">{tx('异地')}</Tag>}
                  {!detail.multi_device && !detail.multi_ip && <Tag>{tx('正常')}</Tag>}
                </Space>
              </Descriptions.Item>
            </Descriptions>
            {detail.sessions?.map((s) => (
              <Card key={s.token_id} size="small" style={{ marginBottom: 12 }}>
                <p>
                  {s.ip} · {s.browser} / {s.os} · {s.device}
                </p>
                <p className="sess-ua">{s.user_agent || '-'}</p>
                <p>
                  {tx('登录')}
                  {s.login_at?.replace('T', ' ').slice(0, 19)}
                </p>
                {canKick && (
                  <Button danger size="small" onClick={() => onKickOne(s)}>
                    {tx('强制下线')}
                  </Button>
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
