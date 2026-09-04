import { Button, Space, Tag, Typography } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import type { FileInfo } from '@/types/api';

const categoryColor: Record<string, string> = {
  image: 'blue',
  document: 'green',
  video: 'purple',
};

const categoryLabel: Record<string, string> = {
  image: '图片',
  document: '文档',
  video: '视频',
};

function formatSize(size: number): string {
  if (size < 1024) return `${size} B`;
  if (size < 1024 * 1024) return `${(size / 1024).toFixed(1)} KB`;
  return `${(size / (1024 * 1024)).toFixed(1)} MB`;
}

export function buildFileColumns(
  onDownload: (record: FileInfo) => void,
  onDelete?: (record: FileInfo) => void,
): ColumnsType<FileInfo> {
  return [
    {
      title: '原始文件名',
      dataIndex: 'original_name',
      ellipsis: true,
      render: (name: string, record) => (
        <Typography.Text>{name || record.filename || record.name}</Typography.Text>
      ),
    },
    {
      title: '分类',
      dataIndex: 'category',
      width: 90,
      render: (category: string) => (
        <Tag color={categoryColor[category]}>{categoryLabel[category] || category || '-'}</Tag>
      ),
    },
    {
      title: '大小',
      dataIndex: 'file_size',
      width: 100,
      render: (fileSize: number, record) => formatSize(fileSize || record.size || 0),
    },
    {
      title: '类型',
      dataIndex: 'mime_type',
      width: 160,
      ellipsis: true,
    },
    {
      title: '上传者',
      dataIndex: 'uploader_name',
      width: 120,
      render: (name: string, record) => name || record.uploader?.name || '-',
    },
    {
      title: '上传时间',
      dataIndex: 'created_at',
      width: 180,
    },
    {
      title: '操作',
      key: 'action',
      width: 160,
      fixed: 'right',
      render: (_, record) => (
        <Space>
          <Button type="link" size="small" onClick={() => onDownload(record)}>
            下载
          </Button>
          {onDelete && (
            <Button type="link" size="small" danger onClick={() => onDelete(record)}>
              删除
            </Button>
          )}
        </Space>
      ),
    },
  ];
}
