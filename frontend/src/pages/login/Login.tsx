import React, { useState, useEffect } from 'react';
import { Form, Input, Button, Card, Tabs, Divider, message } from 'antd';
import { UserOutlined, LockOutlined, MailOutlined, BankOutlined } from '@ant-design/icons';
import { useNavigate, useLocation, useSearchParams } from 'react-router-dom';
import { useDispatch, useSelector } from 'react-redux';

import { login, selectIsAuthenticated } from '@/store/slices/authSlice';
import { fetchCurrentUser } from '@/store/slices/userSlice';
import { getCasLoginURL, getCasStatus, register } from '@/api/auth';
import { AppDispatch } from '@/store';
import styles from './Login.module.css';
import { useTranslation } from 'react-i18next';

interface LocationFromState {
  from?: { pathname?: string };
}

interface RegisterFormValues {
  username: string;
  password: string;
  confirm_password: string;
  real_name: string;
  email: string;
}

function getRedirectPath(state: unknown): string {
  if (state && typeof state === 'object' && 'from' in state) {
    const from = (state as LocationFromState).from;
    if (from?.pathname?.startsWith('/') && !from.pathname.startsWith('//')) return from.pathname;
  }
  return '/dashboard';
}

function getErrorMessage(error: unknown, fallback: string): string {
  if (typeof error === 'string' && error) return error;
  if (error instanceof Error && error.message) return error.message;
  return fallback;
}

const envCasEnabled = import.meta.env.VITE_CAS_ENABLED === 'true';

const Login: React.FC = () => {
  const { t } = useTranslation();
  const [activeTab, setActiveTab] = useState('login');
  const [loading, setLoading] = useState(false);
  const [casEnabled, setCasEnabled] = useState(envCasEnabled);
  const navigate = useNavigate();
  const location = useLocation();
  const [searchParams] = useSearchParams();
  const dispatch = useDispatch<AppDispatch>();
  const isAuthenticated = useSelector(selectIsAuthenticated);

  // 如果已登录，跳转到首页
  useEffect(() => {
    if (isAuthenticated) {
      navigate(getRedirectPath(location.state), { replace: true });
    }
  }, [isAuthenticated, navigate, location.state]);

  useEffect(() => {
    getCasStatus()
      .then((s) => setCasEnabled(Boolean(s?.enabled)))
      .catch(() => setCasEnabled(envCasEnabled));
  }, []);

  useEffect(() => {
    if (searchParams.get('cas_error')) {
      message.error(t('login.casError'));
    }
  }, [searchParams, t]);

  // 登录
  const handleLogin = async (values: { username: string; password: string }) => {
    setLoading(true);
    try {
      await dispatch(login(values)).unwrap();
      await dispatch(fetchCurrentUser()).unwrap();
      message.success(t('login.success'));
      navigate(getRedirectPath(location.state), { replace: true });
    } catch (error: unknown) {
      message.error(getErrorMessage(error, t('login.fail')));
    } finally {
      setLoading(false);
    }
  };

  // 注册
  const handleRegister = async (values: RegisterFormValues) => {
    setLoading(true);
    try {
      await register({
        username: values.username,
        password: values.password,
        real_name: values.real_name,
        email: values.email,
      });
      message.success(t('login.registerSuccess'));
      setActiveTab('login');
    } catch (error: unknown) {
      message.error(getErrorMessage(error, t('login.registerFail')));
    } finally {
      setLoading(false);
    }
  };

  const tabItems = [
    {
      key: 'login',
      label: t('login.tabLogin'),
    },
    {
      key: 'register',
      label: t('login.tabRegister'),
    },
  ];

  return (
    <div className={styles.container}>
      <div className={styles.left}>
        <div className={styles.brand}>
          <div className={styles.identity}>
            <div className={styles.wordmark}>StarByte<span>.</span></div>
            <p className={styles.kicker}>{t('login.kicker')}</p>
          </div>
          <div className={styles.message}>
            <h1>{t('login.headlineLine1')}<br />{t('login.headlineLine2')}</h1>
            <p className={styles.story}>{t('login.story')}</p>
          </div>
          <div className={styles.visual}>
            <div className={styles.connections}>
              <span className={styles.mark} aria-hidden="true"><i /><i /><i /><i /></span>
              <span>{t('login.pillars')}</span>
            </div>
            <p className={styles.caption}>{t('login.caption')}</p>
          </div>
        </div>
      </div>
      <div className={styles.right}>
        <div className={styles.mobileBrand}>StarByte.</div>
        <Card className={styles.card}>
          <header className={styles.cardHeader}>
            <p className={styles.formKicker}>{t('login.formKicker')}</p>
            <h2>{activeTab === 'login' ? t('login.welcomeBack') : t('login.joinUs')}</h2>
            <p className={styles.formHint}>
              {activeTab === 'login' ? t('login.formHintLogin') : t('login.formHintRegister')}
            </p>
          </header>
          <Tabs
            activeKey={activeTab}
            onChange={setActiveTab}
            items={tabItems}
            centered
            size="large"
          />

          {activeTab === 'login' && (
            <>
              {casEnabled && (
                <div className={styles.casBlock}>
                  <Button
                    block
                    size="large"
                    icon={<BankOutlined />}
                    className={styles.casBtn}
                    onClick={() => {
                      window.location.assign(getCasLoginURL(getRedirectPath(location.state)));
                    }}
                  >
                    {t('login.cas')}
                  </Button>
                  <p className={styles.casHint}>{t('login.casHint')}</p>
                </div>
              )}

              {casEnabled && <Divider plain className={styles.casDivider}>{t('login.orLocal')}</Divider>}

              <Form
                name="login"
                onFinish={handleLogin}
                size="large"
                layout="vertical"
                initialValues={{ username: '', password: '' }}
              >
                <Form.Item
                  name="username"
                  label={t('login.username')}
                  rules={[
                    { required: true, message: t('login.usernameRequired') },
                    { min: 3, message: t('login.minChars', { n: 3 }) },
                  ]}
                >
                  <Input autoComplete="username" prefix={<UserOutlined />} placeholder={t('login.usernameOrStudent')} />
                </Form.Item>

                <Form.Item
                  name="password"
                  label={t('login.password')}
                  rules={[
                    { required: true, message: t('login.passwordRequired') },
                    { min: 6, message: t('login.passwordMin') },
                  ]}
                >
                  <Input.Password autoComplete="current-password" prefix={<LockOutlined />} placeholder={t('login.password')} />
                </Form.Item>

                <Form.Item className={styles.submitItem}>
                  <Button type="primary" htmlType="submit" loading={loading} block>
                    {t('login.submit')}
                  </Button>
                </Form.Item>

                <div className={styles.switchTab}>
                  {t('login.hint')}
                  <Button type="link" onClick={() => setActiveTab('register')}>{t('login.goRegister')}</Button>
                </div>
              </Form>
            </>
          )}

          {activeTab === 'register' && (
            <Form
              name="register"
              onFinish={handleRegister}
              size="large"
              layout="vertical"
              className={styles.registerForm}
            >
              <div className={styles.fieldRow}>
                <Form.Item
                  name="username"
                  label={t('login.username')}
                  rules={[
                    { required: true, message: t('login.usernameOnlyRequired') },
                    { min: 3, max: 20, message: t('login.usernameLen') },
                    { pattern: /^[a-zA-Z0-9_]+$/, message: t('login.usernamePattern') },
                  ]}
                >
                  <Input prefix={<UserOutlined />} placeholder={t('login.username')} />
                </Form.Item>

                <Form.Item
                  name="real_name"
                  label={t('login.realName')}
                  rules={[{ required: true, message: t('login.realNameRequired') }]}
                >
                  <Input placeholder={t('login.realName')} />
                </Form.Item>
              </div>

              <Form.Item
                name="email"
                label={t('login.email')}
                rules={[
                  { required: true, message: t('login.emailRequired') },
                  { type: 'email', message: t('login.emailInvalid') },
                ]}
              >
                <Input prefix={<MailOutlined />} placeholder={t('login.email')} />
              </Form.Item>

              <div className={styles.fieldRow}>
                <Form.Item
                  name="password"
                  label={t('login.password')}
                  rules={[
                    { required: true, message: t('login.passwordRequired') },
                    { min: 6, message: t('login.passwordMin') },
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
              </div>

              <Form.Item className={styles.submitItem}>
                <Button type="primary" htmlType="submit" loading={loading} block>
                  {t('login.register')}
                </Button>
              </Form.Item>

              <div className={styles.switchTab}>
                {t('login.hasAccount')}
                <Button type="link" onClick={() => setActiveTab('login')}>{t('login.goLogin')}</Button>
              </div>
            </Form>
          )}
        </Card>
      </div>
    </div>
  );
};

export default Login;
