import { describe, expect, it } from 'vitest';
import en from '../../locales/en-US.json';
import zh from '../../locales/zh-CN.json';

const keys = [
  'title', 'hint', 'refresh', 'create', 'delete', 'restore', 'restoreTitle',
  'restoreWarn', 'restoreConfirmLabel', 'restoreToken', 'policy', 'policyEnabled',
  'retentionDays', 'cron', 'timezone', 'savePolicy', 'storage', 'storageCount',
  'storageSize', 'cliNote', 'deleteConfirm', 'created', 'deleted', 'restoredQueued',
  'policySaved', 'checksum', 'size', 'filename', 'empty', 'statusLabel', 'triggerLabel', 'finished',
  'preview', 'previewTitle', 'previewReady', 'previewNotReady', 'previewChecksum', 'previewDecrypt',
  'previewGzip', 'previewTOC', 'previewContinue', 'drill', 'drillTitle', 'drillWarn',
  'drillDBLabel', 'drillDSNLabel', 'drillPasswordLabel', 'drillConfirmLabel', 'drillToken', 'drillRun', 'drillQueued', 'drillStillRunning', 'drillOk', 'drillFail',
  'encrypted', 'unencrypted', 'encryptionOn',
  'encryptionOff', 'compression', 'walGap',
  'status.0', 'status.1', 'status.2', 'status.3', 'status.4', 'status.5', 'status.6',
  'trigger.manual', 'trigger.scheduled',
] as const;

function pick(obj: Record<string, unknown>, path: string): unknown {
  return path.split('.').reduce<unknown>((acc, key) => {
    if (acc && typeof acc === 'object') {
      return (acc as Record<string, unknown>)[key];
    }
    return undefined;
  }, obj);
}

describe('backup i18n', () => {
  it('has matching zh-CN and en-US keys', () => {
    keys.forEach((key) => {
      expect(pick(zh.backup, key), `zh missing backup.${key}`).toBeTruthy();
      expect(pick(en.backup, key), `en missing backup.${key}`).toBeTruthy();
    });
    expect(zh.menu['/backup']).toBeTruthy();
    expect(en.menu['/backup']).toBeTruthy();
  });
});
