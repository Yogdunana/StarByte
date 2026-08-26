import React, { useState, useEffect, useRef, useCallback } from 'react';
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
    debounce?: number; // 防抖毫秒数，默认 300；设为 0 则仅回车搜索
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
  // 内部搜索值（当外部未提供 value 时使用）
  const [searchValue, setSearchValue] = useState(search?.value || '');
  const debounceTimer = useRef<ReturnType<typeof setTimeout> | null>(null);

  // 同步外部 search.value 变更
  useEffect(() => {
    if (search?.value !== undefined) {
      setSearchValue(search.value);
    }
  }, [search?.value]);

  // 清理定时器
  useEffect(() => {
    return () => {
      if (debounceTimer.current) clearTimeout(debounceTimer.current);
    };
  }, []);

  const debouncedSearch = useCallback(
    (value: string) => {
      if (!search) return;
      const debounce = search.debounce ?? 300;
      if (debounce <= 0) return; // debounce=0 时仅回车触发
      if (debounceTimer.current) clearTimeout(debounceTimer.current);
      debounceTimer.current = setTimeout(() => {
        search.onSearch(value);
      }, debounce);
    },
    [search],
  );

  const handleSearchChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const val = e.target.value;
    setSearchValue(val);
    debouncedSearch(val);
  };

  const handleSearchEnter = () => {
    if (!search) return;
    if (debounceTimer.current) clearTimeout(debounceTimer.current);
    search.onSearch(searchValue);
  };

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
              value={searchValue}
              onChange={handleSearchChange}
              onPressEnter={handleSearchEnter}
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
