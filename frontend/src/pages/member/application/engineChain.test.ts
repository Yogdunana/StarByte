import { describe, expect, it } from 'vitest';
import type { ApplicationProgress, MemberApplication } from '@/types/api';
import {
  currentAllowsTransfer,
  currentStepIndex,
  engineReviewClosed,
  fallbackSteps,
  stepStatus,
} from './engineChain';

describe('engineReviewClosed', () => {
  it('hides actions for approved, rejected, and supplement', () => {
    expect(engineReviewClosed(0)).toBe(false);
    expect(engineReviewClosed(1)).toBe(false);
    expect(engineReviewClosed(3)).toBe(true);
    expect(engineReviewClosed(4)).toBe(true);
    expect(engineReviewClosed(5)).toBe(true);
  });
});

describe('progress helpers', () => {
  it('maps step states and current index', () => {
    const steps = [
      { id: 'start', label: '提交申请', type: 'start', state: 'done' },
      { id: 'officer', label: '干事审批', type: 'approval', state: 'skipped' },
      { id: 'minister', label: '部长审批', type: 'approval', state: 'current', allow_transfer: true },
      { id: 'president', label: '社长审批', type: 'approval', state: 'pending' },
    ];
    expect(currentStepIndex(steps)).toBe(2);
    expect(stepStatus(steps[0], false)).toBe('finish');
    expect(stepStatus(steps[2], false)).toBe('process');
    expect(stepStatus(steps[2], true)).toBe('error');
    expect(currentAllowsTransfer({ steps } as ApplicationProgress)).toBe(true);
  });

  it('builds fallback steps from application status', () => {
    const record = { status: 1, current_stage: '社长审批' } as MemberApplication;
    const steps = fallbackSteps(record);
    expect(steps).toHaveLength(5);
    expect(steps[3].state).toBe('current');
    expect(steps[4].state).toBe('pending');
  });
});
