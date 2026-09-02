import request from './request';
import type { DashboardStats, ChartDataPoint, PieChartData } from '@/types/api';

// 获取仪表盘统计
export function getDashboardStats(): Promise<DashboardStats> {
  return request.get('/stats/dashboard');
}

// 获取用户增长趋势
export function getUserGrowth(params: { start_date: string; end_date: string }): Promise<ChartDataPoint[]> {
  return request.get('/stats/user-growth', { params });
}

// 获取会员分布（按部门）
export function getMemberByDepartment(): Promise<PieChartData[]> {
  return request.get('/stats/members/by-department');
}

// 获取任务统计
export function getTaskStats(params?: { department_id?: string }): Promise<{
  total: number;
  pending: number;
  in_progress: number;
  completed: number;
  cancelled: number;
}> {
  return request.get('/stats/tasks', { params });
}

// 获取实习统计排名
export function getInternshipRanking(params: {
  top?: number;
  department_id?: string;
  start_date?: string;
  end_date?: string;
}): Promise<Array<{ user_id: string; username: string; real_name: string; total_hours: number; rank: number }>> {
  return request.get('/stats/internships/ranking', { params });
}

// 获取面试通过率统计
export function getInterviewStats(params?: {
  start_date?: string;
  end_date?: string;
}): Promise<{
  total: number;
  passed: number;
  rejected: number;
  pass_rate: number;
}> {
  return request.get('/stats/interviews', { params });
}
