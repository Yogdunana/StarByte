import React, { useEffect, useState } from 'react';
import { Card, Col, Empty, Row, Spin, Statistic } from 'antd';
import ReactECharts from 'echarts-for-react';
import { getInterviewStats } from '@/api/interview';
import type { InterviewStats } from '@/types/api';

const StatsPage: React.FC = () => {
  const [loading, setLoading] = useState(true);
  const [stats, setStats] = useState<InterviewStats | null>(null);

  useEffect(() => {
    getInterviewStats()
      .then(setStats)
      .finally(() => setLoading(false));
  }, []);

  if (loading) return <Spin />;
  if (!stats) return <Empty />;

  return (
    <Card title="面试统计">
      <Row gutter={16} style={{ marginBottom: 16 }}>
        <Col span={6}><Statistic title="总场次/记录" value={stats.total} /></Col>
        <Col span={6}><Statistic title="通过" value={stats.pass_count} /></Col>
        <Col span={6}><Statistic title="不通过" value={stats.fail_count} /></Col>
        <Col span={6}><Statistic title="通过率" value={stats.pass_rate} suffix="%" /></Col>
      </Row>
      <Row gutter={16}>
        <Col span={12}>
          <Card size="small" title="评分分布">
            {stats.score_buckets.every((b) => b.count === 0) ? (
              <Empty />
            ) : (
              <ReactECharts
                style={{ height: 300 }}
                option={{
                  tooltip: { trigger: 'axis' },
                  xAxis: { type: 'category', data: stats.score_buckets.map((b) => b.range) },
                  yAxis: { type: 'value', minInterval: 1 },
                  series: [{ type: 'bar', data: stats.score_buckets.map((b) => b.count) }],
                }}
              />
            )}
          </Card>
        </Col>
        <Col span={12}>
          <Card size="small" title="各部门面试人数">
            {stats.by_department.length === 0 ? (
              <Empty />
            ) : (
              <ReactECharts
                style={{ height: 300 }}
                option={{
                  tooltip: { trigger: 'axis' },
                  xAxis: { type: 'category', data: stats.by_department.map((d) => d.department) },
                  yAxis: { type: 'value', minInterval: 1 },
                  series: [
                    { name: '人数', type: 'bar', data: stats.by_department.map((d) => d.count) },
                    { name: '通过', type: 'bar', data: stats.by_department.map((d) => d.pass_count) },
                  ],
                }}
              />
            )}
          </Card>
        </Col>
      </Row>
    </Card>
  );
};

export default StatsPage;
