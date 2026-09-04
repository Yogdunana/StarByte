import React, { useEffect } from 'react';
import { DatePicker, Form, Input, InputNumber, Modal, Select } from 'antd';
import dayjs from 'dayjs';
import type { InterviewSession, MemberDepartmentOption } from '@/types/api';

interface Props {
  open: boolean;
  editing: InterviewSession | null;
  departments: MemberDepartmentOption[];
  onCancel: () => void;
  onSubmit: (values: Record<string, unknown>) => Promise<void>;
}

const SessionFormModal: React.FC<Props> = ({ open, editing, departments, onCancel, onSubmit }) => {
  const [form] = Form.useForm();

  useEffect(() => {
    if (!open) return;
    if (editing) {
      form.setFieldsValue({
        ...editing,
        time_range: [dayjs(editing.start_time), dayjs(editing.end_time)],
      });
    } else {
      form.resetFields();
      form.setFieldsValue({ round: 1, max_candidates: 20 });
    }
  }, [open, editing, form]);

  return (
    <Modal
      title={editing ? '编辑场次' : '新建场次'}
      open={open}
      onCancel={onCancel}
      onOk={() => form.submit()}
      destroyOnClose
    >
      <Form
        form={form}
        layout="vertical"
        onFinish={async (values) => {
          const range = values.time_range as [dayjs.Dayjs, dayjs.Dayjs];
          await onSubmit({
            title: values.title,
            round: values.round,
            department_id: values.department_id,
            start_time: range[0].toISOString(),
            end_time: range[1].toISOString(),
            location: values.location,
            online_link: values.online_link,
            max_candidates: values.max_candidates,
            description: values.description,
          });
        }}
      >
        <Form.Item name="title" label="场次名称" rules={[{ required: true, message: '请输入名称' }]}>
          <Input maxLength={200} />
        </Form.Item>
        <Form.Item name="round" label="轮次" rules={[{ required: true }]}>
          <InputNumber min={1} max={20} style={{ width: '100%' }} />
        </Form.Item>
        <Form.Item name="department_id" label="部门">
          <Select
            allowClear
            options={departments.map((d) => ({ label: d.name, value: d.id }))}
            placeholder="可选"
          />
        </Form.Item>
        <Form.Item name="time_range" label="时间" rules={[{ required: true, message: '请选择时间' }]}>
          <DatePicker.RangePicker showTime style={{ width: '100%' }} />
        </Form.Item>
        <Form.Item name="location" label="地点">
          <Input />
        </Form.Item>
        <Form.Item name="online_link" label="线上链接">
          <Input />
        </Form.Item>
        <Form.Item name="max_candidates" label="最大人数" rules={[{ required: true }]}>
          <InputNumber min={1} max={500} style={{ width: '100%' }} />
        </Form.Item>
        <Form.Item name="description" label="说明">
          <Input.TextArea rows={3} />
        </Form.Item>
      </Form>
    </Modal>
  );
};

export default SessionFormModal;
