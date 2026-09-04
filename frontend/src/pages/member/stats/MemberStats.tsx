import React, { useEffect, useState } from 'react';
import { Card, Col, Empty, Row, Spin } from 'antd';
import ReactECharts from 'echarts-for-react';
import { getApplicationStats, getMemberStats } from '@/api/member';
import type { MemberStatItem } from '@/types/api';

const MemberStats: React.FC = () => {
  const [loading, setLoading] = useState(true);
  const [trend, setTrend] = useState<MemberStatItem[]>([]);
  const [dept, setDept] = useState<MemberStatItem[]>([]);

  useEffect(() => {
    Promise.all([
      getApplicationStats({ group_by: 'date' }),
      getMemberStats({ group_by: 'department' }),
    ])
      .then(([a, m]) => {
        setTrend(a.items || []);
        setDept(m.items || []);
      })
      .finally(() => setLoading(false));
  }, []);

  if (loading) {
    return <Spin />;
  }

  return (
    <Row gutter={16}>
      <Col span={14}>
        <Card title="申请趋势" size="small">
          {trend.length === 0 ? (
            <Empty description="暂无申请数据" />
          ) : (
            <ReactECharts
              style={{ height: 280 }}
              option={{
                tooltip: { trigger: 'axis' },
                xAxis: { type: 'category', data: trend.map((i) => i.label) },
                yAxis: { type: 'value', minInterval: 1 },
                series: [{ type: 'line', data: trend.map((i) => i.count), smooth: true }],
              }}
            />
          )}
        </Card>
      </Col>
      <Col span={10}>
        <Card title="会员部门分布" size="small">
          {dept.length === 0 ? (
            <Empty description="暂无档案数据" />
          ) : (
            <ReactECharts
              style={{ height: 280 }}
              option={{
                tooltip: { trigger: 'item' },
                series: [
                  {
                    type: 'pie',
                    radius: '70%',
                    data: dept.map((i) => ({ name: i.label, value: i.count })),
                  },
                ],
              }}
            />
          )}
        </Card>
      </Col>
    </Row>
  );
};

export default MemberStats;
