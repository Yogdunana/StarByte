import type { ColumnsType } from 'antd/es/table';
import { Button, Space } from 'antd';
import StatusTag from '@/components/StatusTag/StatusTag';
import type { MemberProfile } from '@/types/api';
import { MemberTypeMap, ProfileStatusMap } from '../meta';

interface Options {
  onView: (record: MemberProfile) => void;
  onStatus?: (record: MemberProfile) => void;
}

export function buildProfileColumns(opts: Options): ColumnsType<MemberProfile> {
  return [
    { title: '姓名', dataIndex: 'real_name', width: 100 },
    { title: '学号', dataIndex: 'student_no', width: 120 },
    {
      title: '类型',
      dataIndex: 'member_type',
      width: 90,
      render: (v: number) => <StatusTag status={v} mapping={MemberTypeMap} />,
    },
    {
      title: '部门',
      dataIndex: ['department', 'name'],
      width: 120,
      render: (_, r) => r.department?.name || '-',
    },
    {
      title: '职位',
      dataIndex: ['position', 'name'],
      width: 100,
      render: (_, r) => r.position?.name || '-',
    },
    {
      title: '状态',
      dataIndex: 'status',
      width: 90,
      render: (v: number) => <StatusTag status={v} mapping={ProfileStatusMap} />,
    },
    { title: '入会日期', dataIndex: 'join_date', width: 120, render: (v?: string) => v?.slice(0, 10) || '-' },
    {
      title: '操作',
      key: 'action',
      width: 160,
      render: (_, record) => (
        <Space>
          <Button type="link" size="small" onClick={() => opts.onView(record)}>
            详情
          </Button>
          {opts.onStatus && (
            <Button type="link" size="small" onClick={() => opts.onStatus?.(record)}>
              状态
            </Button>
          )}
        </Space>
      ),
    },
  ];
}
