import React, { useState, useEffect, useCallback } from 'react';
import { Table, Card, Button, Space, Modal, message } from 'antd';
import { ExportOutlined, DatabaseOutlined } from '@ant-design/icons';
import {
  getAuditLogList,
  getAuditLogDetail,
  exportAuditLogs,
  triggerArchive,
  type AuditLogItem,
  type AuditQueryParams,
} from '@/api/audit';
import { usePermission } from '@/hooks/usePermission';
import AuditFilterBar, { type AuditFilterValue } from './AuditFilterBar';
import AuditDetailModal from './AuditDetailModal';
import { buildAuditColumns } from './auditColumns';

const emptyFilter: AuditFilterValue = {
  username: '',
  action: undefined,
  module: '',
  keyword: '',
  ipAddress: '',
  timeRange: null,
};

const AuditList: React.FC = () => {
  const canExport = usePermission('audit:export');
  const canArchive = usePermission('audit:archive');
  const [loading, setLoading] = useState(false);
  const [data, setData] = useState<AuditLogItem[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(20);
  const [filters, setFilters] = useState<AuditFilterValue>(emptyFilter);
  const [detailVisible, setDetailVisible] = useState(false);
  const [detailLoading, setDetailLoading] = useState(false);
  const [detail, setDetail] = useState<AuditLogItem | null>(null);
  const [exporting, setExporting] = useState(false);
  const [archiving, setArchiving] = useState(false);

  const buildParams = useCallback(
    (p: number, ps: number): AuditQueryParams => {
      const params: AuditQueryParams = {
        page: p,
        page_size: ps,
        username: filters.username || undefined,
        action: filters.action,
        module: filters.module || undefined,
        keyword: filters.keyword || undefined,
        ip_address: filters.ipAddress || undefined,
      };
      if (filters.timeRange) {
        params.start_time = filters.timeRange[0].toISOString();
        params.end_time = filters.timeRange[1].toISOString();
      }
      return params;
    },
    [filters],
  );

  const loadList = useCallback(async (params: AuditQueryParams) => {
    setLoading(true);
    try {
      const res = await getAuditLogList(params);
      setData(res.list);
      setTotal(res.total);
    } catch {
      message.error('加载审计日志列表失败');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    loadList(buildParams(page, pageSize));
  }, [page, pageSize, buildParams, loadList]);

  const handleSearch = () => {
    setPage(1);
    loadList(buildParams(1, pageSize));
  };

  const handleReset = () => {
    setFilters(emptyFilter);
    setPage(1);
    loadList({ page: 1, page_size: pageSize });
  };

  const handleViewDetail = async (record: AuditLogItem) => {
    setDetailVisible(true);
    setDetailLoading(true);
    try {
      setDetail(await getAuditLogDetail(record.id));
    } catch {
      message.error('加载审计日志详情失败');
    } finally {
      setDetailLoading(false);
    }
  };

  const handleExport = async (format: 'csv' | 'excel') => {
    setExporting(true);
    try {
      await exportAuditLogs({ ...buildParams(page, pageSize), format });
      message.success('导出成功');
    } catch {
      message.error('导出失败');
    } finally {
      setExporting(false);
    }
  };

  const handleArchive = () => {
    Modal.confirm({
      title: '手动触发归档',
      content: '将归档 90 天前的审计日志到 MinIO 并删除原记录，确定继续吗？',
      onOk: async () => {
        setArchiving(true);
        try {
          const res = await triggerArchive(90);
          message.success(res.message || `成功归档 ${res.record_count} 条日志`);
          loadList(buildParams(page, pageSize));
        } catch {
          message.error('归档失败');
        } finally {
          setArchiving(false);
        }
      },
    });
  };

  return (
    <Card>
      <AuditFilterBar
        value={filters}
        onChange={setFilters}
        onSearch={handleSearch}
        onReset={handleReset}
      />
      <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'flex-end' }}>
        <Space>
          {canExport && (
            <>
              <Button icon={<ExportOutlined />} loading={exporting} onClick={() => handleExport('csv')}>
                导出 CSV
              </Button>
              <Button icon={<ExportOutlined />} loading={exporting} onClick={() => handleExport('excel')}>
                导出 Excel
              </Button>
            </>
          )}
          {canArchive && (
            <Button
              type="primary"
              ghost
              icon={<DatabaseOutlined />}
              loading={archiving}
              onClick={handleArchive}
            >
              归档
            </Button>
          )}
        </Space>
      </div>
      <Table
        columns={buildAuditColumns(handleViewDetail)}
        dataSource={data}
        rowKey="id"
        loading={loading}
        scroll={{ x: 1200 }}
        size="middle"
        pagination={{
          current: page,
          pageSize,
          total,
          showSizeChanger: true,
          showQuickJumper: true,
          showTotal: (t) => `共 ${t} 条`,
          onChange: (p, ps) => {
            setPage(p);
            setPageSize(ps);
          },
        }}
      />
      <AuditDetailModal
        open={detailVisible}
        loading={detailLoading}
        detail={detail}
        onClose={() => {
          setDetailVisible(false);
          setDetail(null);
        }}
      />
    </Card>
  );
};

export default AuditList;
