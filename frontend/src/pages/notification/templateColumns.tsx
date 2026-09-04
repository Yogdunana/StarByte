import { Button, Space, Tag, Typography, Popconfirm } from 'antd';
import {
  EditOutlined,
  DeleteOutlined,
  ExperimentOutlined,
} from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import type { NotificationTemplate } from '@/types/api';
import { statusMap, channelColorMap } from './templateMeta';

const { Text } = Typography;

export interface TemplateColumnHandlers {
  onTest: (record: NotificationTemplate) => void;
  onEdit: (record: NotificationTemplate) => void;
  onDelete: (record: NotificationTemplate) => void;
}

export function getTemplateColumns(
  handlers: TemplateColumnHandlers,
): ColumnsType<NotificationTemplate> {
  return [
    {
      title: '编码',
      dataIndex: 'code',
      key: 'code',
      width: 180,
      render: (code: string) => <Text code>{code}</Text>,
    },
    {
      title: '名称',
      dataIndex: 'name',
      key: 'name',
      width: 160,
    },
    {
      title: '分类',
      dataIndex: 'category',
      key: 'category',
      width: 100,
      render: (cat: string) => cat || '-',
    },
    {
      title: '渠道',
      dataIndex: 'channels',
      key: 'channels',
      width: 200,
      render: (channels: string[]) => (
        <Space wrap size={[4, 4]}>
          {channels?.map((ch) => (
            <Tag key={ch} color={channelColorMap[ch] || 'default'}>
              {ch}
            </Tag>
          ))}
        </Space>
      ),
    },
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
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 160,
      render: (time: string) =>
        new Date(time).toLocaleString('zh-CN', { hour12: false }),
    },
    {
      title: '操作',
      key: 'action',
      width: 200,
      fixed: 'right',
      render: (_: unknown, record: NotificationTemplate) => (
        <Space>
          <Button
            type="link"
            size="small"
            icon={<ExperimentOutlined />}
            onClick={() => handlers.onTest(record)}
          >
            测试
          </Button>
          <Button
            type="link"
            size="small"
            icon={<EditOutlined />}
            onClick={() => handlers.onEdit(record)}
          >
            编辑
          </Button>
          <Popconfirm
            title="确认删除此模板？"
            onConfirm={() => handlers.onDelete(record)}
          >
            <Button type="link" size="small" danger icon={<DeleteOutlined />}>
              删除
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];
}
