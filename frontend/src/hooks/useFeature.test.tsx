import { useState } from 'react';
import { act, render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { evaluateMyFeatures } from '@/api/feature';
import { FeatureProvider, mapEvaluate, useFeatureFlags } from './useFeature';

vi.mock('@/api/feature', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@/api/feature')>();
  return { ...actual, evaluateMyFeatures: vi.fn() };
});

beforeEach(() => {
  vi.mocked(evaluateMyFeatures).mockReset();
  vi.mocked(evaluateMyFeatures).mockResolvedValue({
    'cms.public': { key: 'cms.public', enabled: true, reason: 'boolean' },
  });
});

describe('mapEvaluate', () => {
  it('maps enabled flags and reasons', () => {
    const got = mapEvaluate({
      'cms.public': { key: 'cms.public', enabled: true, reason: 'allowlist_hit' },
      'membership.portal': { key: 'membership.portal', enabled: false, reason: 'disabled' },
    });
    expect(got.flags['cms.public']).toBe(true);
    expect(got.flags['membership.portal']).toBe(false);
    expect(got.reasons['cms.public']).toBe('allowlist_hit');
  });

  it('fails closed on empty payload', () => {
    expect(mapEvaluate(null).flags).toEqual({});
    expect(mapEvaluate(undefined).flags).toEqual({});
  });
});

function Probe() {
  const { loading, flags } = useFeatureFlags();
  if (loading) return <span>loading</span>;
  return <span>{flags['cms.public'] ? 'ready-on' : 'ready-off'}</span>;
}

function RerenderHarness({ inlineKeys }: { inlineKeys?: boolean }) {
  const [, setTick] = useState(0);
  return (
    <div>
      <button type="button" onClick={() => setTick((n) => n + 1)}>rerender</button>
      <FeatureProvider keys={inlineKeys ? ['cms.public'] : undefined}>
        <Probe />
      </FeatureProvider>
    </div>
  );
}

describe('FeatureProvider', () => {
  it('does not refetch when the parent re-renders with the default keys', async () => {
    const user = userEvent.setup();
    await act(async () => { render(<RerenderHarness />); });
    await waitFor(() => expect(screen.getByText('ready-on')).toBeInTheDocument());
    expect(evaluateMyFeatures).toHaveBeenCalledTimes(1);

    await user.click(screen.getByRole('button', { name: 'rerender' }));
    await user.click(screen.getByRole('button', { name: 'rerender' }));
    await waitFor(() => expect(screen.getByText('ready-on')).toBeInTheDocument());
    expect(evaluateMyFeatures).toHaveBeenCalledTimes(1);
  });

  it('does not refetch when the parent re-renders with an inline keys array', async () => {
    const user = userEvent.setup();
    await act(async () => { render(<RerenderHarness inlineKeys />); });
    await waitFor(() => expect(screen.getByText('ready-on')).toBeInTheDocument());
    expect(evaluateMyFeatures).toHaveBeenCalledTimes(1);

    await user.click(screen.getByRole('button', { name: 'rerender' }));
    await user.click(screen.getByRole('button', { name: 'rerender' }));
    await waitFor(() => expect(screen.getByText('ready-on')).toBeInTheDocument());
    expect(evaluateMyFeatures).toHaveBeenCalledTimes(1);
  });
});
