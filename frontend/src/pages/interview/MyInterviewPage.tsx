import React, { useCallback, useEffect, useState } from 'react';
import { Card, Table, Tabs } from 'antd';
import StatusTag from '@/components/StatusTag/StatusTag';
import { getMyInterviews } from '@/api/interview';
import type { Interview, InterviewRecordStatus } from '@/types/api';
import { InterviewStatusMap, ResultMap } from './meta';

const groups: Array<{ key: string; label: string; status?: InterviewRecordStatus }> = [
  { key: 'pending', label: '待面试', status: 0 },
  { key: 'ongoing', label: '进行中', status: 2 },
  { key: 'done', label: '已完成', status: 3 },
  { key: 'all', label: '全部' },
];

const MyInterviewPage: React.FC = () => {
  const [tab, setTab] = useState('pending');
  const [list, setList] = useState<Interview[]>([]);
  const [loading, setLoading] = useState(false);

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const g = groups.find((x) => x.key === tab);
      setList(await getMyInterviews(g?.status));
    } finally {
      setLoading(false);
    }
  }, [tab]);

  useEffect(() => { void load(); }, [load]);

  return (
    <Card title="我的面试">
      <Tabs
        activeKey={tab}
        onChange={setTab}
        items={groups.map((g) => ({
          key: g.key,
          label: g.label,
          children: (
            <Table
              rowKey="id"
              loading={loading}
              dataSource={list}
              columns={[
                { title: '面试者', render: (_, r) => r.applicant.name },
                { title: '场次', dataIndex: 'session_title' },
                { title: '地点', dataIndex: 'location' },
                { title: '预约时间', dataIndex: 'scheduled_time' },
                { title: '面试官', render: (_, r) => r.evaluators.map((e) => e.name).join('、') || '-' },
                { title: '状态', dataIndex: 'status', render: (v: number) => <StatusTag status={v} mapping={InterviewStatusMap} /> },
                { title: '结果', dataIndex: 'result', render: (v: number) => <StatusTag status={v} mapping={ResultMap} /> },
                { title: '得分', dataIndex: 'score', render: (v?: number) => (v == null ? '-' : v) },
              ]}
            />
          ),
        }))}
      />
    </Card>
  );
};

export default MyInterviewPage;
