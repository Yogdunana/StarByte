import React, { useMemo } from 'react';
import type { EChartsOption } from 'echarts';
import Chart from './Chart';
import ChartCard from './ChartCard';
import { CHART_COLORS } from './palette';

export interface BarChartProps {
  categories: string[];
  series: { name: string; data: number[] }[];
  title?: string;
  horizontal?: boolean;
  height?: number | string;
  loading?: boolean;
  theme?: 'light' | 'dark';
}

function buildOption(
  categories: string[],
  series: { name: string; data: number[] }[],
  horizontal: boolean,
): EChartsOption {
  const bars = series.map((s) => ({
    name: s.name,
    type: 'bar' as const,
    data: s.data,
    barMaxWidth: 48,
  }));
  if (horizontal) {
    return {
      color: CHART_COLORS,
      tooltip: { trigger: 'axis' },
      legend: series.length > 1 ? { bottom: 0 } : undefined,
      grid: { left: 80, right: 24, top: 24, bottom: series.length > 1 ? 48 : 24 },
      xAxis: { type: 'value', minInterval: 1 },
      yAxis: { type: 'category', data: categories },
      series: bars,
    };
  }
  return {
    color: CHART_COLORS,
    tooltip: { trigger: 'axis' },
    legend: series.length > 1 ? { bottom: 0 } : undefined,
    grid: { left: 40, right: 16, top: 24, bottom: series.length > 1 ? 48 : 32 },
    xAxis: { type: 'category', data: categories },
    yAxis: { type: 'value', minInterval: 1 },
    series: bars,
  };
}

const BarChart: React.FC<BarChartProps> = ({
  categories, series, title, horizontal = false, height = 300, loading, theme,
}) => {
  const option = useMemo(
    () => buildOption(categories, series, horizontal),
    [categories, series, horizontal],
  );
  const empty = !categories.length;
  const cardHeight = typeof height === 'number' ? height : 300;
  if (title) {
    return <ChartCard title={title} option={option} loading={loading} height={cardHeight} empty={empty} />;
  }
  return <Chart option={option} loading={loading} height={height} theme={theme} />;
};

export default BarChart;
