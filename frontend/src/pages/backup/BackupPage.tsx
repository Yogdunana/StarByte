import React, { useCallback, useEffect, useMemo, useState } from 'react';
import {
  Alert, Button, Card, Form, Input, InputNumber, Modal, Space, Statistic,
  Switch, Table, Tag, Typography, message,
} from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { CloudServerOutlined, DeleteOutlined, ReloadOutlined } from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import {
  createBackup, deleteBackup, getBackupPolicy, getBackups, getBackupStorage,
  restoreBackup, updateBackupPolicy, type BackupPolicy, type BackupRecord, type BackupStorageStats,
} from '@/api/backup';
import { usePermissions } from '@/hooks/usePermission';
import { canRetryRestore, formatBytes, isActiveStatus } from './format';
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

  const onRestore = async () => {
    if (!restoreRow) return;
    if (restoreText.trim() !== t('backup.restoreToken')) return;
    await restoreBackup(restoreRow.id, t('backup.restoreToken'));
    message.success(t('backup.restoredQueued'));
    setRestoreRow(null);
    setRestoreText('');
    void load();
  };

  const columns: ColumnsType<BackupRecord> = [
    {
      title: t('backup.filename'),
      dataIndex: 'filename',
      render: (v: string) => <span className="backup-mono">{v || '-'}</span>,
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
      width: 180,
      render: (_, row) => (
        <Space wrap size="small">
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

      <div className="backup-stats">
        <Card><Statistic title={t('backup.storageCount')} value={storage?.count ?? 0} /></Card>
        <Card><Statistic title={t('backup.storageSize')} value={formatBytes(storage?.size_bytes)} /></Card>
        <Card><Statistic title={t('backup.storage')} value={storage?.prefix || 'backups'} /></Card>
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
