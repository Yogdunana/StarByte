import { describe, expect, it } from 'vitest';
import type { BackupPreview } from '@/api/backup';
import { applyPreviewIfCurrent, canContinueRestore, createPreviewSession } from './preview';

function preview(id: string, ready = true): BackupPreview {
  return {
    id,
    filename: `${id}.dump.gz`,
    size_bytes: 12,
    checksum_ok: ready,
    encrypted: false,
    decrypt_ok: true,
    gzip_ok: ready,
    toc_valid: ready,
    ready,
    compression: 'gzip',
    encryption_configured: false,
  };
}

describe('preview session rejects stale overwrite', () => {
  it('aborts the older check when a newer preview starts', () => {
    const session = createPreviewSession();
    const older = session.begin('backup-a');
    const newer = session.begin('backup-b');
    expect(session.isCurrent(older.gen)).toBe(false);
    expect(session.isCurrent(newer.gen)).toBe(true);
    expect(older.signal.aborted).toBe(true);
    expect(newer.signal.aborted).toBe(false);

    expect(applyPreviewIfCurrent(session, older, preview('backup-a'))).toBeNull();
    expect(applyPreviewIfCurrent(session, newer, preview('backup-b'))?.id).toBe('backup-b');
  });

  it('drops a response whose backup id does not match the ticket', () => {
    const session = createPreviewSession();
    const ticket = session.begin('backup-b');
    expect(applyPreviewIfCurrent(session, ticket, preview('backup-a'))).toBeNull();
    expect(applyPreviewIfCurrent(session, ticket, preview('backup-b'))?.id).toBe('backup-b');
  });

  it('drops late results after cancel', () => {
    const session = createPreviewSession();
    const ticket = session.begin('backup-a');
    session.invalidate();
    expect(ticket.signal.aborted).toBe(true);
    expect(applyPreviewIfCurrent(session, ticket, preview('backup-a'))).toBeNull();
  });
});

describe('canContinueRestore', () => {
  it('requires the ready preview to belong to the selected row', () => {
    expect(canContinueRestore({ id: 'backup-b' }, preview('backup-a'))).toBe(false);
    expect(canContinueRestore({ id: 'backup-b' }, preview('backup-b', false))).toBe(false);
    expect(canContinueRestore(null, preview('backup-b'))).toBe(false);
    expect(canContinueRestore({ id: 'backup-b' }, null)).toBe(false);
    expect(canContinueRestore({ id: 'backup-b' }, preview('backup-b'))).toBe(true);
  });
});
