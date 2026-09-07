import React, { useState } from 'react';
import { Button, Card, Form, Input, message } from 'antd';
import { changePassword } from '@/api/auth';

interface PasswordFormValues {
  old_password: string;
  new_password: string;
  confirm_password: string;
}

function getErrorMessage(error: unknown, fallback: string): string {
  if (typeof error === 'string' && error) return error;
  if (error instanceof Error && error.message) return error.message;
  return fallback;
}

const AccountSettingsPage: React.FC = () => {
  const [form] = Form.useForm<PasswordFormValues>();
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (values: PasswordFormValues) => {
    setLoading(true);
    try {
      await changePassword({
        old_password: values.old_password,
        new_password: values.new_password,
      });
      message.success('密码已更新');
      form.resetFields();
    } catch (error: unknown) {
      message.error(getErrorMessage(error, '修改密码失败'));
    } finally {
      setLoading(false);
    }
  };

  return (
    <Card title="账号设置">
      <Form
        form={form}
        layout="vertical"
        style={{ maxWidth: 420 }}
        onFinish={(values) => void handleSubmit(values)}
      >
        <Form.Item
          name="old_password"
          label="原密码"
          rules={[{ required: true, message: '请输入原密码' }]}
        >
          <Input.Password autoComplete="current-password" />
        </Form.Item>
        <Form.Item
          name="new_password"
          label="新密码"
          rules={[
            { required: true, message: '请输入新密码' },
            { min: 8, message: '至少 8 位，需包含字母和数字' },
            {
              pattern: /^(?=.*[A-Za-z])(?=.*\d).+$/,
              message: '至少 8 位，需包含字母和数字',
            },
          ]}
        >
          <Input.Password autoComplete="new-password" />
        </Form.Item>
        <Form.Item
          name="confirm_password"
          label="确认新密码"
          dependencies={['new_password']}
          rules={[
            { required: true, message: '请再次输入新密码' },
            ({ getFieldValue }) => ({
              validator(_, value: string) {
                if (!value || getFieldValue('new_password') === value) {
                  return Promise.resolve();
                }
                return Promise.reject(new Error('两次输入的密码不一致'));
              },
            }),
          ]}
        >
          <Input.Password autoComplete="new-password" />
        </Form.Item>
        <Form.Item>
          <Button type="primary" htmlType="submit" loading={loading}>
            保存
          </Button>
        </Form.Item>
      </Form>
    </Card>
  );
};

export default AccountSettingsPage;
