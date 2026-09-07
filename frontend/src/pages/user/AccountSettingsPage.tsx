import React, { useState } from 'react';
import { Button, Card, Form, Input, Radio, Select, Space, message } from 'antd';
import { useTranslation } from 'react-i18next';
import { changePassword } from '@/api/auth';
import { useThemeLang } from '@/theme/ThemeLangContext';
import type { AppLang } from '@/i18n';
import type { ThemePreference } from '@/theme/preference';

interface PasswordFormValues {
  old_password: string;
  new_password: string;
  confirm_password: string;
}

const AccountSettingsPage: React.FC = () => {
  const { t } = useTranslation();
  const { preference, setPreference, lang, setLang } = useThemeLang();
  const [form] = Form.useForm<PasswordFormValues>();
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (values: PasswordFormValues) => {
    setLoading(true);
    try {
      await changePassword({
        old_password: values.old_password,
        new_password: values.new_password,
      });
      message.success(t('settings.passwordUpdated'));
      form.resetFields();
    } catch {
      // interceptor
    } finally {
      setLoading(false);
    }
  };

  return (
    <Space direction="vertical" size={16} style={{ width: '100%' }}>
      <Card title={t('settings.appearance')}>
        <Space direction="vertical" size={12}>
          <div>
            <div style={{ marginBottom: 8 }}>{t('topbar.theme')}</div>
            <Radio.Group
              value={preference}
              onChange={(e) => setPreference(e.target.value as ThemePreference)}
              options={[
                { value: 'light', label: t('topbar.themeLight') },
                { value: 'dark', label: t('topbar.themeDark') },
                { value: 'system', label: t('topbar.themeSystem') },
              ]}
            />
          </div>
          <div>
            <div style={{ marginBottom: 8 }}>{t('settings.language')}</div>
            <Select
              style={{ width: 200 }}
              value={lang}
              onChange={(v: AppLang) => setLang(v)}
              options={[
                { value: 'zh-CN', label: '简体中文' },
                { value: 'en-US', label: 'English' },
              ]}
            />
          </div>
        </Space>
      </Card>
      <Card title={t('settings.title')}>
        <Form
          form={form}
          layout="vertical"
          style={{ maxWidth: 420 }}
          onFinish={(values) => void handleSubmit(values)}
        >
          <Form.Item name="old_password" label={t('settings.oldPassword')} rules={[{ required: true }]}>
            <Input.Password autoComplete="current-password" />
          </Form.Item>
          <Form.Item
            name="new_password"
            label={t('settings.newPassword')}
            rules={[
              { required: true },
              { min: 8 },
              { pattern: /^(?=.*[A-Za-z])(?=.*\d).+$/ },
            ]}
          >
            <Input.Password autoComplete="new-password" />
          </Form.Item>
          <Form.Item
            name="confirm_password"
            label={t('settings.confirmPassword')}
            dependencies={['new_password']}
            rules={[
              { required: true },
              ({ getFieldValue }) => ({
                validator(_, value: string) {
                  if (!value || getFieldValue('new_password') === value) {
                    return Promise.resolve();
                  }
                  return Promise.reject(new Error(t('login.confirmPassword')));
                },
              }),
            ]}
          >
            <Input.Password autoComplete="new-password" />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit" loading={loading}>
              {t('settings.save')}
            </Button>
          </Form.Item>
        </Form>
      </Card>
    </Space>
  );
};

export default AccountSettingsPage;
