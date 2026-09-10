import React, { useEffect, useState } from 'react';
import { DatePicker, Form, Input, InputNumber, Modal, Select, message } from 'antd';
import dayjs from 'dayjs';
import type { Activity, CreateActivityParams, UpdateActivityParams } from '@/api/activity';
import { createActivity, updateActivity } from '@/api/activity';
import { ActivityCategoryOptions } from './meta';

interface Props {
  open: boolean;
  editing: Activity | null;
  onClose: () => void;
  onSaved: () => void;
}

const { RangePicker } = DatePicker;

const FormModal: React.FC<Props> = ({ open, editing, onClose, onSaved }) => {
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (open) {
      if (editing) {
        form.setFieldsValue({
          title: editing.title,
          description: editing.description,
          category: editing.category,
          tags: editing.tags,
          location: editing.location,
          max_participants: editing.max_participants,
          time: editing.start_time && editing.end_time
            ? [dayjs(editing.start_time), dayjs(editing.end_time)] : undefined,
        });
      } else {
        form.resetFields();
      }
    }
  }, [open, editing, form]);

  const onOk = async () => {
    const values = await form.validateFields();
    setLoading(true);
    try {
      const [start, end] = values.time;
      const payload: CreateActivityParams | UpdateActivityParams = {
        title: values.title,
        description: values.description,
        category: values.category,
        tags: values.tags,
        location: values.location,
        max_participants: values.max_participants || 0,
        start_time: start.toISOString(),
        end_time: end.toISOString(),
      };
      if (editing) {
        await updateActivity(editing.id, payload);
        message.success('更新成功');
      } else {
        await createActivity(payload);
        message.success('创建成功');
      }
      onSaved();
    } finally {
      setLoading(false);
    }
  };

  return (
    <Modal
      title={editing ? '编辑活动' : '新建活动'}
      open={open} onCancel={onClose} onOk={onOk} confirmLoading={loading}
      width={600} destroyOnClose
    >
      <Form form={form} layout="vertical" preserve={false}>
        <Form.Item name="title" label="活动标题" rules={[{ required: true, message: '请输入标题' }]}>
          <Input maxLength={200} showCount />
        </Form.Item>
        <Form.Item name="time" label="活动时间" rules={[{ required: true, message: '请选择时间' }]}>
          <RangePicker showTime style={{ width: '100%' }} />
        </Form.Item>
        <Form.Item name="category" label="分类">
          <Select options={ActivityCategoryOptions} allowClear />
        </Form.Item>
        <Form.Item name="location" label="地点">
          <Input maxLength={200} />
        </Form.Item>
        <Form.Item name="max_participants" label="人数上限">
          <InputNumber min={0} style={{ width: '100%' }} placeholder="0 表示不限" />
        </Form.Item>
        <Form.Item name="tags" label="标签">
          <Select mode="tags" placeholder="输入标签后回车" />
        </Form.Item>
        <Form.Item name="description" label="活动描述">
          <Input.TextArea rows={4} maxLength={2000} showCount />
        </Form.Item>
      </Form>
    </Modal>
  );
};

export default FormModal;
