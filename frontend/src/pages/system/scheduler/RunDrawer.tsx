import { tx, useLocale } from '@/i18n/text';
import React from 'react';
import { Drawer, Empty, Table, Tag } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import type { SchedulerLogs, SchedulerRun } from '@/types/api';

interface Props {
  open: boolean;
  logs: SchedulerLogs | null;
  onClose: () => void;
  onSelectRun: (runId: string) => void;
}

const statusColor: Record<string, string> = {
  success: 'green',
  failed: 'orange',
  retrying: 'gold',
  dead: 'red',
  running: 'blue',
  pending: 'default',
};

const RunDrawer: React.FC<Props> = ({ open, logs, onClose, onSelectRun }) => {
  useLocale();
  const columns: ColumnsType<SchedulerRun> = [
    {
      title: tx('状态'),
      dataIndex: 'status',
      width: 100,
      render: (v: string) => <Tag color={statusColor[v] || 'default'}>{v}</Tag>,
    },
    { title: tx('次数'), dataIndex: 'attempt', width: 70 },
    { title: tx('开始'), dataIndex: 'started_at', width: 180 },
    {
      title: tx('错误'),
      dataIndex: 'error_text',
      ellipsis: true,
    },
  ];

  return (
    <Drawer title={tx('执行记录')} open={open} onClose={onClose} width={640}>
      {!logs ? (
        <Empty />
      ) : (
        <>
          <Table
            rowKey="id"
            size="small"
            columns={columns}
            dataSource={logs.runs || []}
            pagination={false}
            onRow={(row) => ({ onClick: () => onSelectRun(row.id) })}
          />
          <div className="sched-log" style={{ marginTop: 16 }}>
            {(logs.logs || []).map((l) => `[${l.level}] ${l.line}`).join('\n') || tx('暂无日志')}
          </div>
        </>
      )}
    </Drawer>
  );
};

export default RunDrawer;
