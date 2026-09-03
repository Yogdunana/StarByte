import React, { useState, useEffect, useCallback } from 'react';
import {
  Table,
  Card,
  Button,
  Input,
  Select,
  Space,
  Modal,
  message,
  Tag,
  DatePicker,
  Tooltip,
  Descriptions,
  Typography,
} from 'antd';
import {
  SearchOutlined,
  ReloadOutlined,
  ExportOutlined,
  EyeOutlined,
  DatabaseOutlined,
} from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import dayjs from 'dayjs';

import {
  getAuditLogList,
  getAuditLogDetail,
  exportAuditLogs,
  triggerArchive,
  type AuditLogListItem,
  type AuditLogDetail,
  type ListAuditLogParams,
} from '@/api/audit';

const { RangePicker } = DatePicker;
const { Text, Paragraph } = Typography;

// HTTP 方法颜色映射
const methodColorMap: Record<string, string> = {
  POST: 'green',
  PUT: 'blue',
  PATCH: 'orange',
  DELETE: 'red',
};

// 响应状态码颜色映射
const statusColorMap = (status: number): string => {
  if (status >= 200 && status < 300) return 'success';
  if (status >= 300 && status < 400) return 'blue';
  if (status >= 400 && status < 500) return 'warning';
  return 'error';
};

const AuditList: React.FC = () => {
  const [loading, setLoading] = useState(false);
  const [data, setData] = useState<AuditLogListItem[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(20);

  // 筛选条件
  const [username, setUsername] = useState('');
  const [method, setMethod] = useState<string | undefined>();
  const [path, setPath] = useState('');
  const [ip, setIp] = useState('');
  const [timeRange, setTimeRange] = useState<[dayjs.Dayjs, dayjs.Dayjs] | null>(null);

  // 详情弹窗
  const [detailVisible, setDetailVisible] = useState(false);
  const [detailLoading, setDetailLoading] = useState(false);
  const [detail, setDetail] = useState<AuditLogDetail | null>(null);

  // 导出状态
  const [exporting, setExporting] = useState(false);
  const [archiving, setArchiving] = useState(false);

  // 加载审计日志列表
  const loadList = useCallback(async () => {
    setLoading(true);
    try {
      const params: ListAuditLogParams = {
        page,
        page_size: pageSize,
        username: username || undefined,
        method: method || undefined,
        path: path || undefined,
        ip: ip || undefined,
      };

      if (timeRange) {
        params.start_time = timeRange[0].toISOString();
        params.end_time = timeRange[1].toISOString();
      }

      const res = await getAuditLogList(params);
      setData(res.list);
      setTotal(res.total);
    } catch {
      message.error('加载审计日志列表失败');
    } finally {
      setLoading(false);
    }
  }, [page, pageSize, username, method, path, ip, timeRange]);

  useEffect(() => {
    loadList();
  }, [page, pageSize]); // eslint-disable-line react-hooks/exhaustive-deps

  // 搜索
  const handleSearch = () => {
    setPage(1);
    loadList();
  };

  // 重置
  const handleReset = () => {
    setUsername('');
    setMethod(undefined);
    setPath('');
    setIp('');
    setTimeRange(null);
    setPage(1);
    setTimeout(loadList, 0);
  };

  // 查看详情
  const handleViewDetail = async (record: AuditLogListItem) => {
    setDetailVisible(true);
    setDetailLoading(true);
    try {
      const res = await getAuditLogDetail(record.id);
      setDetail(res);
    } catch {
      message.error('加载审计日志详情失败');
    } finally {
      setDetailLoading(false);
    }
  };

  // 导出
  const handleExport = async (format: 'csv' | 'json') => {
    setExporting(true);
    try {
      const params: Parameters<typeof exportAuditLogs>[0] = {
        format,
        username: username || undefined,
        method: method || undefined,
        path: path || undefined,
        ip: ip || undefined,
      };

      if (timeRange) {
        params.start_time = timeRange[0].toISOString();
        params.end_time = timeRange[1].toISOString();
      }

      await exportAuditLogs(params);
      message.success(`导出成功`);
    } catch {
      message.error('导出失败');
    } finally {
      setExporting(false);
    }
  };

  // 手动归档
  const handleArchive = () => {
    Modal.confirm({
      title: '手动触发归档',
      content: '将归档 90 天前的审计日志到 MinIO，确定继续吗？',
      onOk: async () => {
        setArchiving(true);
        try {
          const res = await triggerArchive(90);
          message.success(res.message || `成功归档 ${res.record_count} 条日志`);
          loadList();
        } catch {
          message.error('归档失败');
        } finally {
          setArchiving(false);
        }
      },
    });
  };

  // 格式化 JSON 字符串
  const formatJSON = (str: string): string => {
    if (!str) return '-';
    try {
      return JSON.stringify(JSON.parse(str), null, 2);
    } catch {
      return str;
    }
  };

  const columns: ColumnsType<AuditLogListItem> = [
    {
      title: '时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 170,
      sorter: true,
      render: (time: string) => (time ? dayjs(time).format('YYYY-MM-DD HH:mm:ss') : '-'),
    },
    {
      title: '用户',
      dataIndex: 'username',
      key: 'username',
      width: 100,
      render: (name: string) => name || '未认证',
    },
    {
      title: '方法',
      dataIndex: 'method',
      key: 'method',
      width: 80,
      render: (m: string) =>
        m ? <Tag color={methodColorMap[m] || 'default'}>{m}</Tag> : '-',
    },
    {
      title: '操作路径',
      dataIndex: 'path',
      key: 'path',
      width: 250,
      ellipsis: { showTitle: false },
      render: (path: string) => (
        <Tooltip title={path}>
          <span style={{ fontFamily: 'monospace', fontSize: 12 }}>{path}</span>
        </Tooltip>
      ),
    },
    {
      title: 'IP',
      dataIndex: 'ip',
      key: 'ip',
      width: 130,
    },
    {
      title: '状态码',
      dataIndex: 'response_status',
      key: 'response_status',
      width: 90,
      render: (status: number) => <Tag color={statusColorMap(status)}>{status}</Tag>,
    },
    {
      title: '耗时',
      dataIndex: 'duration_ms',
      key: 'duration_ms',
      width: 80,
      render: (ms: number) => `${ms}ms`,
      sorter: true,
    },
    {
      title: '请求ID',
      dataIndex: 'request_id',
      key: 'request_id',
      width: 120,
      ellipsis: true,
      render: (id: string) => (
        <Tooltip title={id}>
          <Text type="secondary" style={{ fontSize: 12 }}>
            {id ? id.substring(0, 8) + '...' : '-'}
          </Text>
        </Tooltip>
      ),
    },
    {
      title: '操作',
      key: 'action',
      width: 80,
      fixed: 'right',
      render: (_, record) => (
        <Button
          type="link"
          size="small"
          icon={<EyeOutlined />}
          onClick={() => handleViewDetail(record)}
        >
          详情
        </Button>
      ),
    },
  ];

  return (
    <Card>
      {/* 搜索栏 */}
      <div style={{ marginBottom: 16, display: 'flex', flexWrap: 'wrap', gap: 12, alignItems: 'center' }}>
        <Input
          placeholder="用户名"
          prefix={<SearchOutlined />}
          value={username}
          onChange={(e) => setUsername(e.target.value)}
          style={{ width: 150 }}
          onPressEnter={handleSearch}
          allowClear
        />
        <Select
          placeholder="HTTP 方法"
          value={method}
          onChange={setMethod}
          style={{ width: 120 }}
          allowClear
        >
          <Select.Option value="POST">POST</Select.Option>
          <Select.Option value="PUT">PUT</Select.Option>
          <Select.Option value="PATCH">PATCH</Select.Option>
          <Select.Option value="DELETE">DELETE</Select.Option>
        </Select>
        <Input
          placeholder="请求路径"
          value={path}
          onChange={(e) => setPath(e.target.value)}
          style={{ width: 200 }}
          onPressEnter={handleSearch}
          allowClear
        />
        <Input
          placeholder="IP 地址"
          value={ip}
          onChange={(e) => setIp(e.target.value)}
          style={{ width: 150 }}
          onPressEnter={handleSearch}
          allowClear
        />
        <RangePicker
          showTime
          format="YYYY-MM-DD HH:mm"
          value={timeRange}
          onChange={(values) => {
            if (values && values[0] && values[1]) {
              setTimeRange([values[0], values[1]]);
            } else {
              setTimeRange(null);
            }
          }}
        />
        <Space>
          <Button type="primary" onClick={handleSearch}>
            搜索
          </Button>
          <Button icon={<ReloadOutlined />} onClick={handleReset}>
            重置
          </Button>
        </Space>
        <div style={{ flex: 1 }} />
        <Space>
          <Button
            icon={<ExportOutlined />}
            loading={exporting}
            onClick={() => handleExport('csv')}
          >
            导出 CSV
          </Button>
          <Button
            icon={<ExportOutlined />}
            loading={exporting}
            onClick={() => handleExport('json')}
          >
            导出 JSON
          </Button>
          <Button
            type="primary"
            ghost
            icon={<DatabaseOutlined />}
            loading={archiving}
            onClick={handleArchive}
          >
            归档
          </Button>
        </Space>
      </div>

      {/* 表格 */}
      <Table
        columns={columns}
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
          showTotal: (total) => `共 ${total} 条`,
          onChange: (p, ps) => {
            setPage(p);
            setPageSize(ps);
          },
        }}
      />

      {/* 详情弹窗 */}
      <Modal
        title="审计日志详情"
        open={detailVisible}
        onCancel={() => {
          setDetailVisible(false);
          setDetail(null);
        }}
        footer={[
          <Button key="close" onClick={() => setDetailVisible(false)}>
            关闭
          </Button>,
        ]}
        width={800}
        destroyOnClose
      >
        {detailLoading ? (
          <div style={{ textAlign: 'center', padding: 48 }}>加载中...</div>
        ) : detail ? (
          <div>
            <Descriptions bordered column={2} size="small">
              <Descriptions.Item label="日志ID" span={2}>
                <Text copyable style={{ fontFamily: 'monospace', fontSize: 12 }}>
                  {detail.id}
                </Text>
              </Descriptions.Item>
              <Descriptions.Item label="用户名">{detail.username || '未认证'}</Descriptions.Item>
              <Descriptions.Item label="用户ID">
                {detail.user_id || '-'}
              </Descriptions.Item>
              <Descriptions.Item label="HTTP 方法">
                <Tag color={methodColorMap[detail.method] || 'default'}>{detail.method}</Tag>
              </Descriptions.Item>
              <Descriptions.Item label="响应状态">
                <Tag color={statusColorMap(detail.response_status)}>
                  {detail.response_status}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label="操作" span={2}>
                <Text style={{ fontFamily: 'monospace' }}>{detail.operation}</Text>
              </Descriptions.Item>
              <Descriptions.Item label="请求路径" span={2}>
                <Text style={{ fontFamily: 'monospace' }}>{detail.path}</Text>
              </Descriptions.Item>
              <Descriptions.Item label="IP 地址">{detail.ip}</Descriptions.Item>
              <Descriptions.Item label="耗时">{detail.duration_ms} ms</Descriptions.Item>
              <Descriptions.Item label="请求ID">
                <Text copyable style={{ fontFamily: 'monospace', fontSize: 12 }}>
                  {detail.request_id || '-'}
                </Text>
              </Descriptions.Item>
              <Descriptions.Item label="时间">
                {detail.created_at ? dayjs(detail.created_at).format('YYYY-MM-DD HH:mm:ss') : '-'}
              </Descriptions.Item>
              <Descriptions.Item label="已归档" span={2}>
                {detail.is_archived ? <Tag color="blue">已归档</Tag> : <Tag>未归档</Tag>}
              </Descriptions.Item>
              <Descriptions.Item label="User-Agent" span={2}>
                <Text style={{ fontSize: 12 }}>{detail.user_agent || '-'}</Text>
              </Descriptions.Item>
            </Descriptions>

            <div style={{ marginTop: 16 }}>
              <Text strong>请求参数：</Text>
              <pre
                style={{
                  background: '#f5f5f5',
                  padding: 12,
                  borderRadius: 4,
                  maxHeight: 200,
                  overflow: 'auto',
                  fontSize: 12,
                  whiteSpace: 'pre-wrap',
                  wordBreak: 'break-all',
                }}
              >
                {formatJSON(detail.request_params)}
              </pre>
            </div>

            <div style={{ marginTop: 12 }}>
              <Text strong>响应内容：</Text>
              <pre
                style={{
                  background: '#f5f5f5',
                  padding: 12,
                  borderRadius: 4,
                  maxHeight: 200,
                  overflow: 'auto',
                  fontSize: 12,
                  whiteSpace: 'pre-wrap',
                  wordBreak: 'break-all',
                }}
              >
                {formatJSON(detail.response_body)}
              </pre>
            </div>
          </div>
        ) : (
          <div style={{ textAlign: 'center', padding: 48 }}>无数据</div>
        )}
      </Modal>
    </Card>
  );
};

export default AuditList;
