import React, { useCallback, useEffect, useState } from 'react';
import { Button, Card, Form, Input, InputNumber, Select, Space, Tag, message } from 'antd';
import { useTranslation } from 'react-i18next';
import { getSMTPSettings, testSMTPSettings, updateSMTPSettings } from '@/api/config';
import { usePermission } from '@/hooks/usePermission';
import type { SMTPSSLMode, SMTPSettings } from '@/types/api';

interface FormValues {
  host: string;
  port: number;
  ssl_mode: SMTPSSLMode;
  from: string;
  from_name: string;
  username?: string;
  test_to?: string;
}

const SmtpCard: React.FC = () => {
  const { t } = useTranslation();
  const canUpdate = usePermission('config:update');
  const [form] = Form.useForm<FormValues>();
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [testing, setTesting] = useState(false);
  const [settings, setSettings] = useState<SMTPSettings | null>(null);

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const data = await getSMTPSettings();
      setSettings(data);
      form.setFieldsValue({
        host: data.host,
        port: data.port,
        ssl_mode: data.ssl_mode,
        from: data.from,
        from_name: data.from_name,
        username: data.username,
      });
    } finally {
      setLoading(false);
    }
  }, [form]);

  useEffect(() => { void load(); }, [load]);

  const save = async () => {
    const values = await form.validateFields(['host', 'port', 'ssl_mode', 'from', 'from_name', 'username']);
    setSaving(true);
    try {
      const data = await updateSMTPSettings({
        host: values.host,
        port: values.port,
        ssl_mode: values.ssl_mode,
        from: values.from,
        from_name: values.from_name,
        username: values.username,
      });
      setSettings(data);
      message.success(t('smtp.saved'));
    } finally {
      setSaving(false);
    }
  };

  const sendTest = async () => {
    const to = String(form.getFieldValue('test_to') || '').trim();
    if (!to) {
      message.warning(t('smtp.testToRequired'));
      return;
    }
    setTesting(true);
    try {
      await testSMTPSettings({ to });
      message.success(t('smtp.testOk'));
    } finally {
      setTesting(false);
    }
  };

  return (
    <Card className="page-shell smtp-card" loading={loading} title={t('smtp.title')}>
      <p className="smtp-desc">{t('smtp.desc')}</p>
      <Form form={form} layout="vertical">
        <div className="smtp-grid">
          <Form.Item name="host" label={t('smtp.host')} rules={[{ required: true }]}>
            <Input placeholder="smtp.exmail.qq.com" disabled={!canUpdate} />
          </Form.Item>
          <Form.Item name="port" label={t('smtp.port')} rules={[{ required: true }]}>
            <InputNumber min={1} max={65535} style={{ width: '100%' }} disabled={!canUpdate} />
          </Form.Item>
          <Form.Item name="ssl_mode" label={t('smtp.sslMode')} rules={[{ required: true }]}>
            <Select
              disabled={!canUpdate}
              options={[
                { value: 'implicit', label: t('smtp.sslImplicit') },
                { value: 'starttls', label: t('smtp.sslStarttls') },
                { value: 'none', label: t('smtp.sslNone') },
              ]}
            />
          </Form.Item>
          <Form.Item name="from" label={t('smtp.from')} rules={[{ required: true, type: 'email' }]}>
            <Input disabled={!canUpdate} />
          </Form.Item>
          <Form.Item name="from_name" label={t('smtp.fromName')} rules={[{ required: true }]}>
            <Input disabled={!canUpdate} />
          </Form.Item>
          <Form.Item name="username" label={t('smtp.username')}>
            <Input placeholder={t('smtp.usernameHint') || undefined} disabled={!canUpdate} />
          </Form.Item>
        </div>
        <Form.Item label={t('smtp.password')}>
          <Space wrap>
            <Input.Password
              value={settings?.password_configured ? '********' : ''}
              readOnly
              visibilityToggle={false}
              placeholder={t('smtp.passwordPlaceholder') || undefined}
              style={{ width: 280 }}
            />
            <Tag color={settings?.password_configured ? 'green' : 'orange'}>
              {settings?.password_configured ? t('smtp.passwordSet') : t('smtp.passwordMissing')}
            </Tag>
          </Space>
          <div className="smtp-hint">{t('smtp.passwordHint')}</div>
        </Form.Item>
        {canUpdate && (
          <Space wrap>
            <Button type="primary" loading={saving} onClick={() => void save()}>{t('smtp.save')}</Button>
            <Form.Item name="test_to" style={{ marginBottom: 0 }}>
              <Input placeholder={t('smtp.testTo') || undefined} style={{ width: 240 }} />
            </Form.Item>
            <Button loading={testing} onClick={() => void sendTest()}>{t('smtp.testSend')}</Button>
          </Space>
        )}
      </Form>
    </Card>
  );
};

export default SmtpCard;
