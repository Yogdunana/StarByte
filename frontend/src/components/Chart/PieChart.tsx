import React, { useMemo } from 'react';
import type { EChartsOption } from 'echarts';
import Chart from './Chart';
import ChartCard from './ChartCard';
import { CHART_COLORS } from './palette';

export interface PieChartProps {
  data: { name: string; value: number }[];
  title?: string;
  height?: number | string;
  loading?: boolean;
  theme?: 'light' | 'dark';
}

function buildOption(data: { name: string; value: number }[]): EChartsOption {
  return {
    color: CHART_COLORS,
    tooltip: { trigger: 'item' },
    legend: { bottom: 0 },
    series: [
      {
        type: 'pie',
        radius: ['36%', '64%'],
        data: data.map((d) => ({ name: d.name, value: d.value })),
      },
    ],
  };
}

const PieChart: React.FC<PieChartProps> = ({ data, title, height = 300, loading, theme }) => {
  const option = useMemo(() => buildOption(data), [data]);
  if (title) {
    const cardHeight = typeof height === 'number' ? height : 300;
    return (
      <ChartCard
        title={title}
        option={option}
        loading={loading}
        height={cardHeight}
        empty={!data.length}
      />
    );
  }
  return <Chart option={option} loading={loading} height={height} theme={theme} />;
};

export default PieChart;
