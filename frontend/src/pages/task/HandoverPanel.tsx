import { useCallback, useEffect, useRef, useState } from 'react';
import { Alert, Button, Form, Input, Space, Spin, message } from 'antd';
import { useTranslation } from 'react-i18next';
import { decideTaskHandover, decideTaskTransfer, getTaskHandover, getTaskTransfer, type TaskHandover } from '@/api/taskWorkflow';
import styles from './WorkflowPanel.module.css';

interface Props {
  taskId?: string;
  transferId?: string;
  value?: TaskHandover | null;
  onChanged: () => void;
}

export default function HandoverPanel({ taskId, transferId, value, onChanged }: Props) {
  const { t } = useTranslation();
  const [form] = Form.useForm<{ comment?: string }>();
  const [row, setRow] = useState<TaskHandover | null>(value ?? null);
  const [error, setError] = useState(false);
  const [busy, setBusy] = useState(false);
  const sequence = useRef(0);
  const load = useCallback(async () => {
    const current = ++sequence.current;
    try {
      const next = transferId ? await getTaskTransfer(transferId) : taskId ? await getTaskHandover(taskId) : null;
      if (sequence.current === current) { setRow(next); setError(false); }
    } catch {
      if (sequence.current === current) setError(true);
    }
  }, [taskId, transferId]);
  useEffect(() => {
    if (value) { setRow(value); setError(false); return; }
    void load();
  }, [load, value]);

  const sign = async (requirement: string, decision: 'approve' | 'reject') => {
    if (!row || busy) return;
    const comment = (form.getFieldValue('comment') || '').trim();
    if (!comment) {
      form.setFields([{ name: 'comment', errors: [t('task.handover.commentRequired')] }]);
      return;
    }
    setBusy(true);
    try {
      const next = transferId
        ? await decideTaskTransfer(transferId, requirement, decision, comment, row.revision)
        : await decideTaskHandover(row.task_id, requirement, decision, comment, row.revision);
      setRow(next);
      form.resetFields();
      message.success(decision === 'approve' ? t('task.handover.signed') : t('task.handover.rejected'));
      onChanged();
    } finally {
      setBusy(false);
    }
  };

  return (
    <section className={styles.panel} aria-label={t('task.handover.title')}>
      <div className={styles.heading}>
        <h3>{t('task.handover.title')}</h3>
        <Button size="small" disabled={busy} onClick={() => void load()}>{t('task.engine.refresh')}</Button>
      </div>
      {error && <Alert type="error" showIcon message={t('task.handover.loadFailed')} />}
      {!row && !error && <Spin tip={t('common.loading')}><div className={styles.placeholder} /></Spin>}
      {row && (
        <>
          <p className={styles.hint}>{t(`task.handover.kind.${row.kind}`)} · {t(`task.handover.status.${row.status}`)}</p>
          <p className={styles.hint}>{row.from.name} → {row.to.name}</p>
          {row.source_department && <p className={styles.hint}>{row.source_department} → {row.target_department || row.source_department}</p>}
          <p className={styles.hint}>{t('task.handover.reason')}：{row.reason}</p>
          {row.signatures.map((item) => (
            <p key={item.id} className={styles.hint}>
              {item.signer.name} · {t(`task.handover.requirement.${item.requirement}`, item.requirement)}
              {item.waived ? ` · ${t('task.handover.waived')}` : ''} · {item.comment}
            </p>
          ))}
          {row.status === 'pending' && row.can_sign.length > 0 && (
            <Form form={form} layout="vertical" className={styles.form}>
              <Form.Item name="comment" label={t('task.handover.comment')} rules={[{ max: 2000 }]}>
                <Input.TextArea rows={3} maxLength={2000} />
              </Form.Item>
              <Space wrap>
                {row.can_sign.map((requirement) => (
                  <Space key={requirement} wrap>
                    <Button type="primary" loading={busy} onClick={() => void sign(requirement, 'approve')}>
                      {t('task.handover.approve')} · {t(`task.handover.requirement.${requirement}`, requirement)}
                    </Button>
                    <Button danger loading={busy} onClick={() => void sign(requirement, 'reject')}>
                      {t('task.handover.reject')}
                    </Button>
                  </Space>
                ))}
              </Space>
            </Form>
          )}
        </>
      )}
    </section>
  );
}
