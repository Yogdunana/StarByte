import type { TFunction } from 'i18next';
import type { LeaveStatus } from '@/api/leave';

export function leaveStatusMap(t: TFunction): Record<string, { color: string; text: string }> {
  return {
    pending: { color: 'processing', text: t('leave.status.pending') },
    approved: { color: 'success', text: t('leave.status.approved') },
    rejected: { color: 'error', text: t('leave.status.rejected') },
  };
}

export const LeaveStatuses: LeaveStatus[] = ['pending', 'approved', 'rejected'];

export function leaveTypeLabel(t: TFunction, code: string, fallback: string): string {
  return t(`leave.type.${code}`, { defaultValue: fallback });
}

export function leaveStageLabel(t: TFunction, stage?: string): string {
  if (!stage) return '';
  return t(`leave.stage.${stage}`, { defaultValue: stage });
}
