import { useEffect, useState } from 'react';
import { Form, Input, Modal } from 'antd';
import { assignTask, transferTask } from '@/api/task';
import type { Task } from '@/types/api';
import UserPicker from './UserPicker';
interface Props { task: Task; mode: 'assign' | 'transfer' | null; onClose: () => void; onSaved: () => Promise<void> }
export default function AssignmentModal({ task, mode, onClose, onSaved }: Props) {
  const [form] = Form.useForm<{ assignee: string; reason: string }>();
  const [busy, setBusy] = useState(false);
  useEffect(() => { if (mode) form.resetFields(); }, [mode, form]);
  return <Modal title={mode === 'transfer' ? '转办任务' : '分配负责人'} open={!!mode} onCancel={() => { if (!busy) onClose(); }} onOk={() => form.submit()} confirmLoading={busy} okText="确认交接" cancelText="取消">
    <p>当前负责人：{task.assignee?.name || '尚未分配'}</p>
    <Form form={form} layout="vertical" onFinish={async values => {
      if (!mode) return; setBusy(true);
      try { if (mode === 'assign') await assignTask(task.id, values.assignee); else await transferTask(task.id, values.assignee, values.reason.trim()); onClose(); await onSaved(); } catch { /* Keep selection and reason. */ } finally { setBusy(false); }
    }}>
      <Form.Item name="assignee" label="交给谁处理" rules={[{ required: true, message: '请选择负责人' }, { validator: (_, value) => value && value === task.assignee?.id ? Promise.reject(new Error('请选择另一位负责人')) : Promise.resolve() }]}>{mode && <UserPicker kind={mode} disabled={busy} />}</Form.Item>
      {mode === 'transfer' && <Form.Item name="reason" label="交接说明" rules={[{ required: true, whitespace: true, message: '请说明转办原因和需要接手的内容' }]}><Input.TextArea rows={3} maxLength={500} /></Form.Item>}
    </Form>
  </Modal>;
}
