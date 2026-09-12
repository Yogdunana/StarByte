import { describe, expect, it } from 'vitest';
import { leaveStatusMap, leaveStageLabel, leaveTypeLabel, LeaveStatuses } from './meta';

const t = ((key: string, opts?: { defaultValue?: string }) => {
  const map: Record<string, string> = {
    'leave.status.pending': '待审批',
    'leave.status.approved': '已批准',
    'leave.status.rejected': '已驳回',
    'leave.type.annual': '年假',
    'leave.stage.minister': '待部长审批',
  };
  return map[key] ?? opts?.defaultValue ?? key;
}) as import('i18next').TFunction;

describe('leave meta', () => {
  it('maps approval statuses', () => {
    const map = leaveStatusMap(t);
    expect(map.pending.text).toBe('待审批');
    expect(map.approved.color).toBe('success');
    expect(map.rejected.color).toBe('error');
    expect(LeaveStatuses).toEqual(['pending', 'approved', 'rejected']);
  });

  it('falls back to type name', () => {
    expect(leaveTypeLabel(t, 'annual', '年假')).toBe('年假');
    expect(leaveTypeLabel(t, 'unknown', '其他')).toBe('其他');
    expect(leaveStageLabel(t, 'minister')).toBe('待部长审批');
    expect(leaveStageLabel(t, '')).toBe('');
  });
});
