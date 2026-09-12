import { useCallback, useEffect, useRef, useState } from 'react';
import { Alert, Button, Form, Input, Space, Spin, Steps, message } from 'antd';
import { useTranslation } from 'react-i18next';
import { Link } from 'react-router-dom';
import { actTaskWorkflow, getTaskWorkflow, requestTaskHandover, type TaskWorkflow } from '@/api/taskWorkflow';
import { TASK_ENGINE_STAGES, taskEngineClosed, taskEngineStatus, taskEngineStep } from './engineChain';
import HandoverPanel from './HandoverPanel';
import UserPicker from './UserPicker';
import styles from './WorkflowPanel.module.css';

interface Props { taskId: string; onChanged: () => void }

function assignmentLabel(mode: string): string {
  if (mode === 'department' || mode === 'role' || mode === 'round_robin') return mode;
  return 'manual';
}

export default function WorkflowPanel({ taskId, onChanged }: Props) {
  const { t } = useTranslation();
  const [form] = Form.useForm<{ comment?: string }>();
  const [delegate] = Form.useForm<{ target?: string; reason?: string }>();
  const [flow, setFlow] = useState<TaskWorkflow | null>(null);
  const [error, setError] = useState(false);
  const [busy, setBusy] = useState(false);
  const sequence = useRef(0);
  const invalidate = useCallback(() => { sequence.current++; }, []);
  const load = useCallback(async () => {
    const current = ++sequence.current;
    try {
      const result = await getTaskWorkflow(taskId);
      if (sequence.current === current) { setFlow(result); setError(false); }
    } catch {
      if (sequence.current === current) setError(true);
    }
  }, [taskId]);
  useEffect(() => { setFlow(null); form.resetFields(); delegate.resetFields(); void load(); return invalidate; }, [load, invalidate, form, delegate]);

  const run = async (action: 'start' | 'pause' | 'resume' | 'submit' | 'approve' | 'return' | 'reject' | 'claim', ok: string, requireComment: boolean) => {
    if (!flow || busy) return;
    const comment = (form.getFieldValue('comment') || '').trim();
    if (requireComment && !comment) {
      form.setFields([{ name: 'comment', errors: [t('task.engine.commentRequired')] }]);
      return;
    }
    setBusy(true);
    try {
      const result = await actTaskWorkflow(taskId, action, comment || t('task.engine.updated'), flow.revision);
      setFlow(result);
      form.resetFields();
      message.success(ok);
      onChanged();
    } finally {
      setBusy(false);
    }
  };

  const submitDelegate = async () => {
    if (!flow || busy) return;
    const values = await delegate.validateFields();
    setBusy(true);
    try {
      await requestTaskHandover(taskId, values.target!, values.reason!.trim(), flow.revision);
      delegate.resetFields();
      message.success(t('task.handover.submitted'));
      await load();
      onChanged();
    } finally {
      setBusy(false);
    }
  };

  const closed = flow ? taskEngineClosed(flow.stage) : false;
  const person = (stage: string) => {
    if (!flow) return undefined;
    if (stage === 'assignment') return flow.creator.name;
    if (stage === 'execution') return flow.assignee?.name || t('task.engine.waitingAssignee');
    if (stage === 'review') return flow.reviewer.name;
    if (stage === 'acceptance') return flow.acceptor.name;
    return undefined;
  };

  return (
    <section className={styles.panel} aria-label={t('task.engine.title')}>
      <div className={styles.heading}>
        <h3>{t('task.engine.title')}</h3>
        <Button size="small" disabled={busy} onClick={() => void load()}>{t('task.engine.refresh')}</Button>
      </div>
      {error && <Alert type="error" showIcon message={t('task.engine.loadFailed')} action={<Button onClick={() => void load()}>{t('task.engine.refresh')}</Button>} />}
      {!flow && !error && <Spin tip={t('common.loading')}><div className={styles.placeholder} /></Spin>}
      {flow && (
        <>
          <Steps
            direction="vertical"
            size="small"
            current={taskEngineStep(flow.stage)}
            status={taskEngineStatus(flow.stage)}
            items={TASK_ENGINE_STAGES.map((stage) => ({ title: t(`task.engine.steps.${stage}`), description: person(stage) }))}
          />
          {flow.instance_id && (
            <p>
              <Link to={`/workflow/instances?instance_id=${encodeURIComponent(flow.instance_id)}`}>
                {t('task.engine.viewFlow')}
              </Link>
            </p>
          )}
          <p className={styles.hint}>{t('task.engine.hint')}</p>
          <p className={styles.hint}>{t('task.engine.assignmentMode')}：{t(`task.engine.assignment.${assignmentLabel(flow.assignment_mode)}`)}</p>
          <p className={styles.hint}>{t('task.engine.submission')}：{flow.submission || t('task.engine.submissionEmpty')}</p>
          {closed && <p className={styles.hint}>{t('task.engine.closed')}</p>}
          {!closed && (flow.can_start || flow.can_pause || flow.can_resume || flow.can_submit || flow.can_approve || flow.can_claim) && (
            <Form form={form} layout="vertical" className={styles.form}>
              <Form.Item name="comment" label={t('task.engine.comment')} rules={[{ max: 5000 }]}>
                <Input.TextArea rows={3} maxLength={5000} />
              </Form.Item>
              <Space wrap>
                {flow.can_start && <Button type="primary" loading={busy} onClick={() => void run('start', t('task.engine.updated'), false)}>{t('task.engine.start')}</Button>}
                {flow.can_pause && <Button loading={busy} onClick={() => void run('pause', t('task.engine.updated'), false)}>{t('task.engine.pause')}</Button>}
                {flow.can_resume && <Button type="primary" loading={busy} onClick={() => void run('resume', t('task.engine.updated'), false)}>{t('task.engine.resume')}</Button>}
                {flow.can_claim && <Button type="primary" loading={busy} onClick={() => void run('claim', t('task.engine.claimed'), true)}>{t('task.engine.claim')}</Button>}
                {flow.can_submit && <Button type="primary" loading={busy} onClick={() => void run('submit', t('task.engine.updated'), true)}>{t('task.engine.submit')}</Button>}
                {flow.can_approve && <Button type="primary" loading={busy} onClick={() => void run('approve', t('task.engine.approved'), true)}>{t('task.engine.approve')}</Button>}
                {flow.can_return && <Button loading={busy} onClick={() => void run('return', t('task.engine.updated'), true)}>{t('task.engine.return')}</Button>}
                {flow.can_reject && <Button danger loading={busy} onClick={() => void run('reject', t('task.engine.rejected'), true)}>{t('task.engine.reject')}</Button>}
              </Space>
            </Form>
          )}
          {!closed && flow.can_delegate && (
            <Form form={delegate} layout="vertical" className={styles.form} onFinish={() => void submitDelegate()}>
              <h4>{t('task.handover.delegateTitle')}</h4>
              <Form.Item name="target" label={t('task.handover.target')} rules={[{ required: true, message: t('task.handover.targetRequired') }]}>
                <UserPicker kind="transfer" disabled={busy} />
              </Form.Item>
              <Form.Item name="reason" label={t('task.handover.reason')} rules={[{ required: true, whitespace: true, message: t('task.handover.reasonRequired') }, { max: 2000 }]}>
                <Input.TextArea rows={3} maxLength={2000} />
              </Form.Item>
              <Button type="primary" htmlType="submit" loading={busy}>{t('task.handover.submit')}</Button>
            </Form>
          )}
          {flow.handover && <HandoverPanel taskId={taskId} value={flow.handover} onChanged={() => { void load(); onChanged(); }} />}
        </>
      )}
    </section>
  );
}
