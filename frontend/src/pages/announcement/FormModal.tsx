import React, { useEffect } from 'react';
import { DatePicker, Form, Input, Modal, Select, Switch } from 'antd';
import dayjs from 'dayjs';
import { useTranslation } from 'react-i18next';
import type { Announcement } from '@/api/announcement';
import { AnnouncementCategories } from './meta';

interface Props {
  open: boolean;
  editing: Announcement | null;
  canSchedule?: boolean;
  onCancel: () => void;
  onSubmit: (values: Record<string, unknown>) => Promise<void>;
}

const FormModal: React.FC<Props> = ({ open, editing, canSchedule, onCancel, onSubmit }) => {
  const { t } = useTranslation();
  const [form] = Form.useForm();

  useEffect(() => {
    if (!open) return;
    if (editing) {
      form.setFieldsValue({
        title: editing.title,
        category: editing.category,
        content: editing.content,
        required: editing.required,
        scheduled_at: editing.scheduled_at ? dayjs(editing.scheduled_at) : undefined,
      });
    } else {
      form.resetFields();
      form.setFieldsValue({ required: false });
    }
  }, [open, editing, form]);

  return (
    <Modal
      title={editing ? t('announcement.edit') : t('announcement.create')}
      open={open}
      onCancel={onCancel}
      onOk={() => form.submit()}
      width={720}
      destroyOnClose
      okText={t('announcement.saveDraft')}
    >
      <Form
        form={form}
        layout="vertical"
        onFinish={async (values) => {
          const isDraft = !editing || editing.status === 0;
          const scheduled = canSchedule && values.scheduled_at
            ? (values.scheduled_at as dayjs.Dayjs).toISOString()
            : undefined;
          await onSubmit({
            title: values.title,
            category: values.category,
            content: values.content,
            content_type: 'markdown',
            required: values.required,
            scheduled_at: isDraft && canSchedule ? scheduled : undefined,
            clear_scheduled_at: Boolean(
              editing && (editing.status !== 0 || (!scheduled && editing.scheduled_at)),
            ),
          });
        }}
      >
        <Form.Item name="title" label={t('announcement.title')} rules={[{ required: true }]}>
          <Input maxLength={200} showCount />
        </Form.Item>
        <Form.Item name="category" label={t('announcement.categoryLabel')} rules={[{ required: true }]}>
          <Select
            options={AnnouncementCategories.map((item) => ({
              value: item.value,
              label: t(item.labelKey),
            }))}
          />
        </Form.Item>
        <Form.Item name="content" label={t('announcement.content')}>
          <Input.TextArea rows={10} maxLength={50000} showCount placeholder={t('announcement.contentHint')} />
        </Form.Item>
        {canSchedule && (!editing || editing.status === 0) && (
          <Form.Item name="scheduled_at" label={t('announcement.scheduledAt')}>
            <DatePicker showTime style={{ width: '100%' }} />
          </Form.Item>
        )}
        <Form.Item name="required" label={t('announcement.required')} valuePropName="checked">
          <Switch />
        </Form.Item>
      </Form>
    </Modal>
  );
};

export default FormModal;
