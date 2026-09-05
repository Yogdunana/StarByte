import React, { useEffect } from 'react';
import { DatePicker, Form, Input, Modal, Select } from 'antd';
import dayjs from 'dayjs';
import type { Meeting } from '@/types/api';

interface Props {
  open: boolean;
  editing: Meeting | null;
  onCancel: () => void;
  onSubmit: (values: Record<string, unknown>) => Promise<void>;
}

const FormModal: React.FC<Props> = ({ open, editing, onCancel, onSubmit }) => {
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
      form.setFieldsValue({ meeting_type: 1 });
    }
  }, [open, editing, form]);

  return (
    <Modal
      title={editing ? '编辑会议' : '新建会议'}
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
            description: values.description,
            meeting_type: values.meeting_type,
            location: values.location,
            online_link: values.online_link,
            start_time: range[0].toISOString(),
            end_time: range[1].toISOString(),
          });
        }}
      >
        <Form.Item name="title" label="标题" rules={[{ required: true }]}>
          <Input maxLength={200} />
        </Form.Item>
        <Form.Item name="meeting_type" label="类型" rules={[{ required: true }]}>
          <Select
            options={[
              { value: 1, label: '例会' },
              { value: 2, label: '临时会议' },
              { value: 3, label: '线上会议' },
            ]}
          />
        </Form.Item>
        <Form.Item name="time_range" label="时间" rules={[{ required: true }]}>
          <DatePicker.RangePicker showTime style={{ width: '100%' }} />
        </Form.Item>
        <Form.Item name="location" label="地点">
          <Input />
        </Form.Item>
        <Form.Item name="online_link" label="线上链接">
          <Input />
        </Form.Item>
        <Form.Item name="description" label="说明">
          <Input.TextArea rows={3} />
        </Form.Item>
      </Form>
    </Modal>
  );
};

export default FormModal;
