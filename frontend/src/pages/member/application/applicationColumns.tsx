import { tx } from '@/i18n/text';
import dayjs from 'dayjs';
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
    { title: tx('姓名'), dataIndex: 'real_name', width: 100 },
    { title: tx('学号'), dataIndex: 'student_no', width: 120 },
    {
      title: tx('类型'),
      dataIndex: 'applicant_type',
      width: 80,
      render: (v: number) => <StatusTag status={v} mapping={ApplicantTypeMap} />,
    },
    {
      title: tx('意向部门'),
      dataIndex: 'department_name',
      width: 120,
      render: (v?: string) => v || '-',
    },
    {
      title: tx('状态'),
      dataIndex: 'status',
      width: 110,
      render: (v: number, record) => (
        <>
          <StatusTag status={v} mapping={ApplicationStatusMap} />
          {record.current_stage && <div>{record.current_stage}</div>}
          {record.historical_review_required && <div>{tx('历史待核验')}</div>}
        </>
      ),
    },
    {
      title: tx('提交时间'),
      dataIndex: 'submitted_at',
      width: 160,
      render: (value: string) => dayjs(value).format('YYYY-MM-DD HH:mm'),
    },
    {
      title: tx('操作'),
      key: 'action',
      width: 180,
      render: (_, record) => (
        <Space>
          <Button type="link" size="small" onClick={() => opts.onView(record)}>
            {tx('详情')}
          </Button>
          {opts.showReview &&
            opts.onReview &&
            (record.status !== 3 || record.admission_stage === 'probation') &&
            record.status !== 4 && (
              <Button type="link" size="small" onClick={() => opts.onReview?.(record)}>
                {tx('审核')}
              </Button>
            )}
          {opts.onResubmit && record.status === 5 && (
            <Button type="link" size="small" onClick={() => opts.onResubmit?.(record)}>
              {tx('补充')}
            </Button>
          )}
        </Space>
      ),
    },
  ];
}
