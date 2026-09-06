import React from 'react';
import { Card, Col, List, Progress, Row, Skeleton, Tag, Typography } from 'antd';
import { BarChart, LineChart, PieChart } from '@/components/Chart';
import { formatDateTime } from '@/utils/format';
import type { OverviewMeeting, StatsResult } from '@/api/stats';
import { findSeries, seriesToPie, seriesToXY } from './useDashboardStats';

interface DashboardPanelsProps {
  loading: boolean;
  canReadCharts: boolean;
  member?: StatsResult;
  interview?: StatsResult;
  meeting?: StatsResult;
  task?: StatsResult;
  internship?: StatsResult;
  meetings: OverviewMeeting[];
  taskTodo: number;
  taskOverdue: number;
}

const DashboardPanels: React.FC<DashboardPanelsProps> = ({
  loading, canReadCharts, member, interview, meeting, task, internship,
  meetings, taskTodo, taskOverdue,
}) => {
  const deptPie = seriesToPie(findSeries(member, '部门分布'));
  const iv = seriesToXY(findSeries(interview, '各部门面试人数'));
  const attend = seriesToXY(findSeries(meeting, '出席率'));
  const rank = seriesToXY(findSeries(internship, '时长排行'));
  const statusPie = seriesToPie(findSeries(task, '任务状态'));
  const taskTotal = statusPie.reduce((sum, d) => sum + d.value, 0);
  const taskDone = statusPie.find((d) => d.name === '已完成')?.value ?? 0;
  const onTime = Number((((task?.summary.on_time_rate || 0) * 100)).toFixed(1));
  const completion = taskTotal ? Number(((taskDone / taskTotal) * 100).toFixed(1)) : 0;

  return (
    <>
      <Row gutter={[16, 16]} style={{ marginBottom: 16 }}>
        <Col xs={24} lg={8}>
          {canReadCharts ? (
            <PieChart title="会员分布" data={deptPie} loading={loading} height={280} />
          ) : (
            <Card title="会员分布"><Typography.Text type="secondary">需要 stats:read 查看图表</Typography.Text></Card>
          )}
        </Col>
        <Col xs={24} lg={8}>
          {canReadCharts ? (
            <BarChart
              title="面试数据"
              categories={iv.categories}
              series={[{ name: '面试人数', data: iv.values }]}
              loading={loading}
              height={280}
            />
          ) : (
            <Card title="面试数据"><Typography.Text type="secondary">需要 stats:read 查看图表</Typography.Text></Card>
          )}
        </Col>
        <Col xs={24} lg={8}>
          {canReadCharts ? (
            <LineChart
              title="会议出席率"
              categories={attend.categories}
              series={[{ name: '出席率', data: attend.values }]}
              area
              loading={loading}
              height={280}
            />
          ) : (
            <Card title="会议出席率"><Typography.Text type="secondary">需要 stats:read 查看图表</Typography.Text></Card>
          )}
        </Col>
      </Row>

      <Row gutter={[16, 16]}>
        <Col xs={24} lg={8}>
          <Card title="任务进度">
            {loading ? <Skeleton active paragraph={{ rows: 4 }} /> : (
              <>
                <div style={{ marginBottom: 12 }}>
                  <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                    <span>完成进度</span><span>{completion}%</span>
                  </div>
                  <Progress percent={completion} size="small" />
                </div>
                <div style={{ marginBottom: 12 }}>
                  <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                    <span>按时完成率</span><span>{onTime}%</span>
                  </div>
                  <Progress percent={onTime} size="small" status={onTime < 60 ? 'exception' : 'active'} />
                </div>
                <Typography.Text type="secondary">待办 {taskTodo} · 逾期 {taskOverdue}</Typography.Text>
              </>
            )}
          </Card>
        </Col>
        <Col xs={24} lg={8}>
          {canReadCharts ? (
            <BarChart
              title="实习时长排行"
              categories={rank.categories}
              series={[{ name: '时长', data: rank.values }]}
              horizontal
              loading={loading}
              height={280}
            />
          ) : (
            <Card title="实习时长排行"><Typography.Text type="secondary">需要 stats:read 查看图表</Typography.Text></Card>
          )}
        </Col>
        <Col xs={24} lg={8}>
          <Card title="近期活动">
            <List
              loading={loading}
              dataSource={meetings}
              locale={{ emptyText: '今日暂无会议' }}
              renderItem={(item) => (
                <List.Item>
                  <List.Item.Meta
                    title={item.title}
                    description={formatDateTime(item.start_time, 'YYYY-MM-DD HH:mm')}
                  />
                  <Tag color="blue">会议</Tag>
                </List.Item>
              )}
            />
          </Card>
        </Col>
      </Row>
    </>
  );
};

export default DashboardPanels;
