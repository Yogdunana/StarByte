import React, { useEffect, useState } from 'react';
import { DatePicker, Form, Input, InputNumber, Modal, Select, Switch, Upload, message } from 'antd';
import type { UploadFile } from 'antd/es/upload/interface';
import { UploadOutlined } from '@ant-design/icons';
import dayjs from 'dayjs';
import { useTranslation } from 'react-i18next';
import { uploadFile } from '@/api/file';
import { getRoleList } from '@/api/role';
import { getDepartmentTree } from '@/api/department';
import { getUserList } from '@/api/user';
import type { Announcement, AnnouncementAttachment, AnnouncementAudience } from '@/api/announcement';
import type { Department } from '@/types/api';
import { AnnouncementCategories } from './meta';

interface Props {
  open: boolean;
  editing: Announcement | null;
  canSchedule?: boolean;
  canManage?: boolean;
  onCancel: () => void;
  onSubmit: (values: Record<string, unknown>) => Promise<void>;
}

function flattenDepartments(nodes: Department[], prefix = ''): { value: string; label: string }[] {
  const out: { value: string; label: string }[] = [];
  nodes.forEach((node) => {
    const label = prefix ? `${prefix} / ${node.name}` : node.name;
    out.push({ value: node.id, label });
    if (node.children?.length) out.push(...flattenDepartments(node.children, label));
  });
  return out;
}

const FormModal: React.FC<Props> = ({ open, editing, canSchedule, canManage, onCancel, onSubmit }) => {
  const { t } = useTranslation();
  const [form] = Form.useForm();
  const audienceType = Form.useWatch('audience_type', form) as AnnouncementAudience | undefined;
  const [roleOpts, setRoleOpts] = useState<{ value: string; label: string }[]>([]);
  const [deptOpts, setDeptOpts] = useState<{ value: string; label: string }[]>([]);
  const [userOpts, setUserOpts] = useState<{ value: string; label: string }[]>([]);
  const [fileList, setFileList] = useState<UploadFile[]>([]);

  useEffect(() => {
    if (!open) return;
    void getRoleList({ page: 1, page_size: 100 }).then((res) => {
      setRoleOpts(res.list.map((item) => ({ value: item.id, label: item.name || item.code })));
    }).catch(() => setRoleOpts([]));
    void getDepartmentTree().then((tree) => setDeptOpts(flattenDepartments(tree))).catch(() => setDeptOpts([]));
    void getUserList({ page: 1, page_size: 100, status: 0 }).then((res) => {
      setUserOpts(res.list.map((item) => ({
        value: item.id,
        label: item.real_name || item.username,
      })));
    }).catch(() => setUserOpts([]));
  }, [open]);

  useEffect(() => {
    if (!open) return;
    if (editing) {
      form.setFieldsValue({
        title: editing.title,
        category: editing.category,
        content: editing.content,
        content_type: editing.content_type || 'markdown',
        required: editing.required,
        sort_order: editing.sort_order,
        scheduled_at: editing.scheduled_at ? dayjs(editing.scheduled_at) : undefined,
        expires_at: editing.expires_at ? dayjs(editing.expires_at) : undefined,
        audience_type: editing.audience_type || 'all',
        audience_ids: editing.audience_ids || [],
      });
      setFileList((editing.attachments || []).map((item) => ({
        uid: item.file_id,
        name: item.name,
        status: 'done',
        size: item.size,
      })));
    } else {
      form.resetFields();
      form.setFieldsValue({ required: false, audience_type: 'all', content_type: 'markdown', sort_order: 0 });
      setFileList([]);
    }
  }, [open, editing, form]);

  const collectAttachments = async (): Promise<AnnouncementAttachment[]> => {
    const out: AnnouncementAttachment[] = [];
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
          const expires = values.expires_at
            ? (values.expires_at as dayjs.Dayjs).toISOString()
            : undefined;
          let attachments: AnnouncementAttachment[] = [];
          try {
            attachments = await collectAttachments();
          } catch {
            message.error(t('announcement.uploadFailed'));
            return;
          }
          const audience = (values.audience_type as AnnouncementAudience) || 'all';
          await onSubmit({
            title: values.title,
            category: values.category,
            content: values.content,
            content_type: values.content_type || 'markdown',
            required: values.required,
            sort_order: canManage ? values.sort_order : undefined,
            scheduled_at: isDraft && canSchedule ? scheduled : undefined,
            clear_scheduled_at: Boolean(
              canSchedule &&
                editing &&
                (editing.status !== 0 || (!scheduled && editing.scheduled_at)),
            ),
            expires_at: expires,
            clear_expires_at: Boolean(editing && !expires && editing.expires_at),
            audience_type: audience,
            audience_ids: audience === 'all' ? [] : (values.audience_ids as string[]) || [],
            attachments,
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
        <Form.Item name="content_type" label={t('announcement.contentType')}>
          <Select
            options={[
              { value: 'markdown', label: t('announcement.markdown') },
              { value: 'html', label: t('announcement.html') },
            ]}
          />
        </Form.Item>
        <Form.Item name="content" label={t('announcement.content')}>
          <Input.TextArea rows={10} maxLength={50000} showCount placeholder={t('announcement.contentHint')} />
        </Form.Item>
        <Form.Item label={t('announcement.attachments')}>
          <Upload
            multiple
            maxCount={10}
            fileList={fileList}
            beforeUpload={(file) => {
              if (file.size / 1024 / 1024 > 20) {
                message.error(t('announcement.fileTooLarge'));
                return Upload.LIST_IGNORE;
              }
              return false;
            }}
            onChange={({ fileList: next }) => setFileList(next)}
          >
            <button type="button" style={{ border: '1px dashed #d9d9d9', padding: '4px 15px', borderRadius: 6, background: '#fff', cursor: 'pointer' }}>
              <UploadOutlined /> {t('announcement.upload')}
            </button>
          </Upload>
        </Form.Item>
        <Form.Item name="audience_type" label={t('announcement.audience')}>
          <Select
            options={[
              { value: 'all', label: t('announcement.audienceAll') },
              { value: 'role', label: t('announcement.audienceRole') },
              { value: 'department', label: t('announcement.audienceDept') },
              { value: 'users', label: t('announcement.audienceUsers') },
            ]}
          />
        </Form.Item>
        {audienceType && audienceType !== 'all' && (
          <Form.Item
            name="audience_ids"
            label={t('announcement.audienceIds')}
            rules={[{ required: true, type: 'array', min: 1 }]}
          >
            <Select
              mode="multiple"
              options={audienceType === 'role' ? roleOpts : audienceType === 'department' ? deptOpts : userOpts}
              optionFilterProp="label"
              showSearch
            />
          </Form.Item>
        )}
        {canSchedule && (!editing || editing.status === 0) && (
          <Form.Item name="scheduled_at" label={t('announcement.scheduledAt')}>
            <DatePicker showTime style={{ width: '100%' }} />
          </Form.Item>
        )}
        <Form.Item name="expires_at" label={t('announcement.expiresAt')}>
          <DatePicker showTime style={{ width: '100%' }} />
        </Form.Item>
        <Form.Item name="required" label={t('announcement.required')} valuePropName="checked">
          <Switch />
        </Form.Item>
        {canManage && (
          <Form.Item name="sort_order" label={t('announcement.sortOrder')}>
            <InputNumber style={{ width: '100%' }} />
          </Form.Item>
        )}
      </Form>
    </Modal>
  );
};

export default FormModal;
