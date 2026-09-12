import { useState } from 'react';
import { Button, Form, Input, Space, Steps, message } from 'antd';
import { useTranslation } from 'react-i18next';
import { Link } from 'react-router-dom';
import { approveApplication, rejectApplication } from '@/api/member';
import type { MemberApplication } from '@/types/api';
import { engineReviewClosed } from './engineChain';
import styles from './AdmissionPanel.module.css';

interface Props {
  record: MemberApplication;
  editable: boolean;
  onChanged: () => void;
}

function currentStep(record: MemberApplication): number {
  if (record.status === 3) return 3;
  if (record.status === 1) return 2;
  return 1;
}

export default function EngineChainPanel({ record, editable, onChanged }: Props) {
  const { t } = useTranslation();
  const [form] = Form.useForm<{ comment?: string }>();
  const [busy, setBusy] = useState(false);
  const closed = engineReviewClosed(record.status);
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
        current={currentStep(record)}
        status={record.status === 4 ? 'error' : record.status === 3 ? 'finish' : 'process'}
        items={[
          { title: t('member.engine.steps.apply') },
          { title: t('member.engine.steps.minister') },
          { title: t('member.engine.steps.president') },
          { title: t('member.engine.steps.end') },
        ]}
      />
      {record.flow_instance_id && (
        <p>
          <Link to={`/workflow/instances?instance_id=${encodeURIComponent(record.flow_instance_id)}`}>
            {t('member.engine.viewFlow')}
          </Link>
        </p>
      )}
      <p className={styles.hint}>{t('member.engine.hint')}</p>
      {record.status === 5 && <p className={styles.hint}>{t('member.engine.supplementHint')}</p>}
      {editable && !closed && (
        <Form form={form} layout="vertical" className={styles.form}>
          <Form.Item name="comment" label={t('member.engine.comment')} rules={[{ max: 1000 }]}>
            <Input.TextArea rows={3} maxLength={1000} />
          </Form.Item>
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
          </Space>
        </Form>
      )}
    </section>
  );
}
