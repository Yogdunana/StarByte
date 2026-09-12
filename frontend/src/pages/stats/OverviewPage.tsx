import { tx, useLocale } from '@/i18n/text';
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
  {
    code: 'member-distribution',
    get title() {
      return tx('会员分布');
    },
  },
  {
    code: 'interview-data',
    get title() {
      return tx('面试数据统计');
    },
  },
  {
    code: 'meeting-attendance',
    get title() {
      return tx('会议出席统计');
    },
  },
  {
    code: 'task-progress',
    get title() {
      return tx('任务进度统计');
    },
  },
  {
    code: 'internship-duration',
    get title() {
      return tx('实习时长统计');
    },
  },
];

const OverviewPage: React.FC = () => {
  const uiLanguage = useLocale();
  const canExport = usePermission('stats:export');
  const [range, setRange] = useState<TimeRangeValue>({ preset: 'month' });
  const [departmentId, setDepartmentId] = useState<string>();
  const [granularity, setGranularity] = useState('month');
  const [departments, setDepartments] = useState<MemberDepartmentOption[]>([]);
  const [overview, setOverview] = useState<OverviewResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [results, setResults] = useState<Partial<Record<StatsProviderCode, StatsResult>>>({});
  const [errors, setErrors] = useState<Partial<Record<StatsProviderCode, string>>>({});

  const query = useMemo<StatsQuery>(() => {
    // Invalidate cached labels when the selected language changes.
    void uiLanguage;
    const dates = resolveTimeRange(range);
    return {
      ...dates,
      department_id: departmentId,
      granularity,
    };
  }, [range, departmentId, granularity, uiLanguage]);

  useEffect(() => {
    void getMemberDepartments()
      .then(setDepartments)
      .catch(() => undefined);
  }, []);

  const load = useCallback(async () => {
    setLoading(true);
    const next: Partial<Record<StatsProviderCode, StatsResult>> = {};
    const errs: Partial<Record<StatsProviderCode, string>> = {};
    try {
      const ov = await getStatsOverview();
      setOverview(ov);
    } catch {
      setOverview(null);
    }
    for (const p of PROVIDERS) {
      try {
        next[p.code] = await getStats(p.code, query);
      } catch (e) {
        errs[p.code] = e instanceof Error ? e.message : tx('加载失败');
      }
    }
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
      message.error(tx('导出失败'));
    }
  };

  return (
    <div className="stats-page">
      <div className="stats-hero">
        <div>
          <Typography.Title level={3} style={{ color: '#fff', margin: 0 }}>
            {tx('统计概览')}
          </Typography.Title>
          <Typography.Paragraph style={{ color: 'rgba(255,255,255,.85)', margin: '6px 0 0' }}>
            {tx('会员 / 面试 / 会议 / 任务 / 实习 一期五组图表')}
          </Typography.Paragraph>
        </div>
        <Space wrap size={12}>
          <TimeRangeSelect value={range} onChange={setRange} />
          <Select
            allowClear
            placeholder={tx('部门')}
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
              { value: 'day', label: tx('按日') },
              { value: 'week', label: tx('按周') },
              { value: 'month', label: tx('按月') },
            ]}
          />
        </Space>
      </div>

      <Row gutter={[16, 16]} className="stats-kpis">
        <Col xs={12} md={6}>
          <Card>
            <Statistic title={tx('会员总数')} value={overview?.total_members ?? 0} />
          </Card>
        </Col>
        <Col xs={12} md={6}>
          <Card>
            <Statistic title={tx('本月会议')} value={overview?.total_meetings_this_month ?? 0} />
          </Card>
        </Col>
        <Col xs={12} md={6}>
          <Card>
            <Statistic title={tx('进行中任务')} value={overview?.total_tasks_in_progress ?? 0} />
          </Card>
        </Col>
        <Col xs={12} md={6}>
          <Card>
            <Statistic title={tx('活跃实习')} value={overview?.total_internships_active ?? 0} />
          </Card>
        </Col>
      </Row>

      {PROVIDERS.map((p) => (
        <Card key={p.code} className="stats-section" title={p.title}>
          <ProviderCharts
            result={results[p.code]}
            loading={loading}
            error={errors[p.code]}
            onExport={
              canExport
                ? (format) => {
                    void handleExport(p.code, format);
                  }
                : undefined
            }
          />
        </Card>
      ))}
    </div>
  );
};

export default OverviewPage;
