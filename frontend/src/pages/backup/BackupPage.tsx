import { tx } from '@/i18n/text';
import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import {
  Alert, Button, Card, Form, Input, InputNumber, Modal, Space, Statistic,
  Switch, Table, Tag, Typography, message,
} from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { CloudServerOutlined, DeleteOutlined, ReloadOutlined, SafetyCertificateOutlined } from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import {
  createBackup, deleteBackup, drillIsPending, drillRestoreBackup, getBackupPolicy, getBackups,
  getBackupStorage, getDrillRestore, previewBackup, restoreBackup, updateBackupPolicy,
  type BackupPolicy, type BackupPreview, type BackupRecord, type BackupStorageStats,
} from '@/api/backup';
import { usePermissions } from '@/hooks/usePermission';
import { isCanceledError } from '@/api/error';
import { canRetryRestore, formatBytes, isActiveStatus } from './format';
import { applyPreviewIfCurrent, canContinueRestore, createPreviewSession, failedPreview, previewErrorMessage } from './preview';
import './backup.css';

const statusColor: Record<number, string> = {
  0: 'default',
  1: 'processing',
  2: 'success',
  3: 'error',
  4: 'warning',
  5: 'blue',
  6: 'error',
};

const BackupPage: React.FC = () => {
  const { t } = useTranslation();
  const [canCreate, canDelete, canRestore, canManage] = usePermissions([
    'backup:create', 'backup:delete', 'backup:restore', 'backup:manage',
  ]);
  const [list, setList] = useState<BackupRecord[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(false);
  const [creating, setCreating] = useState(false);
  const [storage, setStorage] = useState<BackupStorageStats>();
  const [form] = Form.useForm<BackupPolicy>();
  const [restoreRow, setRestoreRow] = useState<BackupRecord | null>(null);
  const [restoreText, setRestoreText] = useState('');
  const [previewRow, setPreviewRow] = useState<BackupRecord | null>(null);
  const [preview, setPreview] = useState<BackupPreview | null>(null);
  const [previewing, setPreviewing] = useState(false);
  const previewSession = useRef(createPreviewSession()).current;
  const [drillRow, setDrillRow] = useState<BackupRecord | null>(null);
  const [drillText, setDrillText] = useState('');
  const [drillDB, setDrillDB] = useState('starbyte_drill');
  const [drillDSN, setDrillDSN] = useState('');
  const [drillPassword, setDrillPassword] = useState('');
  const [drilling, setDrilling] = useState(false);

  const load = useCallback(async (p = page) => {
    setLoading(true);
    try {
      const [rows, stats, pol] = await Promise.all([
        getBackups({ page: p, page_size: 10 }),
        getBackupStorage(),
        getBackupPolicy(),
      ]);
      setList(rows?.list || []);
      setTotal(rows?.total || 0);
      setStorage(stats);
      form.setFieldsValue(pol);
    } finally {
      setLoading(false);
    }
  }, [form, page]);

  useEffect(() => { void load(); }, [load]);

  const inflight = useMemo(() => list.some((row) => isActiveStatus(row.status)), [list]);
  useEffect(() => {
    if (!inflight) return undefined;
    const id = window.setInterval(() => { void load(); }, 3000);
    return () => window.clearInterval(id);
  }, [inflight, load]);

  const onCreate = async () => {
    setCreating(true);
    try {
      await createBackup();
      message.success(t('backup.created'));
      void load();
    } finally {
      setCreating(false);
    }
  };

  const onSavePolicy = async () => {
    const values = await form.validateFields();
    await updateBackupPolicy(values);
    message.success(t('backup.policySaved'));
    void load();
  };

  const onDelete = (row: BackupRecord) => {
    Modal.confirm({
      title: t('backup.deleteConfirm'),
      okButtonProps: { danger: true },
      onOk: async () => {
        await deleteBackup(row.id);
        message.success(t('backup.deleted'));
        void load();
      },
    });
  };

  const closePreview = () => {
    previewSession.invalidate();
    setPreviewRow(null);
    setPreview(null);
    setPreviewing(false);
  };

  const onPreview = async (row: BackupRecord) => {
    const ticket = previewSession.begin(row.id);
    setPreviewRow(row);
    setPreview(null);
    setPreviewing(true);
    try {
      const next = applyPreviewIfCurrent(previewSession, ticket, await previewBackup(row.id, ticket.signal));
      if (!next) return;
      setPreview(next);
    } catch (err) {
      if (isCanceledError(err) || !previewSession.isCurrent(ticket.gen)) return;
      setPreview(failedPreview(row, previewErrorMessage(err)));
    } finally {
      if (previewSession.isCurrent(ticket.gen)) setPreviewing(false);
    }
  };

  useEffect(() => () => previewSession.invalidate(), [previewSession]);

  const onRestore = async () => {
    if (!restoreRow) return;
    if (restoreText.trim() !== t('backup.restoreToken')) return;
    await restoreBackup(restoreRow.id, t('backup.restoreToken'));
    message.success(t('backup.restoredQueued'));
    setRestoreRow(null);
    setRestoreText('');
    void load();
  };

  const onDrill = async () => {
    if (!drillRow) return;
    if (drillText.trim() !== t('backup.drillToken')) return;
    if (!drillDB.trim()) return;
    setDrilling(true);
    try {
      let out = await drillRestoreBackup(drillRow.id, {
        target_dbname: drillDB.trim(),
        target_dsn: drillDSN.trim() || undefined,
        target_password: drillPassword || undefined,
        confirmation: t('backup.drillToken'),
      });
      if (drillIsPending(out)) {
        message.info(t('backup.drillQueued'));
        const deadline = Date.now() + 30 * 60 * 1000;
        while (Date.now() < deadline && drillIsPending(out)) {
          await new Promise((resolve) => setTimeout(resolve, 2000));
          out = await getDrillRestore(drillRow.id);
        }
      }
      if (out.restored) {
        message.success(t('backup.drillOk', { db: out.target_dbname }));
        setDrillRow(null);
        setDrillText('');
        setDrillPassword('');
      } else if (drillIsPending(out)) {
        message.warning(t('backup.drillStillRunning', { db: out.target_dbname }));
      } else {
        message.error(out.error || t('backup.drillFail'));
      }
    } finally {
      setDrilling(false);
    }
  };

  const flag = (ok: boolean) => (ok ? <Tag color="success">OK</Tag> : <Tag color="error">{tx('失败')}</Tag>);

  const columns: ColumnsType<BackupRecord> = [
    {
      title: t('backup.filename'),
      dataIndex: 'filename',
      render: (v: string, row) => (
        <Space size={6} wrap>
          <span className="backup-mono">{v || '-'}</span>
          <Tag color={row.encrypted ? 'geekblue' : 'default'}>
            {row.encrypted ? t('backup.encrypted') : t('backup.unencrypted')}
          </Tag>
        </Space>
      ),
    },
    {
      title: t('backup.statusLabel'),
      dataIndex: 'status',
      width: 110,
      render: (v: number) => <Tag color={statusColor[v]}>{t(`backup.status.${v}`)}</Tag>,
    },
    {
      title: t('backup.triggerLabel'),
      dataIndex: 'trigger_source',
      width: 100,
      render: (v: string) => t(`backup.trigger.${v}`, { defaultValue: v }),
    },
    {
      title: t('backup.size'),
      dataIndex: 'size_bytes',
      width: 110,
      render: (v: number) => formatBytes(v),
    },
    {
      title: t('backup.checksum'),
      dataIndex: 'checksum_sha256',
      ellipsis: true,
      render: (v: string) => <span className="backup-mono">{v ? v.slice(0, 12) : '-'}</span>,
    },
    {
      title: t('backup.finished'),
      dataIndex: 'finished_at',
      width: 180,
      render: (v?: string) => v || '-',
    },
    {
      title: t('common.actions'),
      width: 260,
      render: (_, row) => (
        <Space wrap size="small">
          {canRetryRestore(row.status) && (
            <Button type="link" size="small" icon={<SafetyCertificateOutlined />} onClick={() => { void onPreview(row); }}>
              {t('backup.preview')}
            </Button>
          )}
          {canRestore && canRetryRestore(row.status) && (
            <Button type="link" size="small" danger onClick={() => { setRestoreText(''); setRestoreRow(row); }}>
              {t('backup.restore')}
            </Button>
          )}
          {canDelete && !isActiveStatus(row.status) && (
            <Button type="link" size="small" danger icon={<DeleteOutlined />} onClick={() => onDelete(row)}>
              {t('backup.delete')}
            </Button>
          )}
        </Space>
      ),
    },
  ];

  return (
    <div className="backup-page">
      <div className="backup-hero">
        <div>
          <Typography.Title level={3} style={{ margin: 0 }}>{t('backup.title')}</Typography.Title>
          <p className="backup-meta">{t('backup.hint')}</p>
        </div>
        <Space wrap>
          {canCreate && (
            <Button size="large" type="primary" icon={<CloudServerOutlined />} loading={creating} onClick={() => void onCreate()}>
              {t('backup.create')}
            </Button>
          )}
          <Button size="large" icon={<ReloadOutlined />} onClick={() => void load()}>
            {t('backup.refresh')}
          </Button>
        </Space>
      </div>

      <Alert className="backup-wal" type="warning" showIcon message={t('backup.walGap')} />

      <div className="backup-stats">
        <Card><Statistic title={t('backup.storageCount')} value={storage?.count ?? 0} /></Card>
        <Card><Statistic title={t('backup.storageSize')} value={formatBytes(storage?.size_bytes)} /></Card>
        <Card><Statistic title={t('backup.storage')} value={storage?.prefix || 'backups'} /></Card>
        <Card>
          <Statistic
            title={t('backup.compression')}
            value={storage?.compression || 'gzip'}
          />
          <div className="backup-enc-flag">
            {storage?.encryption_enabled ? t('backup.encryptionOn') : t('backup.encryptionOff')}
          </div>
        </Card>
      </div>

      <Card title={t('backup.policy')} style={{ marginBottom: 16 }}>
        <Form form={form} layout="inline" disabled={!canManage} onFinish={() => void onSavePolicy()}>
          <Form.Item name="enabled" label={t('backup.policyEnabled')} valuePropName="checked">
            <Switch />
          </Form.Item>
          <Form.Item name="retention_days" label={t('backup.retentionDays')} rules={[{ required: true }]}>
            <InputNumber min={1} max={3650} />
          </Form.Item>
          <Form.Item name="cron_expr" label={t('backup.cron')} rules={[{ required: true }]}>
            <Input style={{ width: 200 }} placeholder="0 30 2 * * *" />
          </Form.Item>
          <Form.Item name="timezone" label={t('backup.timezone')}>
            <Input style={{ width: 160 }} />
          </Form.Item>
          {canManage && <Button type="primary" htmlType="submit">{t('backup.savePolicy')}</Button>}
        </Form>
      </Card>

      <Card>
        <Table
          rowKey="id"
          loading={loading}
          columns={columns}
          dataSource={list}
          locale={{ emptyText: t('backup.empty') }}
          expandable={{
            expandedRowRender: (row) => (
              <div>
                <p className="backup-mono">{t('backup.checksum')}: {row.checksum_sha256 || '-'}</p>
                {row.error_message ? <Alert type="error" showIcon message={row.error_message} /> : null}
              </div>
            ),
          }}
          pagination={{ current: page, pageSize: 10, total, onChange: (p) => setPage(p) }}
        />
        <Alert className="backup-cli" type="info" showIcon message={t('backup.cliNote')} />
      </Card>

      <Modal
        open={!!previewRow}
        title={t('backup.previewTitle')}
        confirmLoading={previewing}
        okText={t('backup.previewContinue')}
        okButtonProps={{ disabled: !canContinueRestore(previewRow, preview) || !canRestore }}
        onCancel={closePreview}
        onOk={() => {
          if (!canContinueRestore(previewRow, preview) || !previewRow) return;
          setRestoreText('');
          setRestoreRow(previewRow);
          closePreview();
        }}
      >
        {previewing && <p>{t('backup.preview')}…</p>}
        {preview && (
          <div>
            <Alert
              type={preview.ready ? 'success' : 'error'}
              showIcon
              style={{ marginBottom: 16 }}
              message={preview.ready ? t('backup.previewReady') : t('backup.previewNotReady')}
            />
            <p>{t('backup.previewChecksum')}: {flag(preview.checksum_ok)}</p>
            <p>{t('backup.previewDecrypt')}: {preview.encrypted ? flag(preview.decrypt_ok) : t('backup.unencrypted')}</p>
            <p>{t('backup.previewGzip')}: {flag(preview.gzip_ok)}</p>
            <p>{t('backup.previewTOC')}: {flag(preview.toc_valid)}</p>
            {preview.error ? <Alert type="error" showIcon message={preview.error} /> : null}
            {preview.toc ? <pre className="backup-toc">{preview.toc}</pre> : null}
            {canRestore && canContinueRestore(previewRow, preview) && (
              <Button
                style={{ marginTop: 12 }}
                onClick={() => {
                  if (!previewRow) return;
                  setDrillDB('starbyte_drill');
                  setDrillDSN('');
                  setDrillPassword('');
                  setDrillText('');
                  setDrillRow(previewRow);
                  closePreview();
                }}
              >
                {t('backup.drill')}
              </Button>
            )}
          </div>
        )}
      </Modal>

      <Modal
        open={!!drillRow}
        title={t('backup.drillTitle')}
        confirmLoading={drilling}
        okText={t('backup.drillRun')}
        okButtonProps={{ disabled: drillText.trim() !== t('backup.drillToken') || !drillDB.trim() }}
        onCancel={() => { setDrillRow(null); setDrillText(''); setDrillPassword(''); }}
        onOk={() => void onDrill()}
      >
        <Alert type="info" showIcon style={{ marginBottom: 16 }} message={t('backup.drillWarn')} />
        <p>{t('backup.drillDBLabel')}</p>
        <Input value={drillDB} onChange={(e) => setDrillDB(e.target.value)} style={{ marginBottom: 12 }} />
        <p>{t('backup.drillDSNLabel')}</p>
        <Input
          value={drillDSN}
          onChange={(e) => setDrillDSN(e.target.value)}
          placeholder="postgres://user:pass@host:5432/starbyte_drill"
          style={{ marginBottom: 12 }}
        />
        <p>{t('backup.drillPasswordLabel')}</p>
        <Input.Password
          value={drillPassword}
          onChange={(e) => setDrillPassword(e.target.value)}
          style={{ marginBottom: 12 }}
        />
        <p>{t('backup.drillConfirmLabel')}</p>
        <Input
          value={drillText}
          onChange={(e) => setDrillText(e.target.value)}
          placeholder={t('backup.drillToken')}
        />
      </Modal>

      <Modal
        open={!!restoreRow}
        title={t('backup.restoreTitle')}
        okButtonProps={{ danger: true, disabled: restoreText.trim() !== t('backup.restoreToken') }}
        onCancel={() => { setRestoreRow(null); setRestoreText(''); }}
        onOk={() => void onRestore()}
      >
        <Alert type="error" showIcon style={{ marginBottom: 16 }} message={t('backup.restoreWarn')} />
        <p>{t('backup.restoreConfirmLabel')}</p>
        <Input
          value={restoreText}
          onChange={(e) => setRestoreText(e.target.value)}
          placeholder={t('backup.restoreToken')}
        />
      </Modal>
    </div>
  );
};

export default BackupPage;
