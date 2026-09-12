import React, { useEffect, useState } from 'react';
import { DatePicker, Form, Input, Modal, Select, Upload, message } from 'antd';
import type { UploadFile } from 'antd/es/upload/interface';
import { UploadOutlined } from '@ant-design/icons';
import type { Dayjs } from 'dayjs';
import { useTranslation } from 'react-i18next';
import { uploadFile } from '@/api/file';
import type { LeaveAttachment, LeaveType, SubmitLeaveParams } from '@/api/leave';
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
  const [loading, setLoading] = useState(false);
  const [fileList, setFileList] = useState<UploadFile[]>([]);

  useEffect(() => {
    if (!open) return;
    form.resetFields();
    setFileList([]);
  }, [open, form]);

  const collectAttachments = async (): Promise<LeaveAttachment[]> => {
    const out: LeaveAttachment[] = [];
    for (const file of fileList) {
      if (file.originFileObj) {
        const fd = new FormData();
        fd.append('file', file.originFileObj);
        const uploaded = await uploadFile(fd);
        out.push({
          file_id: uploaded.id,
          name: uploaded.original_name || uploaded.name || file.name,
          size: uploaded.size || uploaded.file_size || file.size || 0,
        });
        continue;
      }
      if (file.uid) {
        out.push({ file_id: file.uid, name: file.name, size: file.size || 0 });
      }
    }
    return out;
  };

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
          let attachments: LeaveAttachment[] = [];
          try {
            attachments = await collectAttachments();
          } catch {
            message.error(t('leave.uploadFailed'));
            return;
          }
          await onSubmit({
            leave_type_id: values.leave_type_id,
            start_time: values.range[0].format('YYYY-MM-DDTHH:mm:ssZ'),
            end_time: values.range[1].format('YYYY-MM-DDTHH:mm:ssZ'),
            reason: values.reason,
            attachments,
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
            options={types.filter((item) => item.enabled !== false).map((item) => ({
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
        <Form.Item label={t('leave.attachments')}>
          <Upload
            multiple
            maxCount={5}
            fileList={fileList}
            beforeUpload={(file) => {
              if (file.size / 1024 / 1024 > 20) {
                message.error(t('leave.fileTooLarge'));
                return Upload.LIST_IGNORE;
              }
              return false;
            }}
            onChange={({ fileList: next }) => setFileList(next)}
          >
            <button type="button" className="leave-upload-btn">
              <UploadOutlined /> {t('leave.upload')}
            </button>
          </Upload>
        </Form.Item>
      </Form>
    </Modal>
  );
};

export default FormModal;
