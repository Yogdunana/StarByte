import request from './request';
import type { AxiosResponse } from 'axios';
import { downloadBlob } from '@/utils/download';

export interface StatsQuery {
  start_date?: string;
  end_date?: string;
  department_id?: string;
  group_by?: string;
  granularity?: string;
}

export interface StatsDataPoint {
  label: string;
  value: number;
}

export interface StatsSeries {
  name: string;
  type: string;
  data: StatsDataPoint[];
  x_axis: string[];
}

export interface StatsChartConfig {
  chart_type: string;
  title: string;
  stack?: boolean;
  horizontal?: boolean;
}

export interface StatsResult {
  provider: string;
  summary: Record<string, number>;
  series: StatsSeries[];
  chart_config: StatsChartConfig | null;
}

export interface ProviderInfo {
  name: string;
  display_name: string;
  chart_config: StatsChartConfig | null;
}

export interface OverviewMeeting {
  id: string;
  title: string;
  start_time: string;
}

export interface OverviewResponse {
  total_members: number;
  total_meetings_this_month: number;
  total_tasks_in_progress: number;
  total_internships_active: number;
  pending_approvals: number;
  today_meetings: OverviewMeeting[];
  my_tasks: { todo: number; overdue: number };
  notifications_unread: number;
}

export type StatsProviderCode =
  | 'member-distribution'
  | 'interview-data'
  | 'meeting-attendance'
  | 'task-progress'
  | 'internship-duration';

export function listStatsProviders(): Promise<ProviderInfo[]> {
  return request.get('/stats/providers');
}

export function getStatsOverview(): Promise<OverviewResponse> {
  return request.get('/stats/overview');
}

export function getStats(provider: string, params?: StatsQuery): Promise<StatsResult> {
  return request.get(`/stats/${provider}`, { params });
}

function filenameFromDisposition(header: string | undefined, fallback: string): string {
  if (!header) return fallback;
  const match = /filename\*=UTF-8''([^;]+)|filename="?([^"]+)"?/i.exec(header);
  const raw = match?.[1] || match?.[2];
  return raw ? decodeURIComponent(raw) : fallback;
}

export async function exportStats(
  provider: string,
  format: 'csv' | 'excel',
  params?: StatsQuery,
): Promise<void> {
  const response: AxiosResponse = await request.get(`/stats/export/${provider}`, {
    params: { ...params, format },
    responseType: 'blob',
  });
  const fallback = format === 'excel' ? `stats_${provider}.xlsx` : `stats_${provider}.csv`;
  downloadBlob(
    response.data as Blob,
    filenameFromDisposition(response.headers['content-disposition'] as string | undefined, fallback),
  );
}
