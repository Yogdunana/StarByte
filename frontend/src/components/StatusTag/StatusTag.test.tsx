import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import StatusTag from './StatusTag';

describe('StatusTag', () => {
  const mapping = {
    0: { color: 'green', text: '启用' },
    1: { color: 'red', text: '停用' },
  };

  it('renders mapped text', () => {
    render(<StatusTag status={0} mapping={mapping} />);
    expect(screen.getByText('启用')).toBeInTheDocument();
  });

  it('renders fallback for unknown status', () => {
    render(<StatusTag status={9} mapping={mapping} fallbackText="未知状态" />);
    expect(screen.getByText('未知状态')).toBeInTheDocument();
  });
});
