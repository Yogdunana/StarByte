import { useEffect, useState } from 'react';
import { DatePicker, Form, Input, Modal, Select, Switch } from 'antd';
import dayjs from 'dayjs';
import type { CreateTaskParams, Task } from '@/types/api';
import UserPicker from './UserPicker';
import AssignmentRuleFields from './AssignmentRuleFields';
interface Props { open: boolean; editing: Task | null; onCancel: () => void; onSubmit: (values: CreateTaskParams & { clear_due_date?: boolean }) => Promise<void> }
interface Fields extends Omit<CreateTaskParams, 'due_date'> { due_date?: dayjs.Dayjs; use_workflow?: boolean }
export default function FormModal({ open, editing, onCancel, onSubmit }: Props) {
  const [form] = Form.useForm<Fields>();
  const [busy, setBusy] = useState(false);
 const useWorkflow = Form.useWatch("use_workflow", form);
 const assignmentMode = Form.useWatch(["workflow", "assignment", "mode"], form);
  useEffect(() => {
    if (!open) return;
    form.resetFields();
    form.setFieldsValue(editing ? { title: editing.title, description: editing.description, priority: editing.priority, tags: editing.tags, due_date: editing.due_date ? dayjs(editing.due_date) : undefined } : { priority: 1 });
  }, [open, editing, form]);
  return <Modal title={editing ? '编辑任务' : '安排一项任务'} open={open} onCancel={() => { if (!busy) onCancel(); }} onOk={() => form.submit()} confirmLoading={busy} okText={editing ? '保存修改' : '创建任务'} cancelText="暂不保存">
    <Form form={form} layout="vertical" onFinish={async values => {
      setBusy(true);
      try { const { use_workflow, ...payload } = values;
        const automatic = use_workflow && values.workflow?.assignment?.mode && values.workflow.assignment.mode !== 'manual';
        const workflow = use_workflow ? values.workflow : undefined;
        if (workflow?.assignment && !['role', 'round_robin'].includes(workflow.assignment.mode)) workflow.assignment.role_id = undefined;
        await onSubmit({ ...payload, assignee_id: automatic ? undefined : values.assignee_id, workflow, title: values.title.trim(), due_date: values.due_date?.toISOString(), clear_due_date: !!editing && !values.due_date }); } catch { /* Keep the draft after API failure. */ } finally { setBusy(false); }
    }}>
      <Form.Item name="title" label="任务名称" rules={[{ required: true, whitespace: true, message: '请写明要完成的事情' }]}><Input maxLength={200} placeholder="例如：整理新生见面会活动方案" /></Form.Item>
      <Form.Item name="description" label="任务说明"><Input.TextArea rows={4} placeholder="写清交付内容、协作方式和验收要求" /></Form.Item>
      {!editing && (!useWorkflow || !assignmentMode || assignmentMode === "manual") && <Form.Item name="assignee_id" label="负责人" extra="可先创建，再分配负责人。任务归属默认为你的部门。"><UserPicker kind="create" disabled={busy} /></Form.Item>}
      {!editing && <><Form.Item name="use_workflow" label="审核与正式验收" valuePropName="checked" extra="开启后，执行人提交成果，由指定人员分别审核和验收。"><Switch /></Form.Item>{useWorkflow && <><AssignmentRuleFields mode={assignmentMode} busy={busy} /><Form.Item name={['workflow', 'reviewer_id']} label="审核人" rules={[{ required: true, message: '请选择审核人' }]}><UserPicker kind="create" disabled={busy} /></Form.Item><Form.Item name={['workflow', 'acceptor_id']} label="验收人" rules={[{ required: true, message: '请选择验收人' }]} extra="执行人不能审核或验收自己的交付。"><UserPicker kind="create" disabled={busy} /></Form.Item></>}</>}
      <Form.Item name="priority" label="优先级" rules={[{ required: true }]}><Select options={[{ value: 0, label: '低 · 可以稍后处理' }, { value: 1, label: '中 · 正常推进' }, { value: 2, label: '高 · 优先处理' }, { value: 3, label: '紧急 · 尽快响应' }]} /></Form.Item>
      <Form.Item name="due_date" label="截止时间"><DatePicker showTime style={{ width: '100%' }} /></Form.Item>
      <Form.Item name="tags" label="标签"><Select mode="tags" tokenSeparators={[',', '，']} placeholder="输入标签后按回车" /></Form.Item>
    </Form>
  </Modal>;
}
