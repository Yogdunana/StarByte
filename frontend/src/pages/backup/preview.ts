import type { BackupPreview } from '@/api/backup';

export interface PreviewTicket {
  id: string;
  gen: number;
  signal: AbortSignal;
}

export interface PreviewSession {
  begin(id: string): PreviewTicket;
  isCurrent(gen: number): boolean;
  invalidate(): void;
}

/** Serializes backup integrity checks: a newer begin()/invalidate() makes older tickets stale. */
export function createPreviewSession(): PreviewSession {
  let generation = 0;
  let controller: AbortController | null = null;
  return {
    begin(id: string) {
      controller?.abort();
      controller = new AbortController();
      generation += 1;
      return { id, gen: generation, signal: controller.signal };
    },
    isCurrent(gen: number) {
      return gen === generation;
    },
    invalidate() {
      generation += 1;
      controller?.abort();
      controller = null;
    },
  };
}

/** Drop a late response unless it belongs to the still-current ticket and backup id. */
export function applyPreviewIfCurrent(
  session: Pick<PreviewSession, 'isCurrent'>,
  ticket: Pick<PreviewTicket, 'id' | 'gen'>,
  preview: BackupPreview,
): BackupPreview | null {
  if (!session.isCurrent(ticket.gen) || preview.id !== ticket.id) return null;
  return preview;
}

export function canContinueRestore(
  row: { id: string } | null,
  preview: BackupPreview | null,
): boolean {
  return !!row && !!preview && preview.ready && preview.id === row.id;
}
