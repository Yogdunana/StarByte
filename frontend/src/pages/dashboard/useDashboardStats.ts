import { useCallback, useEffect, useState } from 'react';
import {
  getStats,
  getStatsOverview,
  type OverviewResponse,
  type StatsProviderCode,
  type StatsResult,
  type StatsSeries,
} from '@/api/stats';
import { usePermission } from '@/hooks/usePermission';

export const DASHBOARD_PROVIDERS: StatsProviderCode[] = [
  'member-distribution',
  'interview-data',
  'meeting-attendance',
  'task-progress',
  'internship-duration',
];

export function findSeries(result: StatsResult | undefined, name: string): StatsSeries | undefined {
  return result?.series.find((s) => s.name === name);
}

export function seriesToXY(series: StatsSeries | undefined): { categories: string[]; values: number[] } {
  if (!series || !series.data.length) {
    return { categories: [], values: [] };
  }
  const categories = series.x_axis?.length ? series.x_axis : series.data.map((d) => d.label);
  return { categories, values: series.data.map((d) => d.value) };
}

export function seriesToPie(series: StatsSeries | undefined): { name: string; value: number }[] {
  if (!series) return [];
  return series.data.map((d) => ({ name: d.label, value: d.value }));
}

export interface DashboardStatsState {
  overview: OverviewResponse | null;
  charts: Partial<Record<StatsProviderCode, StatsResult>>;
  errors: Partial<Record<StatsProviderCode, string>>;
  loading: boolean;
  updatedAt: Date | null;
  canReadCharts: boolean;
  reload: () => Promise<void>;
}

export function useDashboardStats(refreshMs?: number): DashboardStatsState {
  const canReadCharts = usePermission('stats:read');
  const [overview, setOverview] = useState<OverviewResponse | null>(null);
  const [charts, setCharts] = useState<Partial<Record<StatsProviderCode, StatsResult>>>({});
  const [errors, setErrors] = useState<Partial<Record<StatsProviderCode, string>>>({});
  const [loading, setLoading] = useState(true);
  const [updatedAt, setUpdatedAt] = useState<Date | null>(null);

  const reload = useCallback(async () => {
    setLoading(true);
    try {
      setOverview(await getStatsOverview());
    } catch {
      setOverview(null);
    }
    if (!canReadCharts) {
      setCharts({});
      setErrors({});
      setUpdatedAt(new Date());
      setLoading(false);
      return;
    }
    const next: Partial<Record<StatsProviderCode, StatsResult>> = {};
    const errs: Partial<Record<StatsProviderCode, string>> = {};
    for (const code of DASHBOARD_PROVIDERS) {
      try {
        next[code] = await getStats(code);
      } catch (e) {
        errs[code] = e instanceof Error ? e.message : '加载失败';
      }
    }
    setCharts(next);
    setErrors(errs);
    setUpdatedAt(new Date());
    setLoading(false);
  }, [canReadCharts]);

  useEffect(() => {
    void reload();
  }, [reload]);

  useEffect(() => {
    if (!refreshMs || refreshMs <= 0) return undefined;
    const id = window.setInterval(() => {
      void reload();
    }, refreshMs);
    return () => window.clearInterval(id);
  }, [reload, refreshMs]);

  return { overview, charts, errors, loading, updatedAt, canReadCharts, reload };
}
