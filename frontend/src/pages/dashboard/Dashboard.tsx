import React from 'react';
import { Button, Card, Col, Row, Skeleton, Space, Statistic } from 'antd';
import {
  CheckCircleOutlined,
  ExpandOutlined,
  ScheduleOutlined,
  TeamOutlined,
  UserOutlined,
} from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import DashboardPanels from './DashboardPanels';
import { useDashboardStats } from './useDashboardStats';
import './dashboard.css';

const Dashboard: React.FC = () => {
  const navigate = useNavigate();
  const { overview, charts, loading, canReadCharts } = useDashboardStats();

  return (
    <div className="dash-page">
      <div className="dash-hero">
        <div>
          <h2>协会工作台</h2>
          <p>会员、面试、会议、任务与实习数据来自统计接口，大屏适合投屏展示。</p>
        </div>
        <Space wrap>
          <Button size="large" onClick={() => navigate('/member/application')}>入会申请</Button>
          <Button size="large" onClick={() => navigate('/stats/overview')}>统计概览</Button>
          {canReadCharts ? (
            <Button type="primary" size="large" icon={<ExpandOutlined />} onClick={() => navigate('/dashboard/bigscreen')}>
              数据大屏
            </Button>
          ) : null}
        </Space>
      </div>

      <Row gutter={[16, 16]} className="dash-kpis">
        <Col xs={12} lg={6}>
          <Card>
            {loading && !overview ? <Skeleton active paragraph={false} /> : (
              <div className="dash-kpi">
                <Statistic title="会员总数" value={overview?.total_members ?? 0} />
                <span className="dash-kpi-icon" style={{ background: '#e6f4ff', color: '#2563eb' }}><UserOutlined /></span>
              </div>
            )}
          </Card>
        </Col>
        <Col xs={12} lg={6}>
          <Card>
            {loading && !overview ? <Skeleton active paragraph={false} /> : (
              <div className="dash-kpi">
                <Statistic title="待审批" value={overview?.pending_approvals ?? 0} />
                <span className="dash-kpi-icon" style={{ background: '#fffbe6', color: '#faad14' }}><ScheduleOutlined /></span>
              </div>
            )}
          </Card>
        </Col>
        <Col xs={12} lg={6}>
          <Card>
            {loading && !overview ? <Skeleton active paragraph={false} /> : (
              <div className="dash-kpi">
                <Statistic title="进行中任务" value={overview?.total_tasks_in_progress ?? 0} />
                <span className="dash-kpi-icon" style={{ background: '#f6ffed', color: '#52c41a' }}><CheckCircleOutlined /></span>
              </div>
            )}
          </Card>
        </Col>
        <Col xs={12} lg={6}>
          <Card>
            {loading && !overview ? <Skeleton active paragraph={false} /> : (
              <div className="dash-kpi">
                <Statistic title="本月会议" value={overview?.total_meetings_this_month ?? 0} />
                <span className="dash-kpi-icon" style={{ background: '#f9f0ff', color: '#722ed1' }}><TeamOutlined /></span>
              </div>
            )}
          </Card>
        </Col>
      </Row>

      <DashboardPanels
        loading={loading}
        canReadCharts={canReadCharts}
        member={charts['member-distribution']}
        interview={charts['interview-data']}
        meeting={charts['meeting-attendance']}
        task={charts['task-progress']}
        internship={charts['internship-duration']}
        meetings={overview?.today_meetings ?? []}
        taskTodo={overview?.my_tasks.todo ?? 0}
        taskOverdue={overview?.my_tasks.overdue ?? 0}
      />
    </div>
  );
};

export default Dashboard;
