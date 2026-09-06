import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import EmptyState from './EmptyState';

describe('EmptyState', () => {
  it('renders title and description', () => {
    render(<EmptyState title="暂无数据" description="请先创建" />);
    expect(screen.getByText('暂无数据')).toBeInTheDocument();
    expect(screen.getByText('请先创建')).toBeInTheDocument();
  });
});
