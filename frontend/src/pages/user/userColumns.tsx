import React from 'react';
import { Button, Space, Tag } from 'antd';
import { EditOutlined, DeleteOutlined } from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import type { UserListItem } from '@/api/user';

const statusMap: Record<number, { color: string; text: string }> = {
  0: { color: 'success', text: '正常' },
  1: { color: 'error', text: '禁用' },
  2: { color: 'warning', text: '锁定' },
};

export interface UserColumnHandlers {
  onEdit: (record: UserListItem) => void;
  onDelete: (record: UserListItem) => void;
}

export function getUserColumns(handlers: UserColumnHandlers): ColumnsType<UserListItem> {
  return [
    { title: '用户名', dataIndex: 'username', key: 'username', width: 120 },
    { title: '真实姓名', dataIndex: 'real_name', key: 'real_name', width: 100 },
    { title: '邮箱', dataIndex: 'email', key: 'email', width: 180 },
    { title: '手机号', dataIndex: 'phone', key: 'phone', width: 130 },
    { title: '部门', dataIndex: 'department_name', key: 'department_name', width: 100 },
    { title: '职位', dataIndex: 'position_name', key: 'position_name', width: 100 },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 80,
      render: (status: number) => {
        const info = statusMap[status] || statusMap[0];
        return <Tag color={info.color}>{info.text}</Tag>;
      },
    },
    {
      title: '最后登录',
      dataIndex: 'last_login_at',
      key: 'last_login_at',
      width: 160,
      render: (time: string) => time || '-',
    },
    {
      title: '操作',
      key: 'action',
      width: 150,
      fixed: 'right',
      render: (_, record) => (
        <Space>
          <Button type="link" size="small" icon={<EditOutlined />} onClick={() => handlers.onEdit(record)}>
            编辑
          </Button>
          <Button type="link" size="small" danger icon={<DeleteOutlined />} onClick={() => handlers.onDelete(record)}>
            删除
          </Button>
        </Space>
      ),
    },
  ];
}
