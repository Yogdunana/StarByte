import { render, screen, fireEvent } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import DataTable from './DataTable';

describe('DataTable', () => {
  const columns = [{ title: '姓名', dataIndex: 'name', key: 'name' }];
  const data = [{ id: '1', name: '张三' }];

  it('renders rows', () => {
    render(
      <DataTable columns={columns} dataSource={data} rowKey="id" />,
    );
    expect(screen.getByText('张三')).toBeInTheDocument();
  });

  it('calls onSearch on Enter when debounce is 0', () => {
    const onSearch = vi.fn();
    render(
      <DataTable
        columns={columns}
        dataSource={data}
        rowKey="id"
        search={{ placeholder: '搜索姓名', onSearch, debounce: 0 }}
      />,
    );
    const input = screen.getByPlaceholderText('搜索姓名');
    fireEvent.change(input, { target: { value: '张' } });
    fireEvent.keyDown(input, { key: 'Enter', code: 'Enter' });
    expect(onSearch).toHaveBeenCalledWith('张');
  });
});
