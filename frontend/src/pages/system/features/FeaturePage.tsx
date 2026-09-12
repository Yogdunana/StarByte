import React, { useCallback, useEffect, useState } from 'react';
import {
  Button, Card, Drawer, Form, Input, InputNumber, Modal, Select, Space, Switch,
  Table, Tabs, Tag, message,
} from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { PlusOutlined, ReloadOutlined } from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import {
  createFeature, evaluateFeature, listFeatureAudits, listFeatures, toggleFeature, updateFeature,
  type FeatureAudit, type FeatureFlag, type FeatureType,
} from '@/api/feature';
import { usePermissions } from '@/hooks/usePermission';
import { formatDateTime } from '@/utils/format';
import './feature.css';

const types: FeatureType[] = ['boolean', 'user_allowlist', 'role_dept', 'percentage'];

interface FlagForm {
  flag_key: string;
  name: string;
  description?: string;
  flag_type: FeatureType;
  enabled?: boolean;
  group_name?: string;
  priority?: number;
  user_ids?: string;
  role_codes?: string;
  department_ids?: string;
  percent?: number;
  salt?: string;
}

function splitLines(raw?: string): string[] {
  return (raw || '').split(/[\n,]/).map((s) => s.trim()).filter(Boolean);
}

function toRules(v: FlagForm) {
  return {
    user_ids: splitLines(v.user_ids),
    role_codes: splitLines(v.role_codes),
    department_ids: splitLines(v.department_ids),
    percent: v.percent ?? 0,
    salt: v.salt || '',
  };
}

const FeaturePage: React.FC = () => {
  const { t } = useTranslation();
  const [canCreate, canUpdate, canManage] = usePermissions(['feature:create', 'feature:update', 'feature:manage']);
  const [list, setList] = useState<FeatureFlag[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [keyword, setKeyword] = useState('');
  const [loading, setLoading] = useState(false);
  const [open, setOpen] = useState(false);
  const [editing, setEditing] = useState<FeatureFlag | null>(null);
  const [form] = Form.useForm<FlagForm>();
  const [audits, setAudits] = useState<FeatureAudit[]>([]);
  const [evalOpen, setEvalOpen] = useState<FeatureFlag | null>(null);
  const [evalUser, setEvalUser] = useState('');
  const [evalResult, setEvalResult] = useState('');

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const res = await listFeatures({ page, page_size: 10, keyword: keyword || undefined });
      setList(res?.list || []);
      setTotal(res?.total || 0);
    } finally {
      setLoading(false);
    }
  }, [page, keyword]);

  const loadAudits = useCallback(async () => {
    const res = await listFeatureAudits({ page: 1, page_size: 20 });
    setAudits(res?.list || []);
  }, []);

  useEffect(() => { void load(); }, [load]);
  useEffect(() => { void loadAudits(); }, [loadAudits]);

  const onSave = async () => {
    const v = await form.validateFields();
    const payload = {
      name: v.name,
      description: v.description,
      flag_type: v.flag_type,
      enabled: v.enabled,
      group_name: v.group_name,
      priority: v.priority,
      rules: toRules(v),
    };
    if (editing) {
      await updateFeature(editing.id, payload);
      message.success(t('common.saved'));
    } else {
      await createFeature({ ...payload, flag_key: v.flag_key, enabled: v.enabled ?? false });
      message.success(t('feature.created'));
    }
    setOpen(false);
    setEditing(null);
    form.resetFields();
    void load();
    void loadAudits();
  };

  const openEdit = (row?: FeatureFlag) => {
    setEditing(row || null);
    form.setFieldsValue(row ? {
      flag_key: row.flag_key,
      name: row.name,
      description: row.description,
      flag_type: row.flag_type,
      enabled: row.enabled,
      group_name: row.group_name,
      priority: row.priority,
      user_ids: (row.rules.user_ids || []).join('\n'),
      role_codes: (row.rules.role_codes || []).join('\n'),
      department_ids: (row.rules.department_ids || []).join('\n'),
      percent: row.rules.percent,
      salt: row.rules.salt,
    } : { flag_type: 'boolean', enabled: false, priority: 0 });
    setOpen(true);
  };

  const columns: ColumnsType<FeatureFlag> = [
    { title: t('feature.key'), dataIndex: 'flag_key', render: (v: string) => <span className="feature-mono">{v}</span> },
    { title: t('feature.name'), dataIndex: 'name' },
    { title: t('feature.type'), dataIndex: 'flag_type', width: 140, render: (v: string) => <Tag>{t(`feature.types.${v}`)}</Tag> },
    {
      title: t('feature.enabled'),
      dataIndex: 'enabled',
      width: 100,
      render: (v: boolean, row) => (
        <Switch
          checked={v}
          disabled={!canManage}
          onChange={async (checked) => {
            await toggleFeature(row.id, checked, 'admin-ui');
            message.success(t('feature.toggled'));
            void load();
            void loadAudits();
          }}
        />
      ),
    },
    { title: t('feature.group'), dataIndex: 'group_name', width: 120 },
    {
      title: t('common.actions'),
      width: 180,
      render: (_, row) => (
        <Space>
          {canUpdate && <Button type="link" size="small" onClick={() => openEdit(row)}>{t('common.edit')}</Button>}
          <Button type="link" size="small" onClick={() => { setEvalOpen(row); setEvalResult(''); }}>{t('feature.evaluate')}</Button>
        </Space>
      ),
    },
  ];

  return (
    <div className="feature-page">
      <Card>
        <p className="feature-hint">{t('feature.intro')}</p>
        <div className="feature-toolbar">
          <Space>
            <Input.Search allowClear placeholder={t('feature.search')} onSearch={(v) => { setPage(1); setKeyword(v); }} />
            <Button icon={<ReloadOutlined />} onClick={() => { void load(); void loadAudits(); }} />
          </Space>
          {canCreate && <Button type="primary" icon={<PlusOutlined />} onClick={() => openEdit()}>{t('feature.create')}</Button>}
        </div>
        <Tabs
          items={[
            {
              key: 'flags',
              label: t('feature.list'),
              children: (
                <Table
                  rowKey="id"
                  loading={loading}
                  columns={columns}
                  dataSource={list}
                  pagination={{ current: page, total, pageSize: 10, onChange: setPage }}
                />
              ),
            },
            {
              key: 'audit',
              label: t('feature.audit'),
              children: (
                <Table
                  rowKey="id"
                  dataSource={audits}
                  pagination={false}
                  columns={[
                    { title: t('feature.key'), dataIndex: 'flag_key' },
                    { title: t('feature.action'), dataIndex: 'action', width: 100 },
                    { title: t('feature.reason'), dataIndex: 'reason' },
                    { title: t('feature.time'), dataIndex: 'created_at', render: (v: string) => formatDateTime(v) },
                  ]}
                />
              ),
            },
          ]}
        />
      </Card>

      <Modal
        open={open}
        title={editing ? t('feature.edit') : t('feature.create')}
        onCancel={() => { setOpen(false); setEditing(null); }}
        onOk={() => void onSave()}
        destroyOnClose
        width={640}
      >
        <Form form={form} layout="vertical">
          <Form.Item name="flag_key" label={t('feature.key')} rules={[{ required: true }]}>
            <Input disabled={Boolean(editing)} placeholder="cms.public" />
          </Form.Item>
          <Form.Item name="name" label={t('feature.name')} rules={[{ required: true }]}><Input /></Form.Item>
          <Form.Item name="description" label={t('feature.description')}><Input.TextArea rows={2} /></Form.Item>
          <Form.Item name="flag_type" label={t('feature.type')} rules={[{ required: true }]}>
            <Select options={types.map((value) => ({ value, label: t(`feature.types.${value}`) }))} />
          </Form.Item>
          <Form.Item name="group_name" label={t('feature.group')}><Input /></Form.Item>
          <Form.Item name="priority" label={t('feature.priority')}><InputNumber min={0} style={{ width: '100%' }} /></Form.Item>
          <Form.Item name="enabled" label={t('feature.enabled')} valuePropName="checked"><Switch /></Form.Item>
          <Form.Item name="user_ids" label={t('feature.userIds')}><Input.TextArea rows={3} placeholder="uuid, one per line" /></Form.Item>
          <Form.Item name="role_codes" label={t('feature.roleCodes')}><Input.TextArea rows={2} placeholder="minister" /></Form.Item>
          <Form.Item name="department_ids" label={t('feature.deptIds')}><Input.TextArea rows={2} /></Form.Item>
          <Form.Item name="percent" label={t('feature.percent')}><InputNumber min={0} max={100} style={{ width: '100%' }} /></Form.Item>
          <Form.Item name="salt" label={t('feature.salt')}><Input /></Form.Item>
        </Form>
      </Modal>

      <Drawer title={t('feature.evaluate')} open={Boolean(evalOpen)} onClose={() => setEvalOpen(null)}>
        <Space direction="vertical" style={{ width: '100%' }}>
          <Input placeholder={t('feature.evalUser')} value={evalUser} onChange={(e) => setEvalUser(e.target.value)} />
          <Button
            type="primary"
            onClick={async () => {
              if (!evalOpen) return;
              const res = await evaluateFeature(evalOpen.id, evalUser || undefined);
              setEvalResult(`${res.enabled ? 'ON' : 'OFF'} · ${res.reason}`);
            }}
          >
            {t('feature.evaluate')}
          </Button>
          {evalResult && <Tag color="blue">{evalResult}</Tag>}
        </Space>
      </Drawer>
    </div>
  );
};

export default FeaturePage;
