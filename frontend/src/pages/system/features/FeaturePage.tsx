import React, { useCallback, useEffect, useState } from 'react';
import {
  Button, Card, DatePicker, Drawer, Form, Input, InputNumber, Modal, Select, Space, Switch,
  Table, Tabs, Tag, message,
} from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { PlusOutlined, ReloadOutlined } from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import {
  createFeature, evaluateFeature, getFeatureAnalytics, listFeatureAudits, listFeatures,
  rollbackFeature, toggleFeature, updateFeature,
  type FeatureAnalytics, type FeatureAudit, type FeatureFlag,
} from '@/api/feature';
import { usePermissions } from '@/hooks/usePermission';
import { formatDateTime } from '@/utils/format';
import { FEATURE_ENVS, FEATURE_TYPES, analyticsRowKey, flagToForm, toRules, type FlagForm } from './form';
import './feature.css';

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
  const flagType = Form.useWatch('flag_type', form);
  const [audits, setAudits] = useState<FeatureAudit[]>([]);
  const [evalOpen, setEvalOpen] = useState<FeatureFlag | null>(null);
  const [evalUser, setEvalUser] = useState('');
  const [evalResult, setEvalResult] = useState('');
  const [statsOpen, setStatsOpen] = useState<FeatureFlag | null>(null);
  const [stats, setStats] = useState<FeatureAnalytics | null>(null);

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
    form.setFieldsValue(flagToForm(row));
    setOpen(true);
  };

  const openStats = async (row: FeatureFlag) => {
    setStatsOpen(row);
    setStats(null);
    const res = await getFeatureAnalytics(row.id, 7);
    setStats(res);
  };

  const onRollback = (row: FeatureFlag) => {
    Modal.confirm({
      title: t('feature.rollback'),
      content: t('feature.rollbackHint'),
      onOk: async () => {
        await rollbackFeature(row.id);
        message.success(t('feature.rolledBack'));
        void load();
        void loadAudits();
      },
    });
  };

  const scheduleColor = (state?: string) => {
    if (state === 'pending') return 'gold';
    if (state === 'expired') return 'default';
    if (state === 'active') return 'green';
    return 'blue';
  };

  const columns: ColumnsType<FeatureFlag> = [
    { title: t('feature.key'), dataIndex: 'flag_key', render: (v: string) => <span className="feature-mono">{v}</span> },
    { title: t('feature.name'), dataIndex: 'name' },
    { title: t('feature.type'), dataIndex: 'flag_type', width: 140, render: (v: string) => <Tag>{t(`feature.types.${v}`)}</Tag> },
    {
      title: t('feature.enabled'),
      dataIndex: 'enabled',
      width: 90,
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
    {
      title: t('feature.effective'),
      dataIndex: 'effective_enabled',
      width: 90,
      render: (v: boolean) => <Tag color={v ? 'green' : 'default'}>{v ? 'ON' : 'OFF'}</Tag>,
    },
    {
      title: t('feature.schedule'),
      dataIndex: 'schedule_state',
      width: 110,
      render: (v?: string) => (v ? <Tag color={scheduleColor(v)}>{t(`feature.scheduleStates.${v}`)}</Tag> : '—'),
    },
    { title: t('feature.group'), dataIndex: 'group_name', width: 100 },
    {
      title: t('common.actions'),
      width: 280,
      render: (_, row) => (
        <Space wrap>
          {canUpdate && <Button type="link" size="small" onClick={() => openEdit(row)}>{t('common.edit')}</Button>}
          <Button type="link" size="small" onClick={() => { setEvalOpen(row); setEvalResult(''); }}>{t('feature.evaluate')}</Button>
          <Button type="link" size="small" onClick={() => void openStats(row)}>{t('feature.analytics')}</Button>
          {canUpdate && <Button type="link" size="small" onClick={() => onRollback(row)}>{t('feature.rollback')}</Button>}
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
                  scroll={{ x: 1100 }}
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
                    { title: t('feature.action'), dataIndex: 'action', width: 120 },
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
        width={720}
      >
        <Form form={form} layout="vertical">
          <Form.Item name="flag_key" label={t('feature.key')} rules={[{ required: true }]}>
            <Input disabled={Boolean(editing)} placeholder="cms.public" />
          </Form.Item>
          <Form.Item name="name" label={t('feature.name')} rules={[{ required: true }]}><Input /></Form.Item>
          <Form.Item name="description" label={t('feature.description')}><Input.TextArea rows={2} /></Form.Item>
          <Form.Item name="flag_type" label={t('feature.type')} rules={[{ required: true }]}>
            <Select
              options={FEATURE_TYPES.map((value) => ({ value, label: t(`feature.types.${value}`) }))}
              onChange={(value) => {
                if (value === 'ab_test' && !(form.getFieldValue('variants') || []).length) {
                  form.setFieldValue('variants', [
                    { key: 'control', weight: 50, enabled: false },
                    { key: 'treatment', weight: 50, enabled: true },
                  ]);
                }
              }}
            />
          </Form.Item>
          <Form.Item name="group_name" label={t('feature.group')}><Input /></Form.Item>
          <Form.Item name="priority" label={t('feature.priority')}><InputNumber min={0} style={{ width: '100%' }} /></Form.Item>
          <Form.Item name="enabled" label={t('feature.enabled')} valuePropName="checked"><Switch /></Form.Item>
          <Form.Item name="environments" label={t('feature.environments')}>
            <Select
              mode="multiple"
              allowClear
              placeholder={t('feature.environmentsAll')}
              options={FEATURE_ENVS.map((value) => ({ value, label: t(`feature.envs.${value}`) }))}
            />
          </Form.Item>
          <Form.Item name="starts_at" label={t('feature.startsAt')}><DatePicker showTime style={{ width: '100%' }} /></Form.Item>
          <Form.Item name="ends_at" label={t('feature.endsAt')}><DatePicker showTime style={{ width: '100%' }} /></Form.Item>
          {flagType === 'user_allowlist' && (
            <Form.Item name="user_ids" label={t('feature.userIds')}><Input.TextArea rows={3} placeholder="uuid, one per line" /></Form.Item>
          )}
          {flagType === 'role_dept' && (
            <>
              <Form.Item name="role_codes" label={t('feature.roleCodes')}><Input.TextArea rows={2} placeholder="minister" /></Form.Item>
              <Form.Item name="department_ids" label={t('feature.deptIds')}><Input.TextArea rows={2} /></Form.Item>
            </>
          )}
          {flagType === 'percentage' && (
            <Form.Item name="percent" label={t('feature.percent')}><InputNumber min={0} max={100} style={{ width: '100%' }} /></Form.Item>
          )}
          {(flagType === 'percentage' || flagType === 'ab_test') && (
            <Form.Item name="salt" label={t('feature.salt')}><Input /></Form.Item>
          )}
          {flagType === 'ab_test' && (
            <Form.List name="variants">
              {(fields, { add, remove }) => (
                <div className="feature-variants">
                  {fields.map((field) => (
                    <Space key={field.key} align="baseline" className="feature-variant-row">
                      <Form.Item name={[field.name, 'key']} rules={[{ required: true }]}>
                        <Input placeholder={t('feature.variant')} />
                      </Form.Item>
                      <Form.Item name={[field.name, 'weight']}>
                        <InputNumber min={0} placeholder={t('feature.weight')} />
                      </Form.Item>
                      <Form.Item name={[field.name, 'enabled']} valuePropName="checked">
                        <Switch checkedChildren="ON" unCheckedChildren="OFF" />
                      </Form.Item>
                      {fields.length > 2 && <Button type="link" onClick={() => remove(field.name)}>{t('common.delete')}</Button>}
                    </Space>
                  ))}
                  <Button type="dashed" onClick={() => add({ key: '', weight: 0, enabled: true })}>{t('feature.addVariant')}</Button>
                </div>
              )}
            </Form.List>
          )}
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
              const variant = res.variant ? ` · ${res.variant}` : '';
              setEvalResult(`${res.enabled ? 'ON' : 'OFF'} · ${res.reason}${variant}`);
            }}
          >
            {t('feature.evaluate')}
          </Button>
          {evalResult && <Tag color="blue">{evalResult}</Tag>}
        </Space>
      </Drawer>

      <Drawer title={t('feature.analytics')} open={Boolean(statsOpen)} onClose={() => { setStatsOpen(null); setStats(null); }}>
        {stats ? (
          <Space direction="vertical" style={{ width: '100%' }}>
            <div>{stats.flag_key}</div>
            <div>{t('feature.exposureTotal')}: {stats.total} / {stats.days}d</div>
            <Table
              rowKey={analyticsRowKey}
              pagination={false}
              size="small"
              dataSource={stats.variants || []}
              columns={[
                { title: t('feature.variant'), dataIndex: 'variant', render: (v: string) => v || '—' },
                { title: t('feature.exposureCount'), dataIndex: 'count' },
                { title: t('feature.enabled'), dataIndex: 'enabled', render: (v: boolean) => (v ? 'ON' : 'OFF') },
              ]}
            />
          </Space>
        ) : <div>{t('common.loading')}</div>}
      </Drawer>
    </div>
  );
};

export default FeaturePage;
