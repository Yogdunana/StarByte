import React, { useEffect } from 'react';
import { DatePicker, Form, Input, Modal, Select } from 'antd';
import type { Dayjs } from 'dayjs';
import { useTranslation } from 'react-i18next';
import type { LeaveType, SubmitLeaveParams } from '@/api/leave';
import { leaveTypeLabel } from './meta';

interface FormValues {
  leave_type_id: string;
  range: [Dayjs, Dayjs];
  reason: string;
}

interface Props {
  open: boolean;
  types: LeaveType[];
  onClose: () => void;
  onSubmit: (values: SubmitLeaveParams) => Promise<void>;
}

const FormModal: React.FC<Props> = ({ open, types, onClose, onSubmit }) => {
  const { t } = useTranslation();
  const [form] = Form.useForm<FormValues>();
  const [loading, setLoading] = React.useState(false);

  useEffect(() => {
    if (!open) return;
    form.resetFields();
  }, [open, form]);

  return (
    <Modal
      title={t('leave.apply')}
      open={open}
      onCancel={onClose}
      confirmLoading={loading}
      onOk={async () => {
        if (loading) return;
        setLoading(true);
        try {
          const values = await form.validateFields();
          await onSubmit({
            leave_type_id: values.leave_type_id,
            start_time: values.range[0].format('YYYY-MM-DDTHH:mm:ssZ'),
            end_time: values.range[1].format('YYYY-MM-DDTHH:mm:ssZ'),
            reason: values.reason,
          });
          onClose();
        } finally {
          setLoading(false);
        }
      }}
      destroyOnClose
    >
      <Form form={form} layout="vertical">
        <Form.Item name="leave_type_id" label={t('leave.typeLabel')} rules={[{ required: true }]}>
          <Select
            options={types.map((item) => ({
              value: item.id,
              label: `${leaveTypeLabel(t, item.code, item.name)}${item.deductible ? ` (${t('leave.deductible')})` : ''}`,
            }))}
          />
        </Form.Item>
        <Form.Item name="range" label={t('leave.range')} rules={[{ required: true }]}>
          <DatePicker.RangePicker showTime style={{ width: '100%' }} />
        </Form.Item>
        <Form.Item name="reason" label={t('leave.reason')} rules={[{ required: true, max: 500 }]}>
          <Input.TextArea rows={4} maxLength={500} showCount />
        </Form.Item>
      </Form>
    </Modal>
  );
};

export default FormModal;
