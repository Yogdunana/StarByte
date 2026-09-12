import type { EChartsOption } from 'echarts';

export function gaugeOption(name: string, value: number): EChartsOption {
  const v = Number.isFinite(value) ? Number(value.toFixed(1)) : 0;
  return {
    series: [
      {
        type: 'gauge',
        min: 0,
        max: 100,
        progress: { show: true, width: 12 },
        axisLine: { lineStyle: { width: 12 } },
        pointer: { length: '60%' },
        detail: { formatter: '{value}%', fontSize: 16, offsetCenter: [0, '70%'] },
        data: [{ value: v, name }],
      },
    ],
  };
}

export function sparkOption(labels: string[], values: number[], name: string): EChartsOption {
  return {
    tooltip: { trigger: 'axis' },
    grid: { left: 36, right: 12, top: 20, bottom: 24 },
    xAxis: { type: 'category', data: labels, boundaryGap: false },
    yAxis: { type: 'value', min: 0 },
    series: [{ name, type: 'line', smooth: true, showSymbol: false, areaStyle: { opacity: 0.16 }, data: values }],
  };
}
