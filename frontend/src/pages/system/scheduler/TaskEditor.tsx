import React from 'react';
import { DatePicker, Form, Input, InputNumber, Modal, Select } from 'antd';
import type { FormInstance } from 'antd/es/form';
import type { SchedulerHandlerInfo, SchedulerTask } from '@/types/api';
import dayjs from 'dayjs';

export interface TaskFormValues {
  name: string;
  code: string;
  schedule_kind: 'cron' | 'once';
  cron_expr?: string;
  run_at?: dayjs.Dayjs;
  timezone?: string;
  handler_key: string;
  payload?: string;
  max_retries?: number;
  timeout_sec?: number;
  shard_key?: string;
}

interface Props {
  open: boolean;
  editing: SchedulerTask | null;
  handlers: SchedulerHandlerInfo[];
  form: FormInstance<TaskFormValues>;
  onCancel: () => void;
  onOk: () => void;
}

export function fillTaskForm(row: SchedulerTask): TaskFormValues {
  return {
    name: row.name,
    code: row.code,
    schedule_kind: row.cron_expr ? 'cron' : 'once',
    cron_expr: row.cron_expr || undefined,
    run_at: row.run_at ? dayjs(row.run_at) : undefined,
    timezone: row.timezone,
    handler_key: row.handler_key,
    payload: row.payload,
    max_retries: row.max_retries,
    timeout_sec: row.timeout_sec,
    shard_key: row.shard_key,
  };
}

const TaskEditor: React.FC<Props> = ({ open, editing, handlers, form, onCancel, onOk }) => {
  const kind = Form.useWatch('schedule_kind', form);
  return (
    <Modal
      title={editing ? '编辑定时任务' : '新建定时任务'}
      open={open}
      onCancel={onCancel}
      onOk={onOk}
      destroyOnClose
      width={560}
    >
      <Form form={form} layout="vertical" initialValues={{
        schedule_kind: 'cron', timezone: 'Asia/Shanghai', max_retries: 3, timeout_sec: 60,
        cron_expr: '0 */5 * * * *',
      }}>
        <Form.Item name="name" label="名称" rules={[{ required: true, message: '请输入名称' }]}>
          <Input />
        </Form.Item>
        <Form.Item name="code" label="编码" rules={[{ required: true, message: '请输入编码' }]}>
          <Input disabled={Boolean(editing)} placeholder="warmup-dict" />
        </Form.Item>
        <Form.Item name="schedule_kind" label="调度方式">
          <Select options={[
            { value: 'cron', label: 'Cron 周期' },
            { value: 'once', label: '一次性' },
          ]} />
        </Form.Item>
        {kind !== 'once' && (
          <Form.Item name="cron_expr" label="Cron（秒 分 时 日 月 周）" rules={[{ required: true }]}>
            <Input placeholder="0 */5 * * * *" />
          </Form.Item>
        )}
        {kind === 'once' && (
          <Form.Item name="run_at" label="执行时间" rules={[{ required: true }]}>
            <DatePicker showTime style={{ width: '100%' }} />
          </Form.Item>
        )}
        <Form.Item name="timezone" label="时区">
          <Input />
        </Form.Item>
        <Form.Item name="handler_key" label="处理器" rules={[{ required: true }]}>
          <Select options={handlers.map((h) => ({ value: h.key, label: `${h.key} · ${h.description}` }))} />
        </Form.Item>
        <Form.Item name="payload" label="Payload">
          <Input.TextArea rows={3} />
        </Form.Item>
        <Form.Item name="max_retries" label="最大重试">
          <InputNumber min={0} max={20} style={{ width: '100%' }} />
        </Form.Item>
        <Form.Item name="timeout_sec" label="超时（秒）">
          <InputNumber min={1} max={3600} style={{ width: '100%' }} />
        </Form.Item>
        <Form.Item name="shard_key" label="分片键（可选）">
          <Input placeholder="留空则所有节点竞锁执行" />
        </Form.Item>
      </Form>
    </Modal>
  );
};

export default TaskEditor;
