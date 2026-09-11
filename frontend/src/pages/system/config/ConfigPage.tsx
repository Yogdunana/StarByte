import React, { useCallback, useEffect, useMemo, useState } from 'react';
import {
  Button, Card, Drawer, Form, Input, InputNumber, Modal, Select, Space, Switch, Table, Tag, message,
} from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { PlusOutlined } from '@ant-design/icons';
import {
  createRuntimeConfig, deleteRuntimeConfig, getRuntimeConfigs, updateRuntimeConfig,
} from '@/api/config';
import { usePermission } from '@/hooks/usePermission';
import type { CreateRuntimeConfigParams, RuntimeConfig, RuntimeConfigType } from '@/types/api';
import SmtpCard from './SmtpCard';
import './config.css';

const protectedKeys = new Set(['internship_config', 'vote_weight_config', 'smtp_settings']);
const hiddenKeys = new Set(['smtp_settings']);

const categoryOptions = [
  { value: 'system', label: '系统' },
  { value: 'business', label: '业务' },
  { value: 'notification', label: '通知' },
  { value: 'security', label: '安全' },
  { value: 'internship', label: '实习' },
  { value: 'meeting', label: '会议' },
];

const typeOptions: { value: RuntimeConfigType; label: string }[] = [
  { value: 'string', label: '字符串' },
  { value: 'number', label: '数字' },
  { value: 'boolean', label: '布尔' },
  { value: 'json', label: 'JSON' },
];

const categoryLabel = Object.fromEntries(categoryOptions.map((c) => [c.value, c.label]));

interface FormValues {
  config_key: string;
  config_type: RuntimeConfigType;
  category: string;
  description?: string;
  is_public?: boolean;
  text_value?: string;
  number_value?: number;
  bool_value?: boolean;
}

function valueFromForm(v: FormValues): string {
  if (v.config_type === 'boolean') return v.bool_value ? 'true' : 'false';
  if (v.config_type === 'number') return String(v.number_value ?? 0);
  return v.text_value ?? '';
}

function fillForm(row: RuntimeConfig): FormValues {
  return {
    config_key: row.config_key,
    config_type: row.config_type,
    category: row.category,
    description: row.description,
    is_public: row.is_public,
    text_value: row.config_type === 'string' || row.config_type === 'json' ? row.config_value : '',
    number_value: row.config_type === 'number' ? Number(row.config_value) : undefined,
    bool_value: row.config_type === 'boolean' ? row.config_value === 'true' || row.config_value === '1' : false,
  };
}

const ConfigPage: React.FC = () => {
  const canCreate = usePermission('config:create');
  const canUpdate = usePermission('config:update');
  const canDelete = usePermission('config:delete');
  const [list, setList] = useState<RuntimeConfig[]>([]);
  const [loading, setLoading] = useState(false);
  const [category, setCategory] = useState<string>();
  const [keyword, setKeyword] = useState('');
  const [open, setOpen] = useState(false);
  const [editing, setEditing] = useState<RuntimeConfig | null>(null);
  const [form] = Form.useForm<FormValues>();
  const watchType = Form.useWatch('config_type', form);

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const rows = await getRuntimeConfigs({ category, keyword: keyword || undefined }) || [];
      setList(rows.filter((row) => !hiddenKeys.has(row.config_key)));
    } finally {
      setLoading(false);
    }
  }, [category, keyword]);

  useEffect(() => { void load(); }, [load]);

  const openCreate = () => {
    setEditing(null);
    form.resetFields();
    form.setFieldsValue({ config_type: 'string', category: 'system', bool_value: false });
    setOpen(true);
  };

  const openEdit = (row: RuntimeConfig) => {
    setEditing(row);
    form.setFieldsValue(fillForm(row));
    setOpen(true);
  };

  const submit = async () => {
    const values = await form.validateFields();
    const payload: CreateRuntimeConfigParams = {
      config_key: values.config_key,
      config_type: values.config_type,
      category: values.category,
      description: values.description,
      is_public: values.is_public,
      config_value: valueFromForm(values),
    };
    if (editing) {
      await updateRuntimeConfig(editing.id, {
        config_value: payload.config_value,
        config_type: payload.config_type,
        category: payload.category,
        description: payload.description,
        is_public: payload.is_public,
      });
      message.success('已保存');
    } else {
      await createRuntimeConfig(payload);
      message.success('已创建');
    }
    setOpen(false);
    void load();
  };

  const columns: ColumnsType<RuntimeConfig> = useMemo(() => [
    { title: '键', dataIndex: 'config_key', width: 220 },
    { title: '分组', dataIndex: 'category', width: 90, render: (v: string) => categoryLabel[v] || v },
    { title: '类型', dataIndex: 'config_type', width: 80, render: (v: string) => <Tag>{v}</Tag> },
    {
      title: '值',
      dataIndex: 'config_value',
      ellipsis: true,
      render: (v: string) => <span className="cfg-value">{v}</span>,
    },
    { title: '说明', dataIndex: 'description', ellipsis: true },
    {
      title: '操作',
      width: 140,
      render: (_, row) => (
        <Space>
          {canUpdate && <Button type="link" size="small" onClick={() => openEdit(row)}>编辑</Button>}
          {canDelete && !protectedKeys.has(row.config_key) && (
            <Button type="link" size="small" danger onClick={() => {
              Modal.confirm({
                title: '删除配置',
                content: `确定删除 ${row.config_key}？`,
                onOk: async () => {
                  await deleteRuntimeConfig(row.id);
                  message.success('已删除');
                  void load();
                },
              });
            }}
            >
              删除
            </Button>
          )}
        </Space>
      ),
    },
  ], [canDelete, canUpdate, load]);

  return (
    <div>
      <div className="cfg-hero">
        <div>
          <h2>运行时配置</h2>
          <p>改完立即生效，不必重启服务。实习开关、投票权重等业务键不可删除。</p>
        </div>
        {canCreate && (
          <Button type="primary" size="large" icon={<PlusOutlined />} onClick={openCreate}>新建配置</Button>
        )}
      </div>
      <SmtpCard />
      <Card className="page-shell">
        <Space style={{ marginBottom: 16 }} wrap>
          <Select
            allowClear
            placeholder="分组"
            style={{ width: 140 }}
            value={category}
            options={categoryOptions}
            onChange={(v) => setCategory(v)}
          />
          <Input.Search allowClear placeholder="搜索键/说明" onSearch={setKeyword} style={{ width: 240 }} />
        </Space>
        <Table rowKey="id" loading={loading} columns={columns} dataSource={list} pagination={false} />
      </Card>
      <Drawer
        title={editing ? '编辑配置' : '新建配置'}
        open={open}
        onClose={() => setOpen(false)}
        extra={<Button type="primary" onClick={() => void submit()}>保存</Button>}
      >
        <Form form={form} layout="vertical">
          <Form.Item name="config_key" label="键" rules={[{ required: true }]}>
            <Input disabled={!!editing} placeholder="site.name" />
          </Form.Item>
          <Form.Item name="category" label="分组" rules={[{ required: true }]}>
            <Select options={categoryOptions} />
          </Form.Item>
          <Form.Item name="config_type" label="类型" rules={[{ required: true }]}>
            <Select options={typeOptions} disabled={!!editing} />
          </Form.Item>
          {watchType === 'boolean' && (
            <Form.Item name="bool_value" label="值" valuePropName="checked"><Switch /></Form.Item>
          )}
          {watchType === 'number' && (
            <Form.Item name="number_value" label="值" rules={[{ required: true }]}>
              <InputNumber style={{ width: '100%' }} />
            </Form.Item>
          )}
          {(watchType === 'string' || watchType === 'json' || !watchType) && (
            <Form.Item name="text_value" label="值">
              <Input.TextArea rows={watchType === 'json' ? 8 : 3} />
            </Form.Item>
          )}
          <Form.Item name="description" label="说明"><Input /></Form.Item>
          <Form.Item name="is_public" label="公开可读" valuePropName="checked"><Switch /></Form.Item>
        </Form>
      </Drawer>
    </div>
  );
};

export default ConfigPage;
