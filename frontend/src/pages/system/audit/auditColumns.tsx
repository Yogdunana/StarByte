import { Button, Tag, Tooltip } from 'antd';
import { EyeOutlined } from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import dayjs from 'dayjs';
import type { AuditLogItem } from '@/api/audit';

export const methodColorMap: Record<string, string> = {
  POST: 'green',
  PUT: 'blue',
  PATCH: 'orange',
  DELETE: 'red',
};

export const actionColorMap: Record<string, string> = {
  CREATE: 'green',
  UPDATE: 'blue',
  DELETE: 'red',
  LOGIN: 'purple',
  LOGOUT: 'default',
};

export const statusColorMap = (status: number): string => {
  if (status >= 200 && status < 300) return 'success';
  if (status >= 300 && status < 400) return 'blue';
  if (status >= 400 && status < 500) return 'warning';
  return 'error';
};

export function buildAuditColumns(
  onView: (record: AuditLogItem) => void,
): ColumnsType<AuditLogItem> {
  return [
    {
      title: '时间',
      dataIndex: 'timestamp',
      key: 'timestamp',
      width: 170,
      render: (time: string) => (time ? dayjs(time).format('YYYY-MM-DD HH:mm:ss') : '-'),
    },
    {
      title: '用户',
      key: 'user',
      width: 120,
      render: (_, record) => record.user?.username || '未认证',
    },
    {
      title: '动作',
      dataIndex: 'action',
      key: 'action',
      width: 100,
      render: (action: string) =>
        action ? <Tag color={actionColorMap[action] || 'default'}>{action}</Tag> : '-',
    },
    {
      title: '模块',
      dataIndex: 'module',
      key: 'module',
      width: 100,
    },
    {
      title: '方法',
      dataIndex: 'method',
      key: 'method',
      width: 80,
      render: (m: string) =>
        m ? <Tag color={methodColorMap[m] || 'default'}>{m}</Tag> : '-',
    },
    {
      title: '路径',
      dataIndex: 'path',
      key: 'path',
      width: 240,
      ellipsis: { showTitle: false },
      render: (path: string) => (
        <Tooltip title={path}>
          <span style={{ fontFamily: 'monospace', fontSize: 12 }}>{path}</span>
        </Tooltip>
      ),
    },
    {
      title: 'IP',
      dataIndex: 'ip_address',
      key: 'ip_address',
      width: 130,
    },
    {
      title: '状态码',
      dataIndex: 'response_code',
      key: 'response_code',
      width: 90,
      render: (status: number) => <Tag color={statusColorMap(status)}>{status}</Tag>,
    },
    {
      title: '耗时',
      dataIndex: 'duration_ms',
      key: 'duration_ms',
      width: 80,
      render: (ms: number) => `${ms}ms`,
    },
    {
      title: '操作',
      key: 'action_btn',
      width: 80,
      fixed: 'right',
      render: (_, record) => (
        <Button type="link" size="small" icon={<EyeOutlined />} onClick={() => onView(record)}>
          详情
        </Button>
      ),
    },
  ];
}

export function formatJSON(str: string): string {
  if (!str) return '-';
  try {
    return JSON.stringify(JSON.parse(str), null, 2);
  } catch {
    return str;
  }
}
