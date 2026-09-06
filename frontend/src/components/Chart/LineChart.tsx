import React, { useMemo } from 'react';
import type { EChartsOption } from 'echarts';
import Chart from './Chart';
import ChartCard from './ChartCard';
import { CHART_COLORS } from './palette';

export interface LineChartProps {
  categories: string[];
  series: { name: string; data: number[] }[];
  title?: string;
  area?: boolean;
  height?: number | string;
  loading?: boolean;
  theme?: 'light' | 'dark';
}

function buildOption(
  categories: string[],
  series: { name: string; data: number[] }[],
  area: boolean,
): EChartsOption {
  return {
    color: CHART_COLORS,
    tooltip: { trigger: 'axis' },
    legend: series.length > 1 ? { bottom: 0 } : undefined,
    grid: { left: 40, right: 16, top: 24, bottom: series.length > 1 ? 48 : 32 },
    xAxis: { type: 'category', data: categories },
    yAxis: { type: 'value' },
    series: series.map((s) => ({
      name: s.name,
      type: 'line' as const,
      smooth: true,
      areaStyle: area ? { opacity: 0.18 } : undefined,
      data: s.data,
    })),
  };
}

const LineChart: React.FC<LineChartProps> = ({
  categories, series, title, area = false, height = 300, loading, theme,
}) => {
  const option = useMemo(
    () => buildOption(categories, series, area),
    [categories, series, area],
  );
  if (title) {
    const cardHeight = typeof height === 'number' ? height : 300;
    return (
      <ChartCard
        title={title}
        option={option}
        loading={loading}
        height={cardHeight}
        empty={!categories.length}
      />
    );
  }
  return <Chart option={option} loading={loading} height={height} theme={theme} />;
};

export default LineChart;
