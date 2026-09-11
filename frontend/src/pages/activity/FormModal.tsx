import React, { useEffect } from 'react';
import { DatePicker, Form, Input, InputNumber, Modal, Select } from 'antd';
import dayjs from 'dayjs';
import type { Activity } from '@/api/activity';
import { ActivityCategoryOptions } from './meta';

interface Props {
  open: boolean;
  editing: Activity | null;
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
      form.setFieldsValue({ max_participants: 0 });
    }
  }, [open, editing, form]);

  return (
    <Modal
      title={editing ? '编辑活动' : '新建活动'}
      open={open}
      onCancel={onCancel}
      onOk={() => form.submit()}
      width={600}
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
            category: values.category,
            tags: values.tags,
            location: values.location,
            max_participants: values.max_participants,
            latitude: values.latitude,
            longitude: values.longitude,
            checkin_radius_m: values.checkin_radius_m,
            start_time: range[0].toISOString(),
            end_time: range[1].toISOString(),
          });
        }}
      >
        <Form.Item name="title" label="标题" rules={[{ required: true }]}>
          <Input maxLength={200} showCount />
        </Form.Item>
        <Form.Item name="time_range" label="时间" rules={[{ required: true }]}>
          <DatePicker.RangePicker showTime style={{ width: '100%' }} />
        </Form.Item>
        <Form.Item name="category" label="分类">
          <Select options={ActivityCategoryOptions} allowClear />
        </Form.Item>
        <Form.Item name="location" label="地点">
          <Input maxLength={200} />
        </Form.Item>
        <Form.Item name="latitude" label="纬度（GPS 围栏，可选）">
          <InputNumber min={-90} max={90} step={0.000001} style={{ width: '100%' }} />
        </Form.Item>
        <Form.Item name="longitude" label="经度（GPS 围栏，可选）">
          <InputNumber min={-180} max={180} step={0.000001} style={{ width: '100%' }} />
        </Form.Item>
        <Form.Item name="checkin_radius_m" label="签到半径（米，未配置则拒绝 GPS 签到）">
          <InputNumber min={0} style={{ width: '100%' }} />
        </Form.Item>
        <Form.Item name="max_participants" label="人数上限（0 表示不限）">
          <InputNumber min={0} style={{ width: '100%' }} />
        </Form.Item>
        <Form.Item name="tags" label="标签">
          <Select mode="tags" placeholder="输入标签后回车" />
        </Form.Item>
        <Form.Item name="description" label="说明">
          <Input.TextArea rows={3} maxLength={2000} showCount />
        </Form.Item>
      </Form>
    </Modal>
  );
};

export default FormModal;
