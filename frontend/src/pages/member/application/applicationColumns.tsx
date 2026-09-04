import type { ColumnsType } from 'antd/es/table';
import { Button, Space } from 'antd';
import StatusTag from '@/components/StatusTag/StatusTag';
import type { MemberApplication } from '@/types/api';
import { ApplicantTypeMap, ApplicationStatusMap } from '../meta';

interface ColumnOptions {
  showReview?: boolean;
  onView: (record: MemberApplication) => void;
  onReview?: (record: MemberApplication) => void;
  onResubmit?: (record: MemberApplication) => void;
}

export function buildApplicationColumns(opts: ColumnOptions): ColumnsType<MemberApplication> {
  return [
    { title: '姓名', dataIndex: 'real_name', width: 100 },
    { title: '学号', dataIndex: 'student_no', width: 120 },
    {
      title: '类型',
      dataIndex: 'applicant_type',
      width: 80,
      render: (v: number) => <StatusTag status={v} mapping={ApplicantTypeMap} />,
    },
    { title: '意向部门', dataIndex: 'department_name', width: 120, render: (v?: string) => v || '-' },
    {
      title: '状态',
      dataIndex: 'status',
      width: 110,
      render: (v: number) => <StatusTag status={v} mapping={ApplicationStatusMap} />,
    },
    { title: '提交时间', dataIndex: 'submitted_at', width: 180 },
    {
      title: '操作',
      key: 'action',
      width: 180,
      render: (_, record) => (
        <Space>
          <Button type="link" size="small" onClick={() => opts.onView(record)}>
            详情
          </Button>
          {opts.showReview && opts.onReview && record.status !== 3 && record.status !== 4 && (
            <Button type="link" size="small" onClick={() => opts.onReview?.(record)}>
              审核
            </Button>
          )}
          {opts.onResubmit && record.status === 5 && (
            <Button type="link" size="small" onClick={() => opts.onResubmit?.(record)}>
              补充
            </Button>
          )}
        </Space>
      ),
    },
  ];
}
