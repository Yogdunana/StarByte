import { tx, useLocale } from '@/i18n/text';
import React, { useState } from 'react';
import { Button, Card, DatePicker, Space, Statistic, Table, Typography, message } from 'antd';
import type { Dayjs } from 'dayjs';
import dayjs from 'dayjs';
import {
  downloadAuditReport,
  getAuditReport,
  type AuditCountItem,
  type AuditReport,
} from '@/api/audit';

const { RangePicker } = DatePicker;
const { Text } = Typography;

const AuditReportPanel: React.FC = () => {
  useLocale();
  const [range, setRange] = useState<[Dayjs, Dayjs] | null>([dayjs().subtract(30, 'day'), dayjs()]);
  const [loading, setLoading] = useState(false);
  const [downloading, setDownloading] = useState<string | null>(null);
  const [report, setReport] = useState<AuditReport | null>(null);

  const timeParams = () => {
    if (!range) return {};
    return {
      start_time: range[0].toISOString(),
      end_time: range[1].toISOString(),
    };
  };

  const load = async () => {
    setLoading(true);
    try {
      setReport(await getAuditReport(timeParams()));
    } catch {
      message.error(tx('生成报告失败'));
    } finally {
      setLoading(false);
    }
  };

  const download = async (format: 'csv' | 'pdf' | 'excel') => {
    setDownloading(format);
    try {
      await downloadAuditReport({ ...timeParams(), format });
      message.success(tx('报告已下载'));
    } catch {
      message.error(tx('下载报告失败'));
    } finally {
      setDownloading(null);
    }
  };

  const countColumns = [
    { title: tx('键'), dataIndex: 'key' },
    { title: tx('数量'), dataIndex: 'count', width: 100 },
  ];

  return (
    <div>
      <Space wrap style={{ marginBottom: 16 }}>
        <RangePicker
          showTime
          value={range}
          onChange={(values) => {
            if (values && values[0] && values[1]) {
              setRange([values[0], values[1]]);
            } else {
              setRange(null);
            }
          }}
        />
        <Button type="primary" loading={loading} onClick={load}>
          {tx('生成 JSON 报告')}
        </Button>
        <Button loading={downloading === 'csv'} onClick={() => download('csv')}>
          {tx('下载 CSV')}
        </Button>
        <Button loading={downloading === 'excel'} onClick={() => download('excel')}>
          {tx('下载 Excel')}
        </Button>
        <Button loading={downloading === 'pdf'} onClick={() => download('pdf')}>
          {tx('下载 PDF')}
        </Button>
      </Space>
      {report && (
        <>
          <Space size={32} style={{ marginBottom: 16 }}>
            <Statistic title={tx('总操作数')} value={report.total} />
            <Statistic title={tx('删除标记')} value={flagCount(report.by_compliance, 'delete')} />
            <Statistic
              title={tx('权限标记')}
              value={flagCount(report.by_compliance, 'permission')}
            />
            <Statistic title={tx('导出标记')} value={flagCount(report.by_compliance, 'export')} />
          </Space>
          {report.note && (
            <Text type="secondary" style={{ display: 'block', marginBottom: 16 }}>
              {report.note}
            </Text>
          )}
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 16 }}>
            <Card size="small" title={tx('按动作')}>
              <Table
                rowKey="key"
                size="small"
                pagination={false}
                columns={countColumns}
                dataSource={report.by_action}
              />
            </Card>
            <Card size="small" title={tx('按模块')}>
              <Table
                rowKey="key"
                size="small"
                pagination={false}
                columns={countColumns}
                dataSource={report.by_module}
              />
            </Card>
            <Card size="small" title={tx('合规标记')}>
              <Table
                rowKey="key"
                size="small"
                pagination={false}
                columns={countColumns}
                dataSource={report.by_compliance}
              />
            </Card>
            <Card size="small" title={tx('高频操作人')}>
              <Table
                rowKey="key"
                size="small"
                pagination={false}
                columns={countColumns}
                dataSource={report.top_operators}
              />
            </Card>
          </div>
        </>
      )}
    </div>
  );
};

function flagCount(items: AuditCountItem[] | undefined, key: string): number {
  return items?.find((i) => i.key === key)?.count ?? 0;
}

export default AuditReportPanel;
