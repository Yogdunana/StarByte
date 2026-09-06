import React from 'react';
import { Card, Tabs } from 'antd';
import type { TabsProps } from 'antd';
import { usePermission } from '@/hooks/usePermission';
import AuditLogPanel from './AuditLogPanel';
import AuditTracePanel from './AuditTracePanel';
import AuditArchivePanel from './AuditArchivePanel';
import AuditReportPanel from './AuditReportPanel';

const AuditList: React.FC = () => {
  const canReport = usePermission('audit:report');
  const items: TabsProps['items'] = [
    { key: 'logs', label: '操作日志', children: <AuditLogPanel /> },
    { key: 'trace', label: '变更追踪', children: <AuditTracePanel /> },
    { key: 'archives', label: '归档查询', children: <AuditArchivePanel /> },
  ];
  if (canReport) {
    items.push({ key: 'report', label: '合规报告', children: <AuditReportPanel /> });
  }
  return (
    <Card>
      <Tabs items={items} />
    </Card>
  );
};

export default AuditList;
