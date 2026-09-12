import { useEffect, useState } from 'react';
import { Button, Form, Input, Select, Space, Steps, message } from 'antd';
import { useTranslation } from 'react-i18next';
import { Link } from 'react-router-dom';
import {
  approveApplication,
  getApplicationProgress,
  getApplicationTransferCandidates,
  rejectApplication,
  transferApplication,
} from '@/api/member';
import type { ApplicationProgress, MemberApplication, TransferCandidate } from '@/types/api';
import {
  currentAllowsTransfer,
  currentStepIndex,
  engineReviewClosed,
  fallbackSteps,
  stepStatus,
} from './engineChain';
import styles from './AdmissionPanel.module.css';

interface Props {
  record: MemberApplication;
  editable: boolean;
  onChanged: () => void;
}

export default function EngineChainPanel({ record, editable, onChanged }: Props) {
  const { t } = useTranslation();
  const [form] = Form.useForm<{ comment?: string; target?: string }>();
  const [busy, setBusy] = useState(false);
  const [progress, setProgress] = useState<ApplicationProgress | null>(null);
  const [candidates, setCandidates] = useState<TransferCandidate[]>([]);
  const closed = engineReviewClosed(record.status);
  const steps = progress?.steps?.length ? progress.steps : fallbackSteps(record);
  const rejected = record.status === 4 || Boolean(progress?.terminated);

  useEffect(() => {
    let active = true;
    getApplicationProgress(record.id)
      .then((row) => {
        if (active) setProgress(row);
      })
      .catch(() => {
        if (active) setProgress(null);
      });
    return () => {
      active = false;
    };
  }, [record.id, record.status, record.current_stage, record.updated_at]);

  useEffect(() => {
    if (!editable || closed) return;
    let active = true;
    getApplicationTransferCandidates(record.id)
      .then((rows) => {
        if (active) setCandidates(rows);
      })
      .catch(() => {
        if (active) setCandidates([]);
      });
    return () => {
      active = false;
    };
  }, [editable, closed, record.id, record.updated_at]);

  const run = async (fn: () => Promise<unknown>, ok: string) => {
    setBusy(true);
    try {
      await fn();
      message.success(ok);
      onChanged();
    } finally {
      setBusy(false);
    }
  };

  return (
    <section className={styles.panel} aria-label={t('member.engine.title')}>
      <div className={styles.heading}>
        <h3>{t('member.engine.title')}</h3>
      </div>
      <Steps
        direction="vertical"
        size="small"
        current={currentStepIndex(steps)}
        status={rejected ? 'error' : record.status === 3 || progress?.completed ? 'finish' : 'process'}
        items={steps.map((step) => ({
          title: step.label,
          description:
            step.state === 'skipped'
              ? t('member.engine.skipped')
              : step.approval_type === 'all'
                ? t('member.engine.countersign')
                : step.approval_type === 'any'
                  ? t('member.engine.orsign')
                  : undefined,
          status: stepStatus(step, rejected),
        }))}
      />
      {record.flow_instance_id && (
        <p>
          <Link to={`/workflow/instances?instance_id=${encodeURIComponent(record.flow_instance_id)}`}>
            {t('member.engine.viewFlow')}
          </Link>
          {' · '}
          <Link to="/workflow/designer">{t('member.engine.editTemplate')}</Link>
        </p>
      )}
      <p className={styles.hint}>{t('member.engine.hint')}</p>
      {record.status === 5 && <p className={styles.hint}>{t('member.engine.supplementHint')}</p>}
      {editable && !closed && (
        <Form form={form} layout="vertical" className={styles.form}>
          <Form.Item name="comment" label={t('member.engine.comment')} rules={[{ max: 1000 }]}>
            <Input.TextArea rows={3} maxLength={1000} />
          </Form.Item>
          {currentAllowsTransfer(progress) && (
            <Form.Item name="target" label={t('member.engine.transferTo')}>
              <Select
                allowClear
                showSearch
                optionFilterProp="label"
                placeholder={t('member.engine.transferPlaceholder')}
                options={candidates.map((item) => ({
                  value: item.id,
                  label: item.department_name ? `${item.name}（${item.department_name}）` : item.name,
                }))}
              />
            </Form.Item>
          )}
          <Space wrap>
            <Button
              type="primary"
              loading={busy}
              onClick={() => {
                void run(async () => {
                  const values = await form.validateFields();
                  await approveApplication(record.id, values.comment || '');
                }, t('member.engine.approved'));
              }}
            >
              {t('member.engine.approve')}
            </Button>
            <Button
              danger
              loading={busy}
              onClick={() => {
                void (async () => {
                  const values = await form.validateFields();
                  const comment = (values.comment || '').trim();
                  if (!comment) {
                    form.setFields([{ name: 'comment', errors: [t('member.engine.commentRequired')] }]);
                    return;
                  }
                  await run(() => rejectApplication(record.id, comment), t('member.engine.rejected'));
                })();
              }}
            >
              {t('member.engine.reject')}
            </Button>
            {currentAllowsTransfer(progress) && (
              <Button
                loading={busy}
                onClick={() => {
                  void (async () => {
                    const values = await form.validateFields();
                    const target = (values.target || '').trim();
                    if (!target) {
                      form.setFields([{ name: 'target', errors: [t('member.engine.transferRequired')] }]);
                      return;
                    }
                    await run(
                      () => transferApplication(record.id, target, values.comment || ''),
                      t('member.engine.transferred'),
                    );
                  })();
                }}
              >
                {t('member.engine.transfer')}
              </Button>
            )}
          </Space>
        </Form>
      )}
    </section>
  );
}
