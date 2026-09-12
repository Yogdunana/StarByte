import React, { useCallback, useEffect, useState } from 'react';
import { Button, Card, Input, Modal, Select, Space, Table, Tag, message } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { useTranslation } from 'react-i18next';
import { approveLeave, getLeaveList, getLeaveStats, getLeaveTodos, rejectLeave } from '@/api/leave';
import type { LeaveApplication, LeaveStats, LeaveStatus } from '@/api/leave';
import { usePermission } from '@/hooks/usePermission';
import { formatDateTime } from '@/utils/format';
import { LeaveStatuses, leaveStageLabel, leaveStatusMap, leaveTypeLabel } from './meta';
import './leave.css';

const ListPage: React.FC = () => {
  const { t } = useTranslation();
  const canApprove = usePermission('leave:approve');
  const statusMap = leaveStatusMap(t);
  const [list, setList] = useState<LeaveApplication[]>([]);
  const [stats, setStats] = useState<LeaveStats | null>(null);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [status, setStatus] = useState<LeaveStatus | ''>('pending');
  const [loading, setLoading] = useState(false);

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const inbox = canApprove && status === 'pending';
      const [res, stat] = await Promise.all([
        inbox
          ? getLeaveTodos({ page, page_size: 10 })
          : getLeaveList({ page, page_size: 10, status: status || undefined }),
        getLeaveStats(),
      ]);
      setList(res.list || []);
      setTotal(res.total);
      setStats(stat);
    } finally {
      setLoading(false);
    }
  }, [canApprove, page, status]);

  useEffect(() => { void load(); }, [load]);

  const decide = (row: LeaveApplication, approve: boolean) => {
    let remark = '';
    Modal.confirm({
      title: approve ? t('leave.approve') : t('leave.reject'),
      content: (
        <Input.TextArea
          rows={3}
          maxLength={500}
          placeholder={t('leave.remark')}
          onChange={(e) => { remark = e.target.value; }}
        />
      ),
      onOk: async () => {
        if (approve) await approveLeave(row.id, remark);
        else await rejectLeave(row.id, remark);
        message.success(approve ? t('leave.approved') : t('leave.rejected'));
        void load();
      },
    });
  };

  const columns: ColumnsType<LeaveApplication> = [
    { title: t('leave.applicant'), render: (_, row) => row.applicant.name || row.applicant.id },
    {
      title: t('leave.typeLabel'),
      render: (_, row) => leaveTypeLabel(t, row.leave_type.code, row.leave_type.name),
    },
    {
      title: t('leave.range'),
      render: (_, row) => `${formatDateTime(row.start_time, 'YYYY-MM-DD HH:mm')} ~ ${formatDateTime(row.end_time, 'YYYY-MM-DD HH:mm')}`,
    },
    { title: t('leave.days'), dataIndex: 'duration_days', width: 80 },
    { title: t('leave.reason'), dataIndex: 'reason', ellipsis: true },
    {
      title: t('leave.attachments'),
      render: (_, row) => row.attachments?.length || 0,
      width: 80,
    },
    {
      title: t('leave.stageLabel'),
      render: (_, row) => row.status === 'pending'
        ? (row.workflow_stage ? <Tag color="blue">{leaveStageLabel(t, row.workflow_stage)}</Tag> : t('leave.legacySingle'))
        : '—',
    },
    {
      title: t('leave.statusLabel'),
      dataIndex: 'status',
      width: 100,
      render: (v: LeaveStatus) => <Tag color={statusMap[v]?.color}>{statusMap[v]?.text}</Tag>,
    },
  ];

  if (canApprove) {
    columns.push({
      title: t('common.actions'),
      width: 160,
      render: (_, row) => row.status === 'pending' ? (
        <Space>
          <Button type="link" size="small" onClick={() => decide(row, true)}>{t('leave.approve')}</Button>
          <Button type="link" size="small" danger onClick={() => decide(row, false)}>{t('leave.reject')}</Button>
        </Space>
      ) : null,
    });
  }

  return (
    <div>
      <div className="leave-hero">
        <div>
          <h2>{t('leave.todoTitle')}</h2>
          <p>{t('leave.listHint', { total: stats?.total ?? 0, pending: stats?.by_status?.pending ?? 0 })}</p>
        </div>
      </div>
      <Card>
        <Space style={{ marginBottom: 16 }}>
          <Select
            allowClear
            placeholder={t('leave.statusLabel')}
            style={{ width: 160 }}
            value={status || undefined}
            onChange={(v) => { setPage(1); setStatus(v || ''); }}
            options={LeaveStatuses.map((s) => ({ value: s, label: statusMap[s].text }))}
          />
        </Space>
        <Table
          rowKey="id"
          loading={loading}
          columns={columns}
          dataSource={list}
          pagination={{ current: page, total, pageSize: 10, onChange: setPage }}
        />
      </Card>
    </div>
  );
};

export default ListPage;
