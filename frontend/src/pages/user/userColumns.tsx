import { tx } from '@/i18n/text';
import { Button, Space, Tag } from 'antd';
import { EditOutlined, DeleteOutlined } from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import type { UserListItem } from '@/api/user';

const statusMap: Record<number, { color: string; text: string }> = {
  0: {
    color: 'success',
    get text() {
      return tx('正常');
    },
  },
  1: {
    color: 'error',
    get text() {
      return tx('禁用');
    },
  },
  2: {
    color: 'warning',
    get text() {
      return tx('锁定');
    },
  },
};

export interface UserColumnHandlers {
  onEdit: (record: UserListItem) => void;
  onDelete: (record: UserListItem) => void;
}

export function getUserColumns(handlers: UserColumnHandlers): ColumnsType<UserListItem> {
  return [
    { title: tx('用户名'), dataIndex: 'username', key: 'username', width: 120 },
    { title: tx('真实姓名'), dataIndex: 'real_name', key: 'real_name', width: 100 },
    { title: tx('邮箱'), dataIndex: 'email', key: 'email', width: 180 },
    { title: tx('手机号'), dataIndex: 'phone', key: 'phone', width: 130 },
    { title: tx('部门'), dataIndex: 'department_name', key: 'department_name', width: 100 },
    { title: tx('职位'), dataIndex: 'position_name', key: 'position_name', width: 100 },
    {
      title: tx('状态'),
      dataIndex: 'status',
      key: 'status',
      width: 80,
      render: (status: number) => {
        const info = statusMap[status] || statusMap[0];
        return <Tag color={info.color}>{info.text}</Tag>;
      },
    },
    {
      title: tx('最后登录'),
      dataIndex: 'last_login_at',
      key: 'last_login_at',
      width: 160,
      render: (time: string) => time || '-',
    },
    {
      title: tx('操作'),
      key: 'action',
      width: 150,
      fixed: 'right',
      render: (_, record) => (
        <Space>
          <Button
            type="link"
            size="small"
            icon={<EditOutlined />}
            onClick={() => handlers.onEdit(record)}
          >
            {tx('编辑')}
          </Button>
          <Button
            type="link"
            size="small"
            danger
            icon={<DeleteOutlined />}
            onClick={() => handlers.onDelete(record)}
          >
            {tx('删除')}
          </Button>
        </Space>
      ),
    },
  ];
}
