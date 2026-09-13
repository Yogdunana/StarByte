import { tx } from '@/i18n/text';
import type { ApplicationProgress, ApplicationProgressStep, MemberApplication } from '@/types/api';

export function engineReviewClosed(status: number): boolean {
  return status === 3 || status === 4 || status === 5;
}

export function stepStatus(
  step: ApplicationProgressStep,
  rejected: boolean,
): 'wait' | 'process' | 'finish' | 'error' {
  if (rejected && (step.state === 'current' || step.state === 'done') && step.type === 'approval') {
    return step.state === 'current' ? 'error' : 'finish';
  }
  if (step.state === 'done') return 'finish';
  if (step.state === 'current') return rejected ? 'error' : 'process';
  if (step.state === 'skipped') return 'finish';
  return 'wait';
}

export function currentStepIndex(steps: ApplicationProgressStep[]): number {
  const current = steps.findIndex((step) => step.state === 'current');
  if (current >= 0) return current;
  const lastDone = [...steps]
    .reverse()
    .findIndex((step) => step.state === 'done' || step.state === 'skipped');
  if (lastDone >= 0) return steps.length - 1 - lastDone;
  return 0;
}

export function fallbackSteps(_record: MemberApplication): ApplicationProgressStep[] {
  return [
    {
      id: 'unavailable',
      label: tx('审批进度暂不可用'),
      type: 'approval',
      state: 'pending',
      allow_transfer: false,
    },
  ];
}

export function currentAllowsTransfer(progress: ApplicationProgress | null): boolean {
  if (!progress) return false;
  return progress.steps.some((step) => step.state === 'current' && step.allow_transfer);
}
