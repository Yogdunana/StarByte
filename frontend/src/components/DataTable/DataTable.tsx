import React from 'react';
import { Table, Card, Input, Space } from 'antd';
import { SearchOutlined } from '@ant-design/icons';
import type { ColumnsType, TablePaginationConfig } from 'antd/es/table';
import type { TablePagination } from '@/types/common';

export interface DataTableProps<T> {
  columns: ColumnsType<T>;
  dataSource: T[];
  rowKey: string | ((record: T) => string);
  loading?: boolean;
  pagination?: TablePagination;
  search?: {
    placeholder: string;
    onSearch: (value: string) => void;
    value?: string;
  };
  toolbar?: React.ReactNode;
  scroll?: { x?: number; y?: number };
}

/**
 * 数据表格 — 封装分页、搜索、工具栏
 */
function DataTable<T extends Record<string, unknown>>({
  columns,
  dataSource,
  rowKey,
  loading = false,
  pagination,
  search,
  toolbar,
  scroll,
}: DataTableProps<T>) {
  const paginationConfig: TablePaginationConfig | undefined = pagination
    ? {
        current: pagination.page,
        pageSize: pagination.pageSize,
        total: pagination.total,
        showSizeChanger: true,
        showQuickJumper: true,
        showTotal: (total) => `共 ${total} 条`,
        onChange: (page, pageSize) => pagination.onChange(page, pageSize),
      }
    : undefined;

  return (
    <Card>
      {(search || toolbar) && (
        <div
          style={{
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
            marginBottom: 16,
          }}
        >
          {search ? (
            <Input
              placeholder={search.placeholder}
              prefix={<SearchOutlined />}
              allowClear
              value={search.value}
              onChange={(e) => search.onSearch(e.target.value)}
              onPressEnter={() => search.onSearch(search.value || '')}
              style={{ width: 240 }}
            />
          ) : (
            <div />
          )}
          {toolbar && <Space>{toolbar}</Space>}
        </div>
      )}
      <Table<T>
        columns={columns}
        dataSource={dataSource}
        rowKey={rowKey}
        loading={loading}
        pagination={paginationConfig}
        scroll={scroll}
      />
    </Card>
  );
}

export default DataTable;
