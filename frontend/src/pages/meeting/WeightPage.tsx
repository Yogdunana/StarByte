import React, { useEffect, useState } from 'react';
import { Button, Card, Form, InputNumber, Space, message } from 'antd';
import { usePermission } from '@/hooks/usePermission';
import { getVoteWeightConfig, updateVoteWeightConfig } from '@/api/meeting';

const fields = [
  { key: 'president', label: '社长' },
  { key: 'vice_president', label: '副社长' },
  { key: 'minister', label: '部长' },
  { key: 'deputy', label: '副部长' },
  { key: 'officer', label: '干事' },
];

const WeightPage: React.FC = () => {
  const canEdit = usePermission('system:config');
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    getVoteWeightConfig()
      .then((cfg) => {
        form.setFieldsValue({
          default_weight: cfg.default_weight,
          ...cfg.weights,
        });
      })
      .finally(() => setLoading(false));
  }, [form]);

  return (
    <Card title="投票权重配置" loading={loading}>
      <p>加权投票按职务读取此处配置，不硬编码。社长可用系统配置权限修改。</p>
      <Form
        form={form}
        layout="vertical"
        disabled={!canEdit}
        onFinish={async (values) => {
          const weights: Record<string, number> = {};
          fields.forEach((f) => { weights[f.key] = Number(values[f.key]); });
          weights.vice_minister = weights.deputy;
          await updateVoteWeightConfig({
            weights,
            default_weight: Number(values.default_weight),
          });
          message.success('已保存');
        }}
      >
        {fields.map((f) => (
          <Form.Item key={f.key} name={f.key} label={f.label} rules={[{ required: true }]}>
            <InputNumber min={0.1} step={0.5} style={{ width: 200 }} />
          </Form.Item>
        ))}
        <Form.Item name="default_weight" label="默认权重" rules={[{ required: true }]}>
          <InputNumber min={0.1} step={0.5} style={{ width: 200 }} />
        </Form.Item>
        {canEdit && (
          <Space>
            <Button type="primary" htmlType="submit">保存</Button>
          </Space>
        )}
      </Form>
    </Card>
  );
};

export default WeightPage;
