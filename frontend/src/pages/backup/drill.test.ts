import { describe, expect, it } from 'vitest';
import { drillIsPending, type BackupDrillResult } from '@/api/backup';

function result(partial: Partial<BackupDrillResult>): BackupDrillResult {
  return {
    id: 'id',
    filename: 'a.dump.gz',
    ready: false,
    restored: false,
    target_host: 'postgres',
    target_port: 5432,
    target_dbname: 'starbyte_drill',
    ...partial,
  };
}

describe('drillIsPending', () => {
  it('treats queued and running as pending', () => {
    expect(drillIsPending(result({ queued: true, status: 'queued' }))).toBe(true);
    expect(drillIsPending(result({ queued: true, status: 'running' }))).toBe(true);
  });

  it('stops on restored or failed', () => {
    expect(drillIsPending(result({ queued: true, restored: true, status: 'restored' }))).toBe(false);
    expect(drillIsPending(result({ queued: false, status: 'failed', error: 'boom' }))).toBe(false);
    expect(drillIsPending(result({ status: 'restored', restored: true }))).toBe(false);
  });
});
