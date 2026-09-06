import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { Card, Col, Row, Select, Space, Statistic, Typography, message } from 'antd';
import {
  exportStats,
  getStats,
  getStatsOverview,
  type OverviewResponse,
  type StatsProviderCode,
  type StatsQuery,
  type StatsResult,
} from '@/api/stats';
import { getMemberDepartments } from '@/api/member';
import type { MemberDepartmentOption } from '@/types/api';
import { usePermission } from '@/hooks/usePermission';
import TimeRangeSelect, { resolveTimeRange, type TimeRangeValue } from './TimeRangeSelect';
import ProviderCharts from './ProviderCharts';
import './stats.css';

const PROVIDERS: Array<{ code: StatsProviderCode; title: string }> = [
  { code: 'member-distribution', title: '会员分布' },
  { code: 'interview-data', title: '面试数据统计' },
  { code: 'meeting-attendance', title: '会议出席统计' },
  { code: 'task-progress', title: '任务进度统计' },
  { code: 'internship-duration', title: '实习时长统计' },
];

const OverviewPage: React.FC = () => {
  const canExport = usePermission('stats:export');
  const [range, setRange] = useState<TimeRangeValue>({ preset: 'month' });
  const [departmentId, setDepartmentId] = useState<string>();
  const [groupBy, setGroupBy] = useState<string | undefined>();
  const [granularity, setGranularity] = useState('month');
  const [departments, setDepartments] = useState<MemberDepartmentOption[]>([]);
  const [overview, setOverview] = useState<OverviewResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [results, setResults] = useState<Partial<Record<StatsProviderCode, StatsResult>>>({});
  const [errors, setErrors] = useState<Partial<Record<StatsProviderCode, string>>>({});

  const query = useMemo<StatsQuery>(() => {
    const dates = resolveTimeRange(range);
    return {
      ...dates,
      department_id: departmentId,
      group_by: groupBy,
      granularity,
    };
  }, [range, departmentId, groupBy, granularity]);

  useEffect(() => {
    void getMemberDepartments().then(setDepartments).catch(() => undefined);
  }, []);

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const ov = await getStatsOverview();
      setOverview(ov);
    } catch {
      setOverview(null);
    }
    const next: Partial<Record<StatsProviderCode, StatsResult>> = {};
    const errs: Partial<Record<StatsProviderCode, string>> = {};
    await Promise.all(
      PROVIDERS.map(async (p) => {
        try {
          next[p.code] = await getStats(p.code, query);
        } catch (e) {
          errs[p.code] = e instanceof Error ? e.message : '加载失败';
        }
      }),
    );
    setResults(next);
    setErrors(errs);
    setLoading(false);
  }, [query]);

  useEffect(() => {
    void load();
  }, [load]);

  const handleExport = async (code: StatsProviderCode, format: 'csv' | 'excel') => {
    try {
      await exportStats(code, format, query);
    } catch {
      message.error('导出失败');
    }
  };

  return (
    <div className="stats-page">
      <div className="stats-hero">
        <div>
          <Typography.Title level={3} style={{ color: '#fff', margin: 0 }}>统计概览</Typography.Title>
          <Typography.Paragraph style={{ color: 'rgba(255,255,255,.85)', margin: '6px 0 0' }}>
            会员 / 面试 / 会议 / 任务 / 实习 一期五组图表
          </Typography.Paragraph>
        </div>
        <Space wrap size={12}>
          <TimeRangeSelect value={range} onChange={setRange} />
          <Select
            allowClear
            placeholder="部门"
            style={{ minWidth: 160 }}
            value={departmentId}
            onChange={setDepartmentId}
            options={departments.map((d) => ({ value: d.id, label: d.name }))}
          />
          <Select
            value={granularity}
            style={{ width: 110 }}
            onChange={setGranularity}
            options={[
              { value: 'day', label: '按日' },
              { value: 'week', label: '按周' },
              { value: 'month', label: '按月' },
            ]}
          />
          <Select
            allowClear
            placeholder="分组"
            style={{ width: 120 }}
            value={groupBy}
            onChange={setGroupBy}
            options={[
              { value: 'department', label: '部门' },
              { value: 'date', label: '日期' },
              { value: 'status', label: '状态' },
              { value: 'grade', label: '年级' },
              { value: 'type', label: '类型' },
            ]}
          />
        </Space>
      </div>

      <Row gutter={[16, 16]} className="stats-kpis">
        <Col xs={12} md={6}><Card><Statistic title="会员总数" value={overview?.total_members ?? 0} /></Card></Col>
        <Col xs={12} md={6}><Card><Statistic title="本月会议" value={overview?.total_meetings_this_month ?? 0} /></Card></Col>
        <Col xs={12} md={6}><Card><Statistic title="进行中任务" value={overview?.total_tasks_in_progress ?? 0} /></Card></Col>
        <Col xs={12} md={6}><Card><Statistic title="活跃实习" value={overview?.total_internships_active ?? 0} /></Card></Col>
      </Row>

      {PROVIDERS.map((p) => (
        <Card key={p.code} className="stats-section" title={p.title}>
          <ProviderCharts
            result={results[p.code]}
            loading={loading}
            error={errors[p.code]}
            onExport={canExport ? (format) => { void handleExport(p.code, format); } : undefined}
          />
        </Card>
      ))}
    </div>
  );
};

export default OverviewPage;
