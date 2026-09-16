import { renderHook } from '@testing-library/react';
import { Grid } from 'antd';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { useViewport } from './useViewport';

vi.mock('antd', async (importOriginal) => {
  const actual = await importOriginal<typeof import('antd')>();
  return {
    ...actual,
    Grid: {
      ...actual.Grid,
      useBreakpoint: vi.fn(() => ({})),
    },
  };
});

const breakpoint = Grid.useBreakpoint as unknown as ReturnType<typeof vi.fn>;

describe('useViewport', () => {
  beforeEach(() => {
    breakpoint.mockReset();
  });

  it('treats missing breakpoints as phone', () => {
    breakpoint.mockReturnValue({});
    expect(renderHook(() => useViewport()).result.current).toEqual({
      phone: true,
      tablet: false,
      compact: true,
      desktop: false,
    });
  });

  it('classifies tablet between md and lg', () => {
    breakpoint.mockReturnValue({ xs: true, sm: true, md: true, lg: false });
    expect(renderHook(() => useViewport()).result.current).toEqual({
      phone: false,
      tablet: true,
      compact: true,
      desktop: false,
    });
  });

  it('classifies desktop from lg', () => {
    breakpoint.mockReturnValue({ xs: true, sm: true, md: true, lg: true, xl: true });
    expect(renderHook(() => useViewport()).result.current).toEqual({
      phone: false,
      tablet: false,
      compact: false,
      desktop: true,
    });
  });
});
