import React from 'react';
import { Alert, Button, Card, Dropdown, Empty, Space } from 'antd';
import { DownloadOutlined } from '@ant-design/icons';
import type { EChartsOption } from 'echarts';
import { useECharts } from './useECharts';

export interface ChartCardProps {
  title: string;
  loading?: boolean;
  error?: string;
  option?: EChartsOption;
  empty?: boolean;
  height?: number;
  onExport?: (format: 'csv' | 'excel') => void;
  actions?: React.ReactNode;
}

const ChartCard: React.FC<ChartCardProps> = ({
  title,
  loading = false,
  error,
  option,
  empty,
  height = 300,
  onExport,
  actions,
}) => {
  const hideChart = !!error || (!!empty && !loading);
  const { ref, chart } = useECharts(hideChart ? undefined : option, loading && !error);

  const extra = (
    <Space size={8}>
      {actions}
      <Button
        size="small"
        icon={<DownloadOutlined />}
        onClick={() => {
          const inst = chart.current;
          if (!inst) return;
          const url = inst.getDataURL({ type: 'png', pixelRatio: 2, backgroundColor: '#fff' });
          const a = document.createElement('a');
          a.href = url;
          a.download = `${title}.png`;
          a.click();
        }}
        disabled={hideChart || loading}
      >
        PNG
      </Button>
      {onExport ? (
        <Dropdown
          menu={{
            items: [
              { key: 'excel', label: 'Excel' },
              { key: 'csv', label: 'CSV' },
            ],
            onClick: ({ key }) => onExport(key === 'csv' ? 'csv' : 'excel'),
          }}
        >
          <Button size="small" disabled={loading}>导出</Button>
        </Dropdown>
      ) : null}
    </Space>
  );

  return (
    <Card title={title} extra={extra} size="small">
      {error ? <Alert type="error" message={error} showIcon style={{ marginBottom: 12 }} /> : null}
      <div style={{ position: 'relative', height }}>
        {empty && !loading && !error ? (
          <div style={{ position: 'absolute', inset: 0, zIndex: 1, display: 'flex', alignItems: 'center', justifyContent: 'center', background: '#fff' }}>
            <Empty description="暂无数据" />
          </div>
        ) : null}
        <div ref={ref} style={{ width: '100%', height, visibility: error ? 'hidden' : 'visible' }} />
      </div>
    </Card>
  );
};

export default ChartCard;
