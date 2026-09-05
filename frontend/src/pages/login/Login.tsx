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
      message.success('登录成功');
      navigate(getRedirectPath(location.state), { replace: true });
    } catch (error: unknown) {
      message.error(getErrorMessage(error, '登录失败'));
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
      message.success('注册成功，请登录');
      setActiveTab('login');
    } catch (error: unknown) {
      message.error(getErrorMessage(error, '注册失败'));
    } finally {
      setLoading(false);
    }
  };

  const tabItems = [
    {
      key: 'login',
      label: '登录',
    },
    {
      key: 'register',
      label: '注册',
    },
  ];

  return (
    <div className={styles.container}>
      <div className={styles.left}>
        <div className={styles.brand}>
          <h1>StarByte</h1>
          <p>计算机协会一体化管理平台</p>
          <p style={{ marginTop: 12, fontSize: 16 }}>2026 秋季招新进行中 · 登录后即可提交入会申请</p>
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
                  { required: true, message: '请输入用户名' },
                  { min: 3, message: '用户名至少3个字符' },
                ]}
              >
                <Input prefix={<UserOutlined />} placeholder="用户名" />
              </Form.Item>

              <Form.Item
                name="password"
                rules={[
                  { required: true, message: '请输入密码' },
                  { min: 6, message: '密码至少6个字符' },
                ]}
              >
                <Input.Password prefix={<LockOutlined />} placeholder="密码" />
              </Form.Item>

              <Form.Item>
                <Button type="primary" htmlType="submit" loading={loading} block>
                  登录
                </Button>
              </Form.Item>

              <div style={{ textAlign: 'center', color: '#999' }}>
                还没有账号？
                <a onClick={() => setActiveTab('register')}>立即注册</a>
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
                <Input prefix={<UserOutlined />} placeholder="用户名" />
              </Form.Item>

              <Form.Item
                name="real_name"
                rules={[{ required: true, message: '请输入真实姓名' }]}
              >
                <Input placeholder="真实姓名" />
              </Form.Item>

              <Form.Item
                name="email"
                rules={[
                  { required: true, message: '请输入邮箱' },
                  { type: 'email', message: '请输入有效的邮箱地址' },
                ]}
              >
                <Input prefix={<MailOutlined />} placeholder="邮箱" />
              </Form.Item>

              <Form.Item
                name="password"
                rules={[
                  { required: true, message: '请输入密码' },
                  { min: 6, message: '密码至少6个字符' },
                ]}
              >
                <Input.Password prefix={<LockOutlined />} placeholder="密码" />
              </Form.Item>

              <Form.Item
                name="confirm_password"
                dependencies={['password']}
                rules={[
                  { required: true, message: '请确认密码' },
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
                <Input.Password prefix={<LockOutlined />} placeholder="确认密码" />
              </Form.Item>

              <Form.Item>
                <Button type="primary" htmlType="submit" loading={loading} block>
                  注册
                </Button>
              </Form.Item>

              <div style={{ textAlign: 'center', color: '#999' }}>
                已有账号？
                <a onClick={() => setActiveTab('login')}>立即登录</a>
              </div>
            </Form>
          )}
        </Card>
      </div>
    </div>
  );
};

export default Login;
