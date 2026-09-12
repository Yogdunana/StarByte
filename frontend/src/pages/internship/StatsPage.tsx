import { tx, useLocale } from '@/i18n/text';
import React, { useEffect, useState } from 'react';
import { Card, Col, Empty, Row, Select, Space, Table } from 'antd';
import ReactECharts from 'echarts-for-react';
import {
  getInternshipDepartmentStats,
  getInternshipDurationStats,
  getInternshipRanking,
} from '@/api/internship';
import type {
  InternshipDepartmentStat,
  InternshipDurationItem,
  InternshipRanking,
} from '@/types/api';
import './internship.css';

const StatsPage: React.FC = () => {
  useLocale();
  const [groupBy, setGroupBy] = useState<'user' | 'department' | 'month'>('month');
  const [sortBy, setSortBy] = useState<'duration' | 'count'>('duration');
  const [duration, setDuration] = useState<InternshipDurationItem[]>([]);
  const [ranking, setRanking] = useState<InternshipRanking[]>([]);
  const [departments, setDepartments] = useState<InternshipDepartmentStat[]>([]);
  const [rankError, setRankError] = useState('');

  useEffect(() => {
    void getInternshipDurationStats({ group_by: groupBy }).then((res) =>
      setDuration(res.items || []),
    );
  }, [groupBy]);

  useEffect(() => {
    void getInternshipRanking({ sort_by: sortBy, limit: 15 })
      .then((res) => {
        setRanking(res.rankings || []);
        setRankError('');
      })
      .catch((err: unknown) => {
        setRanking([]);
        setRankError(err instanceof Error ? err.message : tx('排行榜暂不可见'));
      });
  }, [sortBy]);

  useEffect(() => {
    void getInternshipDepartmentStats().then((res) => setDepartments(res.items || []));
  }, []);

  return (
    <div>
      <div className="intern-hero">
        <div>
          <h2>{tx('实习统计与排行')}</h2>
          <p>{tx('按时长、次数和部门查看培养投入，方便招新后的骨干选拔。')}</p>
        </div>
      </div>
      <Row gutter={16} style={{ marginBottom: 16 }}>
        <Col span={14}>
          <Card
            title={tx('时长分布')}
            extra={
              <Select
                value={groupBy}
                style={{ width: 120 }}
                onChange={setGroupBy}
                options={[
                  { value: 'month', label: tx('按月') },
                  { value: 'user', label: tx('按人') },
                  { value: 'department', label: tx('按部门') },
                ]}
              />
            }
          >
            {duration.length === 0 ? (
              <Empty />
            ) : (
              <ReactECharts
                style={{ height: 320 }}
                option={{
                  tooltip: { trigger: 'axis' },
                  xAxis: { type: 'category', data: duration.map((i) => i.name) },
                  yAxis: { type: 'value', name: tx('天') },
                  series: [
                    {
                      type: 'bar',
                      data: duration.map((i) => i.duration_days),
                      itemStyle: { color: '#2563eb', borderRadius: [8, 8, 0, 0] },
                    },
                  ],
                }}
              />
            )}
          </Card>
        </Col>
        <Col span={10}>
          <Card title={tx('部门分布')}>
            {departments.length === 0 ? (
              <Empty />
            ) : (
              <ReactECharts
                style={{ height: 320 }}
                option={{
                  tooltip: { trigger: 'item' },
                  series: [
                    {
                      type: 'pie',
                      radius: ['42%', '68%'],
                      data: departments.map((d) => ({
                        name: d.department?.name || tx('未分配'),
                        value: d.duration_days,
                      })),
                    },
                  ],
                }}
              />
            )}
          </Card>
        </Col>
      </Row>
      <Card
        title={tx('排行榜')}
        extra={
          <Space>
            <Select
              value={sortBy}
              style={{ width: 140 }}
              onChange={setSortBy}
              options={[
                { value: 'duration', label: tx('按时长') },
                { value: 'count', label: tx('按次数') },
              ]}
            />
          </Space>
        }
      >
        {rankError ? (
          <Empty description={rankError} />
        ) : (
          <Table
            rowKey={(r) => `${r.rank}-${r.user.id}`}
            dataSource={ranking}
            pagination={false}
            columns={[
              { title: tx('名次'), dataIndex: 'rank', width: 70 },
              { title: tx('成员'), key: 'user', render: (_, r) => r.user.name },
              { title: tx('部门'), key: 'dept', render: (_, r) => r.department?.name || '-' },
              {
                title: tx('总时长'),
                dataIndex: 'total_duration_days',
                render: (v: number) => tx('{{value0}} 天', { value0: v }),
              },
              { title: tx('次数'), dataIndex: 'internship_count' },
              { title: tx('最近实习'), dataIndex: 'latest_internship' },
            ]}
          />
        )}
      </Card>
    </div>
  );
};

export default StatsPage;
