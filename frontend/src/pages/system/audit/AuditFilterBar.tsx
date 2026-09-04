import React from 'react';
import { Button, Input, Select, Space, DatePicker } from 'antd';
import { SearchOutlined, ReloadOutlined } from '@ant-design/icons';
import type { Dayjs } from 'dayjs';
import dayjs from 'dayjs';

const { RangePicker } = DatePicker;

export interface AuditFilterValue {
  username: string;
  action?: string;
  module: string;
  keyword: string;
  ipAddress: string;
  timeRange: [Dayjs, Dayjs] | null;
}

interface AuditFilterBarProps {
  value: AuditFilterValue;
  onChange: (next: AuditFilterValue) => void;
  onSearch: () => void;
  onReset: () => void;
}

const rangePresets = [
  { label: '今天', value: [dayjs().startOf('day'), dayjs().endOf('day')] as [Dayjs, Dayjs] },
  { label: '本周', value: [dayjs().startOf('week'), dayjs().endOf('week')] as [Dayjs, Dayjs] },
  { label: '本月', value: [dayjs().startOf('month'), dayjs().endOf('month')] as [Dayjs, Dayjs] },
  { label: '最近 90 天', value: [dayjs().subtract(90, 'day'), dayjs()] as [Dayjs, Dayjs] },
];

const AuditFilterBar: React.FC<AuditFilterBarProps> = ({
  value,
  onChange,
  onSearch,
  onReset,
}) => {
  const patch = (partial: Partial<AuditFilterValue>) => onChange({ ...value, ...partial });

  return (
    <div style={{ marginBottom: 16, display: 'flex', flexWrap: 'wrap', gap: 12, alignItems: 'center' }}>
      <Input
        placeholder="用户名"
        prefix={<SearchOutlined />}
        value={value.username}
        onChange={(e) => patch({ username: e.target.value })}
        style={{ width: 150 }}
        onPressEnter={onSearch}
        allowClear
      />
      <Select
        placeholder="操作类型"
        value={value.action}
        onChange={(action) => patch({ action })}
        style={{ width: 140 }}
        allowClear
        options={[
          { value: 'CREATE', label: 'CREATE' },
          { value: 'UPDATE', label: 'UPDATE' },
          { value: 'DELETE', label: 'DELETE' },
          { value: 'LOGIN', label: 'LOGIN' },
          { value: 'LOGOUT', label: 'LOGOUT' },
        ]}
      />
      <Input
        placeholder="模块"
        value={value.module}
        onChange={(e) => patch({ module: e.target.value })}
        style={{ width: 140 }}
        onPressEnter={onSearch}
        allowClear
      />
      <Input
        placeholder="关键词（路径/参数）"
        value={value.keyword}
        onChange={(e) => patch({ keyword: e.target.value })}
        style={{ width: 200 }}
        onPressEnter={onSearch}
        allowClear
      />
      <Input
        placeholder="IP 地址"
        value={value.ipAddress}
        onChange={(e) => patch({ ipAddress: e.target.value })}
        style={{ width: 150 }}
        onPressEnter={onSearch}
        allowClear
      />
      <RangePicker
        showTime
        format="YYYY-MM-DD HH:mm"
        value={value.timeRange}
        presets={rangePresets}
        onChange={(values) => {
          if (values && values[0] && values[1]) {
            patch({ timeRange: [values[0], values[1]] });
          } else {
            patch({ timeRange: null });
          }
        }}
      />
      <Space>
        <Button type="primary" onClick={onSearch}>
          搜索
        </Button>
        <Button icon={<ReloadOutlined />} onClick={onReset}>
          重置
        </Button>
      </Space>
    </div>
  );
};

export default AuditFilterBar;
