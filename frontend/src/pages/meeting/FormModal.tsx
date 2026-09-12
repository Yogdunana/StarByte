import { tx, useLocale } from '@/i18n/text';
import { useEffect, useState } from 'react';
import { DatePicker, Form, Input, Modal, Select } from 'antd';
import dayjs from 'dayjs';
import type { CreateMeetingParams, Meeting } from '@/types/api';
interface Props {
  open: boolean;
  editing: Meeting | null;
  onCancel: () => void;
  onSubmit: (values: CreateMeetingParams) => Promise<void>;
}
interface Fields {
  title: string;
  description?: string;
  meeting_type: number;
  location?: string;
  online_link?: string;
  time_range: [dayjs.Dayjs, dayjs.Dayjs];
}
export default function FormModal({ open, editing, onCancel, onSubmit }: Props) {
  useLocale();
  const [form] = Form.useForm<Fields>();
  const [busy, setBusy] = useState(false);
  useEffect(() => {
    if (!open) return;
    form.resetFields();
    if (editing)
      form.setFieldsValue({
        ...editing,
        time_range: [dayjs(editing.start_time), dayjs(editing.end_time)],
      });
    else form.setFieldsValue({ meeting_type: 1 });
  }, [open, editing, form]);
  const submit = async (values: Fields) => {
    setBusy(true);
    try {
      const { time_range: range, ...data } = values;
      await onSubmit({
        ...data,
        title: data.title.trim(),
        start_time: range[0].toISOString(),
        end_time: range[1].toISOString(),
      });
    } catch {
      /* Preserve input and show the API error. */
    } finally {
      setBusy(false);
    }
  };
  return (
    <Modal
      title={editing ? tx('编辑会议') : tx('新建会议')}
      open={open}
      onCancel={() => {
        if (!busy) onCancel();
      }}
      onOk={() => form.submit()}
      confirmLoading={busy}
      okText={tx('保存会议')}
      cancelText={tx('暂不保存')}
    >
      <Form form={form} layout="vertical" onFinish={(values) => void submit(values)}>
        <Form.Item
          name="title"
          label={tx('会议标题')}
          rules={[{ required: true, whitespace: true, message: tx('请填写会议标题') }]}
        >
          <Input maxLength={200} placeholder={tx('例如：新学期工作安排会')} />
        </Form.Item>
        <Form.Item name="meeting_type" label={tx('会议类型')} rules={[{ required: true }]}>
          <Select
            options={[
              { value: 1, label: tx('例会') },
              { value: 2, label: tx('临时会议') },
              { value: 3, label: tx('线上会议') },
            ]}
          />
        </Form.Item>
        <Form.Item
          name="time_range"
          label={tx('开始与结束时间')}
          rules={[
            { required: true, message: tx('请选择会议时间') },
            {
              validator: (_, value: Fields['time_range'] | undefined) =>
                !value || value[1]?.isAfter(value[0])
                  ? Promise.resolve()
                  : Promise.reject(new Error(tx('结束时间必须晚于开始时间'))),
            },
          ]}
        >
          <DatePicker.RangePicker showTime style={{ width: '100%' }} />
        </Form.Item>
        <Form.Item name="location" label={tx('地点')}>
          <Input maxLength={200} placeholder={tx('教学楼、教室或线上')} />
        </Form.Item>
        <Form.Item
          name="online_link"
          label={tx('线上会议链接')}
          rules={[{ type: 'url', message: tx('请填写完整的 http 或 https 链接') }]}
        >
          <Input maxLength={500} placeholder="https://" />
        </Form.Item>
        <Form.Item name="description" label={tx('会议说明')}>
          <Input.TextArea rows={3} placeholder={tx('写明会议目的和需要提前准备的内容')} />
        </Form.Item>
      </Form>
    </Modal>
  );
}
