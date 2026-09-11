import React, { useMemo, useState } from 'react';
import { Button, Card, Form, Input, message } from 'antd';
import { LockOutlined, MailOutlined, UserOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { useDispatch } from 'react-redux';
import { useTranslation } from 'react-i18next';

import { registerWithCasToken } from '@/api/auth';
import { setToken } from '@/store/slices/authSlice';
import { fetchCurrentUser } from '@/store/slices/userSlice';
import type { AppDispatch } from '@/store';
import { clearCASRegisterDraft, loadCASRegisterDraft } from './casRegisterDraft';
import styles from './Login.module.css';

function safeRedirect(path: string | undefined): string {
  if (!path || !path.startsWith('/') || path.startsWith('//')) {
    return '/dashboard';
  }
  return path;
}

interface RegisterFormValues {
  username: string;
  password: string;
  confirm_password: string;
  real_name: string;
  email: string;
}

const CasRegister: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const dispatch = useDispatch<AppDispatch>();
  const [loading, setLoading] = useState(false);
  const draft = useMemo(() => loadCASRegisterDraft(), []);

  if (!draft) {
    return (
      <div className={styles.container}>
        <div className={styles.right} style={{ gridColumn: '1 / -1' }}>
          <Card className={styles.card}>
            <p className={styles.formKicker}>CAMPUS BIND</p>
            <h2>{t('login.casBind')}</h2>
            <p className={styles.formHint}>{t('login.casBindNeedAuth')}</p>
            <Button type="primary" block onClick={() => navigate('/login', { replace: true })}>
              {t('login.goLogin')}
            </Button>
          </Card>
        </div>
      </div>
    );
  }

  const handleFinish = async (values: RegisterFormValues) => {
    setLoading(true);
    try {
      const result = await registerWithCasToken({
        token: draft.token,
        username: values.username,
        password: values.password,
        real_name: values.real_name,
        email: values.email,
      });
      clearCASRegisterDraft();
      dispatch(
        setToken({
          accessToken: result.access_token,
          refreshToken: result.refresh_token,
        }),
      );
      await dispatch(fetchCurrentUser()).unwrap();
      message.success(t('login.success'));
      navigate(safeRedirect(result.redirect || draft.redirect), { replace: true });
    } catch (error: unknown) {
      const msg = error instanceof Error && error.message ? error.message : t('login.registerFail');
      message.error(msg);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className={styles.container}>
      <div className={styles.right} style={{ gridColumn: '1 / -1' }}>
        <Card className={styles.card}>
          <p className={styles.formKicker}>CAMPUS BIND</p>
          <h2>{t('login.casBind')}</h2>
          <p className={styles.formHint}>{t('login.casBindHint')}</p>
          <Form
            name="cas-register"
            layout="vertical"
            size="large"
            onFinish={handleFinish}
            initialValues={{
              student_no: draft.student_no,
              real_name: draft.real_name,
              email: draft.email,
            }}
          >
            <Form.Item name="student_no" label={t('login.studentNo')}>
              <Input readOnly disabled />
            </Form.Item>
            <Form.Item
              name="username"
              label={t('login.username')}
              rules={[
                { required: true, message: t('login.usernameOnlyRequired') },
                { min: 3, max: 20, message: t('login.usernameLen') },
                { pattern: /^[a-zA-Z0-9_]+$/, message: t('login.usernamePattern') },
              ]}
            >
              <Input prefix={<UserOutlined />} autoComplete="username" placeholder={t('login.username')} />
            </Form.Item>
            <Form.Item
              name="real_name"
              label={t('login.realName')}
              rules={[{ required: true, message: t('login.realNameRequired') }]}
            >
              <Input placeholder={t('login.realName')} />
            </Form.Item>
            <Form.Item
              name="email"
              label={t('login.email')}
              rules={[{ type: 'email', message: t('login.emailInvalid') }]}
            >
              <Input prefix={<MailOutlined />} placeholder={t('login.email')} />
            </Form.Item>
            <Form.Item
              name="password"
              label={t('login.password')}
              rules={[
                { required: true, message: t('login.passwordRequired') },
                {
                  validator(_, value) {
                    if (!value) return Promise.resolve();
                    if (value.length >= 8 && /[A-Za-z]/.test(value) && /\d/.test(value)) {
                      return Promise.resolve();
                    }
                    return Promise.reject(new Error(t('login.passwordStrength')));
                  },
                },
              ]}
            >
              <Input.Password autoComplete="new-password" prefix={<LockOutlined />} placeholder={t('login.password')} />
            </Form.Item>
            <Form.Item
              name="confirm_password"
              label={t('login.confirmPassword')}
              dependencies={['password']}
              rules={[
                { required: true, message: t('login.passwordRequired') },
                ({ getFieldValue }) => ({
                  validator(_, value) {
                    if (!value || getFieldValue('password') === value) {
                      return Promise.resolve();
                    }
                    return Promise.reject(new Error(t('login.passwordMismatch')));
                  },
                }),
              ]}
            >
              <Input.Password prefix={<LockOutlined />} placeholder={t('login.confirmPassword')} />
            </Form.Item>
            <Form.Item>
              <Button type="primary" htmlType="submit" loading={loading} block>
                {t('login.casBindSubmit')}
              </Button>
            </Form.Item>
          </Form>
        </Card>
      </div>
    </div>
  );
};

export default CasRegister;
