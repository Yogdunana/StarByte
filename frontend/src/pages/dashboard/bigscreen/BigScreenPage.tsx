import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { Button, Space, Typography } from 'antd';
import { CompressOutlined, ExpandOutlined, ReloadOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { BarChart, LineChart, PieChart } from '@/components/Chart';
import { formatDateTime } from '@/utils/format';
import { findSeries, seriesToPie, seriesToXY, useDashboardStats } from '../useDashboardStats';
import './bigscreen.css';

const REFRESH_MS = 5 * 60 * 1000;

function BigScreenClock() {
  const [now, setNow] = useState(() => new Date());
  useEffect(() => {
    const id = window.setInterval(() => setNow(new Date()), 1000);
    return () => window.clearInterval(id);
  }, []);
  return <span>{formatDateTime(now)}</span>;
}

async function leaveFullscreen() {
  if (document.fullscreenElement) {
    await document.exitFullscreen();
  }
}

const BigScreenPage: React.FC = () => {
  const navigate = useNavigate();
  const { overview, charts, loading, updatedAt, reload } = useDashboardStats(REFRESH_MS);
  const [fullscreen, setFullscreen] = useState(false);

  useEffect(() => {
    const onFs = () => setFullscreen(Boolean(document.fullscreenElement));
    document.addEventListener('fullscreenchange', onFs);
    return () => {
      document.removeEventListener('fullscreenchange', onFs);
      void leaveFullscreen();
    };
  }, []);

  const toggleFullscreen = useCallback(() => {
    if (document.fullscreenElement) {
      void document.exitFullscreen();
      return;
    }
    void document.documentElement.requestFullscreen();
  }, []);

  const goDashboard = useCallback(() => {
    void leaveFullscreen().finally(() => navigate('/dashboard'));
  }, [navigate]);

  const member = useMemo(() => seriesToPie(findSeries(charts['member-distribution'], '部门分布')), [charts]);
  const interview = useMemo(() => seriesToXY(findSeries(charts['interview-data'], '各部门面试人数')), [charts]);
  const meeting = useMemo(() => seriesToXY(findSeries(charts['meeting-attendance'], '出席率')), [charts]);
  const rank = useMemo(() => seriesToXY(findSeries(charts['internship-duration'], '时长排行')), [charts]);
  const taskPie = useMemo(() => seriesToPie(findSeries(charts['task-progress'], '任务状态')), [charts]);
  const interviewSeries = useMemo(() => [{ name: '面试人数', data: interview.values }], [interview]);
  const meetingSeries = useMemo(() => [{ name: '出席率', data: meeting.values }], [meeting]);
  const rankSeries = useMemo(() => [{ name: '时长', data: rank.values }], [rank]);

  return (
    <div className="bs-root">
      <header className="bs-header">
        <div>
          <Typography.Title level={3} className="bs-title">StarByte 数据大屏</Typography.Title>
          <p className="bs-sub">协会核心数据 · 每 5 分钟自动刷新 · 支持 F11 / 按钮全屏</p>
        </div>
        <Space wrap size={16} className="bs-meta">
          <BigScreenClock />
          <span>上次刷新 {updatedAt ? formatDateTime(updatedAt, 'HH:mm:ss') : '--'}</span>
          <Button ghost icon={<ReloadOutlined />} onClick={() => { void reload(); }}>刷新</Button>
          <Button ghost icon={fullscreen ? <CompressOutlined /> : <ExpandOutlined />} onClick={toggleFullscreen}>
            {fullscreen ? '退出全屏' : '全屏'}
          </Button>
          <Button ghost onClick={goDashboard}>返回工作台</Button>
        </Space>
      </header>

      <section className="bs-kpis">
        <div className="bs-kpi"><span>会员</span><strong>{overview?.total_members ?? 0}</strong></div>
        <div className="bs-kpi"><span>待审批</span><strong>{overview?.pending_approvals ?? 0}</strong></div>
        <div className="bs-kpi"><span>进行中任务</span><strong>{overview?.total_tasks_in_progress ?? 0}</strong></div>
        <div className="bs-kpi"><span>本月会议</span><strong>{overview?.total_meetings_this_month ?? 0}</strong></div>
        <div className="bs-kpi"><span>活跃实习</span><strong>{overview?.total_internships_active ?? 0}</strong></div>
        <div className="bs-kpi"><span>未读通知</span><strong>{overview?.notifications_unread ?? 0}</strong></div>
      </section>

      <section className="bs-grid">
        <div className="bs-panel">
          <h4>会员分布</h4>
          <PieChart data={member} loading={loading} height="100%" theme="dark" />
        </div>
        <div className="bs-panel">
          <h4>面试数据</h4>
          <BarChart
            categories={interview.categories}
            series={interviewSeries}
            loading={loading}
            height="100%"
            theme="dark"
          />
        </div>
        <div className="bs-panel">
          <h4>会议出席率</h4>
          <LineChart
            categories={meeting.categories}
            series={meetingSeries}
            area
            loading={loading}
            height="100%"
            theme="dark"
          />
        </div>
        <div className="bs-panel">
          <h4>任务进度</h4>
          <PieChart data={taskPie} loading={loading} height="100%" theme="dark" />
        </div>
        <div className="bs-panel bs-panel-wide">
          <h4>实习时长排行</h4>
          <BarChart
            categories={rank.categories}
            series={rankSeries}
            horizontal
            loading={loading}
            height="100%"
            theme="dark"
          />
        </div>
      </section>
    </div>
  );
};

export default BigScreenPage;
