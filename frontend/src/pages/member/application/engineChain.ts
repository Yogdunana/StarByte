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

export function fallbackSteps(record: MemberApplication): ApplicationProgressStep[] {
  const current =
    record.status === 3 ? 4 : record.status === 1 ? 3 : record.current_stage === '干事审批' ? 1 : 2;
  const labels = [
    { id: 'start', label: tx('提交申请'), type: 'start' },
    { id: 'officer', label: tx('干事审批'), type: 'approval' },
    { id: 'minister', label: tx('部长审批'), type: 'approval' },
    { id: 'president', label: tx('社长审批'), type: 'approval' },
    { id: 'end', label: tx('结束'), type: 'end' },
  ];
  return labels.map((item, index) => ({
    ...item,
    state:
      record.status === 4 && index === current
        ? 'current'
        : index < current
          ? 'done'
          : index === current
            ? 'current'
            : 'pending',
    allow_transfer: item.type === 'approval',
  }));
}

export function currentAllowsTransfer(progress: ApplicationProgress | null): boolean {
  if (!progress) return true;
  return progress.steps.some((step) => step.state === 'current' && step.allow_transfer);
}
