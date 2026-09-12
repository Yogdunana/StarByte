import React, { useCallback, useEffect, useState } from 'react';
import { Button, Card, Select, Space, Table, Tag, message } from 'antd';
import { PlusOutlined } from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import { useTranslation } from 'react-i18next';
import { getLeaveBalances, getLeaveTypes, getMyLeaveList, submitLeave } from '@/api/leave';
import type { LeaveApplication, LeaveBalance, LeaveStatus, LeaveType } from '@/api/leave';
import { formatDateTime } from '@/utils/format';
import FormModal from './FormModal';
import { LeaveStatuses, leaveStatusMap, leaveTypeLabel } from './meta';
import './leave.css';

const MyPage: React.FC = () => {
  const { t } = useTranslation();
  const statusMap = leaveStatusMap(t);
  const [types, setTypes] = useState<LeaveType[]>([]);
  const [balances, setBalances] = useState<LeaveBalance[]>([]);
  const [list, setList] = useState<LeaveApplication[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [status, setStatus] = useState<LeaveStatus | ''>('');
  const [loading, setLoading] = useState(false);
  const [open, setOpen] = useState(false);

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const [typeRows, balRows, res] = await Promise.all([
        getLeaveTypes(),
        getLeaveBalances(),
        getMyLeaveList({ page, page_size: 10, status: status || undefined }),
      ]);
      setTypes(typeRows || []);
      setBalances(balRows || []);
      setList(res.list || []);
      setTotal(res.total);
    } finally {
      setLoading(false);
    }
  }, [page, status]);

  useEffect(() => { void load(); }, [load]);

  const columns: ColumnsType<LeaveApplication> = [
    {
      title: t('leave.typeLabel'),
      render: (_, row) => leaveTypeLabel(t, row.leave_type.code, row.leave_type.name),
    },
    {
      title: t('leave.range'),
      render: (_, row) => `${formatDateTime(row.start_time, 'YYYY-MM-DD HH:mm')} ~ ${formatDateTime(row.end_time, 'YYYY-MM-DD HH:mm')}`,
    },
    { title: t('leave.days'), dataIndex: 'duration_days' },
    { title: t('leave.reason'), dataIndex: 'reason', ellipsis: true },
    {
      title: t('leave.statusLabel'),
      dataIndex: 'status',
      render: (v: LeaveStatus) => <Tag color={statusMap[v]?.color}>{statusMap[v]?.text}</Tag>,
    },
  ];

  return (
    <div>
      <div className="leave-hero">
        <div>
          <h2>{t('leave.myTitle')}</h2>
          <p>{t('leave.myHint')}</p>
        </div>
        <Button type="primary" icon={<PlusOutlined />} onClick={() => setOpen(true)}>
          {t('leave.apply')}
        </Button>
      </div>
      <div className="leave-balance-grid">
        {balances.map((row) => (
          <Card key={row.id} size="small" className="leave-balance-card" title={leaveTypeLabel(t, row.leave_type.code, row.leave_type.name)}>
            <div className="days">{row.remaining_days}</div>
            <div>{t('leave.balanceHint', { used: row.used_days, total: row.total_days })}</div>
          </Card>
        ))}
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
      <FormModal
        open={open}
        types={types}
        onClose={() => setOpen(false)}
        onSubmit={async (values) => {
          await submitLeave(values);
          message.success(t('leave.submitted'));
          void load();
        }}
      />
    </div>
  );
};

export default MyPage;
