import { renderHook } from '@testing-library/react';
import { expect, it, vi } from 'vitest';
import { useSelector } from 'react-redux';
import { useHasRole } from './usePermission';

vi.mock('react-redux', () => ({ useSelector: vi.fn() }));

it('does not treat technical administration as an association office', () => {
  vi.mocked(useSelector).mockReturnValue(['super_admin']);
  expect(renderHook(() => useHasRole('president')).result.current).toBe(false);
});
it('recognizes separately appointed concurrent offices', () => {
  vi.mocked(useSelector).mockReturnValue(['president', 'minister']);
  expect(renderHook(() => useHasRole('president')).result.current).toBe(true);
  expect(renderHook(() => useHasRole('minister')).result.current).toBe(true);
});
