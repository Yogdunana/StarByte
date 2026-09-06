import { useEffect, useRef } from 'react';
import * as echarts from 'echarts';
import type { EChartsOption } from 'echarts';

export function useECharts(
  option: EChartsOption | undefined,
  loading: boolean,
  theme: 'light' | 'dark' = 'light',
  autoResize = true,
) {
  const ref = useRef<HTMLDivElement>(null);
  const chartRef = useRef<echarts.ECharts | null>(null);

  useEffect(() => {
    const el = ref.current;
    if (!el) {
      return undefined;
    }
    const old = echarts.getInstanceByDom(el);
    old?.dispose();
    const chart = echarts.init(el, theme === 'dark' ? 'dark' : undefined);
    chartRef.current = chart;
    const onResize = () => {
      chart.resize();
    };
    let observer: ResizeObserver | undefined;
    if (autoResize) {
      window.addEventListener('resize', onResize);
      observer = new ResizeObserver(onResize);
      observer.observe(el);
    }
    return () => {
      observer?.disconnect();
      if (autoResize) {
        window.removeEventListener('resize', onResize);
      }
      chart.dispose();
      chartRef.current = null;
    };
  }, [theme, autoResize]);

  useEffect(() => {
    const chart = chartRef.current;
    if (!chart) {
      return;
    }
    if (loading) {
      chart.showLoading();
      return;
    }
    chart.hideLoading();
    if (option) {
      chart.setOption(option, true);
      requestAnimationFrame(() => {
        chart.resize();
      });
    }
  }, [option, loading, theme]);

  return { ref, chart: chartRef };
}
