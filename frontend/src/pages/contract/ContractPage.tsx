import React, { useCallback, useEffect, useState } from 'react';
import { Button, Card, DatePicker, Form, Input, InputNumber, Modal, Select, Space, Table, Upload, message } from 'antd';
import { PlusOutlined, UploadOutlined } from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import type { UploadFile } from 'antd/es/upload/interface';
import dayjs from 'dayjs';
import { useTranslation } from 'react-i18next';
import { usePermission } from '@/hooks/usePermission';
import { uploadFile } from '@/api/file';
import {
  createContract,
  deleteContract,
  getContractTemplates,
  getContracts,
  getExpiringContracts,
  updateContract,
  type ContractItem,
  type ContractTemplate,
} from '@/api/contract';

const ContractPage: React.FC = () => {
  const { t } = useTranslation();
  const canCreate = usePermission('contract:create');
  const canManage = usePermission('contract:manage');
  const [list, setList] = useState<ContractItem[]>([]);
  const [expiring, setExpiring] = useState<ContractItem[]>([]);
  const [tpls, setTpls] = useState<ContractTemplate[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(false);
  const [open, setOpen] = useState(false);
  const [editing, setEditing] = useState<ContractItem | null>(null);
  const [fileList, setFileList] = useState<UploadFile[]>([]);
  const [form] = Form.useForm();

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const [res, soon] = await Promise.all([getContracts({ page, page_size: 10 }), getExpiringContracts(30)]);
      setList(res.list || []);
      setTotal(res.total);
      setExpiring(soon || []);
    } finally {
      setLoading(false);
    }
  }, [page]);

  useEffect(() => { void load(); }, [load]);
  useEffect(() => { void getContractTemplates().then(setTpls).catch(() => undefined); }, []);

  const typeLabel = (v: number) => t(`contract.type.${v}`);
  const statusLabel = (v: number) => t(`contract.status.${v}`);

  const columns: ColumnsType<ContractItem> = [
    { title: t('contract.title'), dataIndex: 'title' },
    { title: t('contract.party'), dataIndex: 'party_name', width: 140 },
    { title: t('contract.typeLabel'), dataIndex: 'contract_type', width: 100, render: typeLabel },
    { title: t('contract.statusLabel'), dataIndex: 'status', width: 90, render: statusLabel },
    { title: t('contract.amount'), dataIndex: 'amount', width: 100, render: (v?: number) => (v == null ? '-' : v.toFixed(2)) },
    { title: t('contract.expiredAt'), dataIndex: 'expired_at', width: 120, render: (v?: string) => v?.slice(0, 10) || '-' },
    { title: t('contract.file'), dataIndex: 'file_name', width: 120, render: (v?: string, r?: ContractItem) => (r?.file_id ? (v || r.file_id) : '-') },
    {
      title: t('common.actions'),
      width: 140,
      render: (_, record) => (
        <Space>
          {canManage && (
            <Button type="link" size="small" onClick={() => {
              setEditing(record);
              form.setFieldsValue({
                ...record,
                start_at: record.start_at ? dayjs(record.start_at.slice(0, 10)) : undefined,
                expired_at: record.expired_at ? dayjs(record.expired_at.slice(0, 10)) : undefined,
              });
              setFileList([]);
              setOpen(true);
            }}
            >{t('common.edit')}</Button>
          )}
          {canManage && record.status !== 1 && (
            <Button type="link" size="small" danger onClick={() => {
              void deleteContract(record.id).then(() => { message.success(t('common.deleted')); void load(); });
            }}
            >{t('common.delete')}</Button>
          )}
        </Space>
      ),
    },
  ];

  return (
    <div>
      {expiring.length > 0 && (
        <Card size="small" style={{ marginBottom: 16 }} title={t('contract.expiring')}>
          {expiring.map((c) => <div key={c.id}>{c.title} · {c.expired_at?.slice(0, 10)}</div>)}
        </Card>
      )}
      <Card
        title={t('contract.list')}
        extra={canCreate && (
          <Button type="primary" icon={<PlusOutlined />} onClick={() => { setEditing(null); form.resetFields(); setFileList([]); setOpen(true); }}>
            {t('contract.create')}
          </Button>
        )}
      >
        <Table rowKey="id" loading={loading} columns={columns} dataSource={list} pagination={{ current: page, pageSize: 10, total, onChange: setPage }} />
      </Card>
      <Modal open={open} title={editing ? t('common.edit') : t('contract.create')} onCancel={() => setOpen(false)} onOk={() => form.submit()} destroyOnClose width={560}>
        <Form
          form={form}
          layout="vertical"
          onFinish={async (values: {
            title: string; party_name: string; contract_type: number; amount?: number;
            template_id?: string; status?: number; start_at?: dayjs.Dayjs; expired_at?: dayjs.Dayjs; file_id?: string;
          }) => {
            let fileId = values.file_id || editing?.file_id;
            const raw = fileList[0]?.originFileObj;
            if (raw) {
              const fd = new FormData();
              fd.append('file', raw);
              const uploaded = await uploadFile(fd);
              fileId = uploaded.id;
            }
            const payload = {
              title: values.title,
              party_name: values.party_name,
              contract_type: values.contract_type,
              amount: values.amount,
              template_id: values.template_id,
              status: values.status,
              start_at: values.start_at ? `${values.start_at.format('YYYY-MM-DD')}T00:00:00Z` : undefined,
              expired_at: values.expired_at ? `${values.expired_at.format('YYYY-MM-DD')}T00:00:00Z` : undefined,
              file_id: fileId,
            };
            if (editing) await updateContract(editing.id, payload);
            else await createContract(payload);
            message.success(t('common.saved'));
            setOpen(false);
            void load();
          }}
        >
          <Form.Item name="title" label={t('contract.title')} rules={[{ required: true }]}><Input /></Form.Item>
          <Form.Item name="party_name" label={t('contract.party')} rules={[{ required: true }]}><Input /></Form.Item>
          <Form.Item name="contract_type" label={t('contract.typeLabel')} rules={[{ required: true }]}>
            <Select options={[1, 2, 3, 4].map((v) => ({ value: v, label: t(`contract.type.${v}`) }))} />
          </Form.Item>
          <Form.Item name="amount" label={t('contract.amount')}><InputNumber min={0} style={{ width: '100%' }} /></Form.Item>
          <Form.Item name="template_id" label={t('contract.template')}>
            <Select allowClear options={tpls.map((x) => ({ value: x.id, label: x.name }))} />
          </Form.Item>
          <Form.Item name="status" label={t('contract.statusLabel')} initialValue={0}>
            <Select options={[0, 1, 2, 3].map((v) => ({ value: v, label: t(`contract.status.${v}`) }))} />
          </Form.Item>
          <Form.Item name="start_at" label={t('contract.startAt')}><DatePicker style={{ width: '100%' }} /></Form.Item>
          <Form.Item name="expired_at" label={t('contract.expiredAt')}><DatePicker style={{ width: '100%' }} /></Form.Item>
          <Form.Item label={t('contract.file')}>
            <Upload
              accept=".pdf,.doc,.docx,.xls,.xlsx,.ppt,.pptx,.png,.jpg,.jpeg,.gif,.webp"
              fileList={fileList}
              beforeUpload={(file) => { setFileList([{ uid: file.uid, name: file.name, status: 'done', originFileObj: file }]); return false; }}
              onRemove={() => setFileList([])}
              maxCount={1}
            >
              <Button icon={<UploadOutlined />}>{t('contract.upload')}</Button>
            </Upload>
            <div style={{ color: '#888', marginTop: 8 }}>{t('contract.uploadHint')}</div>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default ContractPage;
