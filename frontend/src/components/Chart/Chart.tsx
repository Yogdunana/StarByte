import React from 'react';
import type { EChartsOption } from 'echarts';
import { useSelector } from 'react-redux';
import { selectTheme } from '@/store/slices/appSlice';
import { useECharts } from './useECharts';

export interface ChartProps {
  option?: EChartsOption;
  loading?: boolean;
  height?: number | string;
  autoResize?: boolean;
  theme?: 'light' | 'dark';
  className?: string;
}

const Chart: React.FC<ChartProps> = ({
  option,
  loading = false,
  height = 300,
  autoResize = true,
  theme: themeProp,
  className,
}) => {
  const storeTheme = useSelector(selectTheme);
  const theme = themeProp ?? storeTheme;
  const { ref } = useECharts(option, loading, theme, autoResize);
  const h = typeof height === 'number' ? `${height}px` : height;
  return <div className={className} ref={ref} style={{ width: '100%', height: h }} />;
};

export default Chart;
