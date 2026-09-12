import { useEffect, useState } from 'react';
import { Form, Input, Modal } from 'antd';
import { useTranslation } from 'react-i18next';
import { assignTask, transferTask } from '@/api/task';
import type { Task } from '@/types/api';
import UserPicker from './UserPicker';
interface Props { task: Task; mode: 'assign' | 'transfer' | null; onClose: () => void; onSaved: () => Promise<void> }
export default function AssignmentModal({ task, mode, onClose, onSaved }: Props) {
  const { t } = useTranslation();
  const [form] = Form.useForm<{ assignee: string; reason: string }>();
  const [busy, setBusy] = useState(false);
  useEffect(() => { if (mode) form.resetFields(); }, [mode, form]);
  return <Modal
    title={mode === 'transfer' ? t('task.handover.delegateTitle') : t('task.form.manual')}
    open={!!mode}
    onCancel={() => { if (!busy) onClose(); }}
    onOk={() => form.submit()}
    confirmLoading={busy}
    okText={t('task.handover.submit')}
    cancelText={t('common.cancel', { defaultValue: '取消' })}
  >
    <p>{t('task.engine.waitingAssignee')}：{task.assignee?.name || t('task.engine.waitingAssignee')}</p>
    <Form form={form} layout="vertical" onFinish={async values => {
      if (!mode) return; setBusy(true);
      try { if (mode === 'assign') await assignTask(task.id, values.assignee); else await transferTask(task.id, values.assignee, values.reason.trim()); onClose(); await onSaved(); } catch { /* Keep selection and reason. */ } finally { setBusy(false); }
    }}>
      <Form.Item name="assignee" label={t('task.handover.target')} rules={[{ required: true, message: t('task.handover.targetRequired') }, { validator: (_, value) => value && value === task.assignee?.id ? Promise.reject(new Error(t('task.handover.targetRequired'))) : Promise.resolve() }]}>{mode && <UserPicker kind={mode} disabled={busy} />}</Form.Item>
      {mode === 'transfer' && <Form.Item name="reason" label={t('task.handover.reason')} rules={[{ required: true, whitespace: true, message: t('task.handover.reasonRequired') }]}><Input.TextArea rows={3} maxLength={500} /></Form.Item>}
    </Form>
  </Modal>;
}
