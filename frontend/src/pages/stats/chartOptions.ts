import type { EChartsOption } from 'echarts';
import type { StatsSeries } from '@/api/stats';

export function seriesEmpty(series: StatsSeries | undefined): boolean {
  return !series || !series.data || series.data.length === 0;
}

export function pieOption(series: StatsSeries): EChartsOption {
  return {
    tooltip: { trigger: 'item' },
    legend: { bottom: 0 },
    series: [
      {
        name: series.name,
        type: 'pie',
        radius: ['36%', '64%'],
        data: series.data.map((d) => ({ name: d.label, value: d.value })),
      },
    ],
  };
}

export function barOption(series: StatsSeries, horizontal = false): EChartsOption {
  const cats = series.x_axis?.length ? series.x_axis : series.data.map((d) => d.label);
  const values = series.data.map((d) => d.value);
  if (horizontal) {
    return {
      tooltip: { trigger: 'axis' },
      grid: { left: 80, right: 24, top: 24, bottom: 24 },
      xAxis: { type: 'value', minInterval: 1 },
      yAxis: { type: 'category', data: cats },
      series: [{ name: series.name, type: 'bar', data: values, barMaxWidth: 48 }],
    };
  }
  return {
    tooltip: { trigger: 'axis' },
    grid: { left: 40, right: 16, top: 24, bottom: 32 },
    xAxis: { type: 'category', data: cats },
    yAxis: { type: 'value', minInterval: 1 },
    series: [{ name: series.name, type: 'bar', data: values, barMaxWidth: 48 }],
  };
}

export function lineOption(series: StatsSeries): EChartsOption {
  const cats = series.x_axis?.length ? series.x_axis : series.data.map((d) => d.label);
  return {
    tooltip: { trigger: 'axis' },
    grid: { left: 40, right: 16, top: 24, bottom: 32 },
    xAxis: { type: 'category', data: cats },
    yAxis: { type: 'value' },
    series: [{ name: series.name, type: 'line', smooth: true, data: series.data.map((d) => d.value) }],
  };
}

export function gaugeOption(series: StatsSeries): EChartsOption {
  const value = series.data[0]?.value ?? 0;
  return {
    series: [
      {
        type: 'gauge',
        min: 0,
        max: 100,
        progress: { show: true },
        detail: { formatter: '{value}%', fontSize: 16 },
        data: [{ value: Number(value.toFixed(1)), name: series.name }],
      },
    ],
  };
}

export function calendarOption(series: StatsSeries): EChartsOption {
  const values = series.data.map((d) => d.value);
  const max = values.length ? Math.max(...values) : 1;
  const labels = series.data.map((d) => d.label).filter(Boolean).sort();
  const year = String(new Date().getFullYear());
  const calRange: string | [string, string] = labels.length
    ? (labels[0].slice(0, 4) === labels[labels.length - 1].slice(0, 4)
      ? labels[0].slice(0, 4)
      : [labels[0], labels[labels.length - 1]])
    : year;
  return {
    tooltip: {
      formatter: (params: unknown) => {
        const p = params as { value?: [string, number] };
        return `${p.value?.[0] || ''}：${p.value?.[1] ?? 0}`;
      },
    },
    visualMap: { min: 0, max, orient: 'horizontal', left: 'center', bottom: 0, calculable: true },
    calendar: { range: calRange, left: 48, right: 16, top: 32, bottom: 48, cellSize: ['auto', 16] },
    series: [
      {
        type: 'heatmap',
        coordinateSystem: 'calendar',
        data: series.data.map((d) => [d.label, d.value]),
      },
    ],
  };
}

export function stackedBarOption(seriesList: StatsSeries[]): EChartsOption {
  const cats: string[] = [];
  seriesList.forEach((s) => {
    (s.x_axis?.length ? s.x_axis : s.data.map((d) => d.label)).forEach((c) => {
      if (!cats.includes(c)) cats.push(c);
    });
  });
  return {
    tooltip: { trigger: 'axis' },
    legend: { bottom: 0 },
    grid: { left: 40, right: 16, top: 24, bottom: 48 },
    xAxis: { type: 'category', data: cats },
    yAxis: { type: 'value', minInterval: 1 },
    series: seriesList.map((s) => ({
      name: s.name,
      type: 'bar' as const,
      stack: 'status',
      barMaxWidth: 48,
      data: cats.map((cat) => s.data.find((d) => d.label === cat)?.value ?? 0),
    })),
  };
}
