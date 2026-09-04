import React from 'react';
import { Button, Descriptions, Modal, Tag, Typography } from 'antd';
import dayjs from 'dayjs';
import type { AuditLogItem } from '@/api/audit';
import { actionColorMap, formatJSON, methodColorMap, statusColorMap } from './auditColumns';

const { Text } = Typography;

interface AuditDetailModalProps {
  open: boolean;
  loading: boolean;
  detail: AuditLogItem | null;
  onClose: () => void;
}

const AuditDetailModal: React.FC<AuditDetailModalProps> = ({
  open,
  loading,
  detail,
  onClose,
}) => {
  return (
    <Modal
      title="审计日志详情"
      open={open}
      onCancel={onClose}
      footer={[
        <Button key="close" onClick={onClose}>
          关闭
        </Button>,
      ]}
      width={800}
      destroyOnClose
    >
      {loading ? (
        <div style={{ textAlign: 'center', padding: 48 }}>加载中...</div>
      ) : detail ? (
        <div>
          <Descriptions bordered column={2} size="small">
            <Descriptions.Item label="日志ID" span={2}>
              <Text copyable style={{ fontFamily: 'monospace', fontSize: 12 }}>
                {detail.id}
              </Text>
            </Descriptions.Item>
            <Descriptions.Item label="用户名">{detail.user?.username || '未认证'}</Descriptions.Item>
            <Descriptions.Item label="姓名">{detail.user?.real_name || '-'}</Descriptions.Item>
            <Descriptions.Item label="动作">
              <Tag color={actionColorMap[detail.action] || 'default'}>{detail.action}</Tag>
            </Descriptions.Item>
            <Descriptions.Item label="模块">{detail.module || '-'}</Descriptions.Item>
            <Descriptions.Item label="HTTP 方法">
              <Tag color={methodColorMap[detail.method] || 'default'}>{detail.method}</Tag>
            </Descriptions.Item>
            <Descriptions.Item label="响应状态">
              <Tag color={statusColorMap(detail.response_code)}>{detail.response_code}</Tag>
            </Descriptions.Item>
            <Descriptions.Item label="请求路径" span={2}>
              <Text style={{ fontFamily: 'monospace' }}>{detail.path}</Text>
            </Descriptions.Item>
            <Descriptions.Item label="IP 地址">{detail.ip_address}</Descriptions.Item>
            <Descriptions.Item label="耗时">{detail.duration_ms} ms</Descriptions.Item>
            <Descriptions.Item label="时间" span={2}>
              {detail.timestamp ? dayjs(detail.timestamp).format('YYYY-MM-DD HH:mm:ss') : '-'}
            </Descriptions.Item>
            <Descriptions.Item label="User-Agent" span={2}>
              <Text style={{ fontSize: 12 }}>{detail.user_agent || '-'}</Text>
            </Descriptions.Item>
          </Descriptions>
          <div style={{ marginTop: 16 }}>
            <Text strong>请求参数：</Text>
            <pre style={preStyle}>{formatJSON(detail.request_body)}</pre>
          </div>
        </div>
      ) : (
        <div style={{ textAlign: 'center', padding: 48 }}>无数据</div>
      )}
    </Modal>
  );
};

const preStyle: React.CSSProperties = {
  background: '#f5f5f5',
  padding: 12,
  borderRadius: 4,
  maxHeight: 200,
  overflow: 'auto',
  fontSize: 12,
  whiteSpace: 'pre-wrap',
  wordBreak: 'break-all',
};

export default AuditDetailModal;
