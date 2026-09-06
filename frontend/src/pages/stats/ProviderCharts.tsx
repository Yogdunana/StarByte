import React from 'react';
import { Col, Row, Statistic } from 'antd';
import ChartCard from '@/components/Chart/ChartCard';
import type { StatsResult, StatsSeries } from '@/api/stats';
import {
  barOption,
  calendarOption,
  gaugeOption,
  lineOption,
  pieOption,
  seriesEmpty,
  stackedBarOption,
} from './chartOptions';

interface ProviderChartsProps {
  result: StatsResult | undefined;
  loading: boolean;
  error?: string;
  onExport?: (format: 'csv' | 'excel') => void;
}

function byName(series: StatsSeries[], name: string): StatsSeries | undefined {
  return series.find((s) => s.name === name);
}

const ProviderCharts: React.FC<ProviderChartsProps> = ({ result, loading, error, onExport }) => {
  const series = result?.series || [];
  const summary = result?.summary || {};

  if (result?.provider === 'member-distribution') {
    const dept = byName(series, '部门分布');
    const grade = byName(series, '年级分布');
    const trend = byName(series, '增长趋势');
    return (
      <Row gutter={[16, 16]}>
        <Col xs={24} md={8}>
          <ChartCard title="部门分布" loading={loading} error={error} empty={seriesEmpty(dept)} option={dept ? pieOption(dept) : undefined} height={280} onExport={onExport} />
        </Col>
        <Col xs={24} md={8}>
          <ChartCard title="年级分布" loading={loading} error={error} empty={seriesEmpty(grade)} option={grade ? barOption(grade) : undefined} height={280} />
        </Col>
        <Col xs={24} md={8}>
          <ChartCard title="增长趋势" loading={loading} error={error} empty={seriesEmpty(trend)} option={trend ? lineOption(trend) : undefined} height={280} />
        </Col>
      </Row>
    );
  }

  if (result?.provider === 'interview-data') {
    const gauge = byName(series, '通过率');
    const dept = byName(series, '各部门面试人数');
    const hist = byName(series, '评分分布');
    const trend = byName(series, '面试趋势');
    return (
      <Row gutter={[16, 16]}>
        <Col xs={24} md={6}>
          <ChartCard title="通过率" loading={loading} error={error} empty={seriesEmpty(gauge)} option={gauge ? gaugeOption(gauge) : undefined} height={280} onExport={onExport} />
        </Col>
        <Col xs={24} md={6}>
          <ChartCard title="各部门面试人数" loading={loading} error={error} empty={seriesEmpty(dept)} option={dept ? barOption(dept) : undefined} height={280} />
        </Col>
        <Col xs={24} md={6}>
          <ChartCard title="评分分布" loading={loading} error={error} empty={seriesEmpty(hist)} option={hist ? barOption(hist) : undefined} height={280} />
        </Col>
        <Col xs={24} md={6}>
          <ChartCard title="面试趋势" loading={loading} error={error} empty={seriesEmpty(trend)} option={trend ? lineOption(trend) : undefined} height={280} />
        </Col>
      </Row>
    );
  }

  if (result?.provider === 'meeting-attendance') {
    const rate = byName(series, '出席率');
    const dept = byName(series, '各部门会议数');
    const cal = byName(series, '会议频率');
    return (
      <Row gutter={[16, 16]}>
        <Col xs={24} md={8}>
          <ChartCard title="出席率" loading={loading} error={error} empty={seriesEmpty(rate)} option={rate ? lineOption(rate) : undefined} height={280} onExport={onExport} />
        </Col>
        <Col xs={24} md={8}>
          <ChartCard title="各部门会议数" loading={loading} error={error} empty={seriesEmpty(dept)} option={dept ? barOption(dept) : undefined} height={280} />
        </Col>
        <Col xs={24} md={8}>
          <ChartCard title="会议频率" loading={loading} error={error} empty={seriesEmpty(cal)} option={cal ? calendarOption(cal) : undefined} height={280} />
        </Col>
      </Row>
    );
  }

  if (result?.provider === 'task-progress') {
    const pie = byName(series, '任务状态');
    const stacked = series.filter((s) => s.name !== '任务状态' && s.type === 'bar');
    return (
      <>
        <Row gutter={16} style={{ marginBottom: 12 }}>
          <Col span={8}><Statistic title="任务总数" value={summary.total_tasks || 0} /></Col>
          <Col span={8}><Statistic title="按时完成率" value={Number(((summary.on_time_rate || 0) * 100).toFixed(1))} suffix="%" /></Col>
          <Col span={8}><Statistic title="进行中" value={summary.in_progress || 0} /></Col>
        </Row>
        <Row gutter={[16, 16]}>
          <Col xs={24} md={10}>
            <ChartCard title="任务状态" loading={loading} error={error} empty={seriesEmpty(pie)} option={pie ? pieOption(pie) : undefined} height={280} onExport={onExport} />
          </Col>
          <Col xs={24} md={14}>
            <ChartCard title="状态趋势" loading={loading} error={error} empty={stacked.every(seriesEmpty)} option={stacked.length ? stackedBarOption(stacked) : undefined} height={280} />
          </Col>
        </Row>
      </>
    );
  }

  if (result?.provider === 'internship-duration') {
    const rank = byName(series, '时长排行');
    const dept = byName(series, '部门平均时长');
    const trend = byName(series, '时长趋势');
    return (
      <Row gutter={[16, 16]}>
        <Col xs={24} md={8}>
          <ChartCard title="时长排行" loading={loading} error={error} empty={seriesEmpty(rank)} option={rank ? barOption(rank, true) : undefined} height={320} onExport={onExport} />
        </Col>
        <Col xs={24} md={8}>
          <ChartCard title="部门平均时长" loading={loading} error={error} empty={seriesEmpty(dept)} option={dept ? barOption(dept) : undefined} height={320} />
        </Col>
        <Col xs={24} md={8}>
          <ChartCard title="时长趋势" loading={loading} error={error} empty={seriesEmpty(trend)} option={trend ? lineOption(trend) : undefined} height={320} />
        </Col>
      </Row>
    );
  }

  return (
    <ChartCard title={result?.chart_config?.title || '统计'} loading={loading} error={error} empty={!series.length} option={series[0] ? barOption(series[0]) : undefined} onExport={onExport} />
  );
};

export default ProviderCharts;
