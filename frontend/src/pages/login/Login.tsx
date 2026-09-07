import React, { useState, useEffect } from 'react';
import { Form, Input, Button, Card, Tabs, message } from 'antd';
import { UserOutlined, LockOutlined, MailOutlined } from '@ant-design/icons';
import { useNavigate, useLocation } from 'react-router-dom';
import { useDispatch, useSelector } from 'react-redux';

import { login, selectIsAuthenticated } from '@/store/slices/authSlice';
import { fetchCurrentUser } from '@/store/slices/userSlice';
import { register } from '@/api/auth';
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
    if (from?.pathname) return from.pathname;
  }
  return '/dashboard';
}

function getErrorMessage(error: unknown, fallback: string): string {
  if (typeof error === 'string' && error) return error;
  if (error instanceof Error && error.message) return error.message;
  return fallback;
}

const Login: React.FC = () => {
  const { t } = useTranslation();
  const [activeTab, setActiveTab] = useState('login');
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();
  const location = useLocation();
  const dispatch = useDispatch<AppDispatch>();
  const isAuthenticated = useSelector(selectIsAuthenticated);

  // 如果已登录，跳转到首页
  useEffect(() => {
    if (isAuthenticated) {
      navigate(getRedirectPath(location.state), { replace: true });
    }
  }, [isAuthenticated, navigate, location.state]);

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
          <h1>StarByte</h1>
          <p>{t('login.brand')}</p>
          <p style={{ marginTop: 12, fontSize: 16 }}>{t('login.recruit')}</p>
        </div>
      </div>
      <div className={styles.right}>
        <Card className={styles.card}>
          <Tabs
            activeKey={activeTab}
            onChange={setActiveTab}
            items={tabItems}
            centered
            size="large"
          />

          {activeTab === 'login' && (
            <Form
              name="login"
              onFinish={handleLogin}
              size="large"
              autoComplete="off"
              initialValues={{ username: '', password: '' }}
            >
              <Form.Item
                name="username"
                rules={[
                  { required: true, message: t('login.usernameRequired') },
                  { min: 3, message: '至少 3 个字符' },
                ]}
              >
                <Input prefix={<UserOutlined />} placeholder={t('login.usernameOrStudent')} />
              </Form.Item>

              <Form.Item
                name="password"
                rules={[
                  { required: true, message: t('login.passwordRequired') },
                  { min: 6, message: t('login.passwordRequired') },
                ]}
              >
                <Input.Password prefix={<LockOutlined />} placeholder={t('login.password')} />
              </Form.Item>

              <Form.Item>
                <Button type="primary" htmlType="submit" loading={loading} block>
                  {t('login.submit')}
                </Button>
              </Form.Item>

              <div style={{ textAlign: 'center', color: '#999' }}>
                {t('login.hint')}
                <a onClick={() => setActiveTab('register')}>{t('login.goRegister')}</a>
              </div>
            </Form>
          )}

          {activeTab === 'register' && (
            <Form
              name="register"
              onFinish={handleRegister}
              size="large"
              autoComplete="off"
            >
              <Form.Item
                name="username"
                rules={[
                  { required: true, message: '请输入用户名' },
                  { min: 3, max: 20, message: '用户名长度为3-20个字符' },
                  { pattern: /^[a-zA-Z0-9_]+$/, message: '用户名只能包含字母、数字和下划线' },
                ]}
              >
                <Input prefix={<UserOutlined />} placeholder={t('login.username')} />
              </Form.Item>

              <Form.Item
                name="real_name"
                rules={[{ required: true, message: '请输入真实姓名' }]}
              >
                <Input placeholder={t('login.realName')} />
              </Form.Item>

              <Form.Item
                name="email"
                rules={[
                  { required: true, message: '请输入邮箱' },
                  { type: 'email', message: '请输入有效的邮箱地址' },
                ]}
              >
                <Input prefix={<MailOutlined />} placeholder={t('login.email')} />
              </Form.Item>

              <Form.Item
                name="password"
                rules={[
                  { required: true, message: '请输入密码' },
                  { min: 6, message: '密码至少6个字符' },
                ]}
              >
                <Input.Password prefix={<LockOutlined />} placeholder={t('login.password')} />
              </Form.Item>

              <Form.Item
                name="confirm_password"
                dependencies={['password']}
                rules={[
                  { required: true, message: t('login.passwordRequired') },
                  ({ getFieldValue }) => ({
                    validator(_, value) {
                      if (!value || getFieldValue('password') === value) {
                        return Promise.resolve();
                      }
                      return Promise.reject(new Error('两次输入的密码不一致'));
                    },
                  }),
                ]}
              >
                <Input.Password prefix={<LockOutlined />} placeholder={t('login.confirmPassword')} />
              </Form.Item>

              <Form.Item>
                <Button type="primary" htmlType="submit" loading={loading} block>
                  {t('login.register')}
                </Button>
              </Form.Item>

              <div style={{ textAlign: 'center', color: '#999' }}>
                {t('login.hasAccount')}
                <a onClick={() => setActiveTab('login')}>{t('login.goLogin')}</a>
              </div>
            </Form>
          )}
        </Card>
      </div>
    </div>
  );
};

export default Login;
