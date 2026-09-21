import LanguageSelect from '@/components/LanguageSelect/LanguageSelect';
import React, { useState, useEffect } from 'react';
import { Form, Input, Button, Card, Tabs, Divider, Select, message } from 'antd';
import { UserOutlined, LockOutlined, MailOutlined, BankOutlined } from '@ant-design/icons';
import { useNavigate, useLocation, useSearchParams, Link } from 'react-router-dom';
import { useDispatch, useSelector } from 'react-redux';

import { login, selectIsAuthenticated } from '@/store/slices/authSlice';
import { fetchCurrentUser } from '@/store/slices/userSlice';
import { getCasLoginURL, getCasStatus, register, resendVerification } from '@/api/auth';
import { apiErrorCode } from '@/api/error';
import { AppDispatch } from '@/store';
import { motion } from 'motion/react';
import styles from './Login.module.css';
import { fadeUp, staggerEnter } from '@/motion/tokens';
import { useTranslation } from 'react-i18next';
import { PUBLIC_CMS_KEYS } from '@/api/feature';
import { resolveRedirect, sanitizeNext } from '@/utils/nextPath';
import { FeatureProvider, useFeature } from '@/hooks/useFeature';

interface LocationFromState {
  from?: { pathname?: string };
}

interface RegisterFormValues {
  username: string;
  password: string;
  confirm_password: string;
  real_name: string;
  email: string;
  gender: 1 | 2;
}

function getRedirectPath(state: unknown, nextQuery?: string | null): string {
  const fromQuery = resolveRedirect(nextQuery, '');
  if (fromQuery) return fromQuery;
  if (state && typeof state === 'object' && 'from' in state) {
    const from = (state as LocationFromState).from;
    // 走同一个 sanitizeNext，别在这里另写一套弱判断
    // （原判断只拒了 //，漏了反斜杠与控制字符）。
    return sanitizeNext(from?.pathname) || '/dashboard';
  }
  return '/dashboard';
}

function getErrorMessage(error: unknown, fallback: string): string {
  if (typeof error === 'string' && error) return error;
  if (error instanceof Error && error.message) return error.message;
  return fallback;
}

const envCasEnabled = import.meta.env.VITE_CAS_ENABLED === 'true';

const LoginPublicFooter: React.FC = () => {
  const { t } = useTranslation();
  const { enabled, loading } = useFeature('cms.public');
  if (loading || !enabled) return null;
  return (
    <footer className={styles.footer}>
      <Link to="/about-us">{t('login.footerAbout')}</Link>
      <Link to="/docs/association-charter">{t('login.footerCharter')}</Link>
      <Link to="/docs">{t('login.footerDocs')}</Link>
    </footer>
  );
};

const Login: React.FC = () => {
  const { t } = useTranslation();
  const [activeTab, setActiveTab] = useState('login');
  const [loading, setLoading] = useState(false);
  const [needVerify, setNeedVerify] = useState(false);
  const [verifyId, setVerifyId] = useState('');
  const [resendLoading, setResendLoading] = useState(false);
  const [loginForm] = Form.useForm<{ username: string; password: string }>();
  const [casEnabled, setCasEnabled] = useState(envCasEnabled);
  const navigate = useNavigate();
  const location = useLocation();
  const [searchParams] = useSearchParams();
  const dispatch = useDispatch<AppDispatch>();
  const isAuthenticated = useSelector(selectIsAuthenticated);

  // 如果已登录，跳转到首页
  useEffect(() => {
    if (isAuthenticated) {
      navigate(getRedirectPath(location.state, searchParams.get('next')), { replace: true });
    }
  }, [isAuthenticated, navigate, location.state, searchParams]);

  useEffect(() => {
    getCasStatus()
      .then((s) => setCasEnabled(Boolean(s?.enabled)))
      .catch(() => setCasEnabled(envCasEnabled));
  }, []);

  useEffect(() => {
    const reason = searchParams.get('cas_error');
    if (reason === 'email_unverified') {
      message.error(t('login.casErrorEmail'));
      setNeedVerify(true);
    } else if (reason) {
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
      navigate(getRedirectPath(location.state, searchParams.get('next')), { replace: true });
    } catch (error: unknown) {
      if (apiErrorCode(error) === 2015) {
        setNeedVerify(true);
        setVerifyId(values.username);
      } else {
        message.error(getErrorMessage(error, t('login.fail')));
      }
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
        gender: values.gender,
      });
      message.success(t('login.registerSuccess'));
      setActiveTab('login');
      setNeedVerify(true);
      setVerifyId(values.email || values.username);
    } catch (error: unknown) {
      message.error(getErrorMessage(error, t('login.registerFail')));
    } finally {
      setLoading(false);
    }
  };

  const handleResend = async () => {
    const identifier = (verifyId || loginForm.getFieldValue('username') || '').trim();
    if (!identifier) {
      message.error(t('login.usernameRequired'));
      return;
    }
    setResendLoading(true);
    try {
      await resendVerification(identifier);
      message.success(t('login.resendOk'));
    } catch (error: unknown) {
      message.error(getErrorMessage(error, t('login.registerFail')));
    } finally {
      setResendLoading(false);
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
        <motion.div
          className={styles.brand}
          variants={staggerEnter}
          initial="hidden"
          animate="show"
        >
          <motion.div className={styles.identity} variants={fadeUp}>
            <div className={styles.wordmark}>
              StarByte<span>.</span>
            </div>
            <p className={styles.kicker}>{t('login.kicker')}</p>
          </motion.div>
          <motion.div className={styles.message} variants={fadeUp}>
            <h1>
              {t('login.headlineLine1')}
              <br />
              {t('login.headlineLine2')}
            </h1>
            <p className={styles.story}>{t('login.story')}</p>
          </motion.div>
          <motion.div className={styles.visual} variants={fadeUp}>
            <div className={styles.connections}>
              <span className={styles.mark} aria-hidden="true">
                <i />
                <i />
                <i />
                <i />
              </span>
              <span>{t('login.pillars')}</span>
            </div>
            <p className={styles.caption}>{t('login.caption')}</p>
          </motion.div>
        </motion.div>
      </div>
      <motion.div className={styles.right} variants={fadeUp} initial="hidden" animate="show">
        <div className={styles.mobileBrand}>StarByte.</div>
        <Card className={styles.card}>
          <header className={styles.cardHeader}>
            <div className={styles.cardHeaderTop}>
              <p className={styles.formKicker}>{t('login.formKicker')}</p>
              <LanguageSelect />
            </div>
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
                      window.location.assign(
                        getCasLoginURL(getRedirectPath(location.state, searchParams.get('next'))),
                      );
                    }}
                  >
                    {t('login.cas')}
                  </Button>
                  <p className={styles.casHint}>{t('login.casHint')}</p>
                </div>
              )}

              {casEnabled && (
                <Divider plain className={styles.casDivider}>
                  {t('login.orLocal')}
                </Divider>
              )}

              <Form
                form={loginForm}
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
                  <Input
                    autoComplete="username"
                    prefix={<UserOutlined />}
                    placeholder={t('login.usernameOrStudent')}
                  />
                </Form.Item>

                <Form.Item
                  name="password"
                  label={t('login.password')}
                  rules={[
                    { required: true, message: t('login.passwordRequired') },
                    { min: 6, message: t('login.passwordMin') },
                  ]}
                >
                  <Input.Password
                    autoComplete="current-password"
                    prefix={<LockOutlined />}
                    placeholder={t('login.password')}
                  />
                </Form.Item>

                <Form.Item className={styles.submitItem}>
                  <Button type="primary" htmlType="submit" loading={loading} block>
                    {t('login.submit')}
                  </Button>
                </Form.Item>

                {(needVerify || verifyId) && (
                  <div className={styles.switchTab}>
                    {t('login.resendHint')}
                    <Button type="link" loading={resendLoading} onClick={() => void handleResend()}>
                      {t('login.resend')}
                    </Button>
                  </div>
                )}

                <div className={styles.switchTab}>
                  {t('login.hint')}
                  <Button type="link" onClick={() => setActiveTab('register')}>
                    {t('login.goRegister')}
                  </Button>
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

              <Form.Item
                name="gender"
                label={t('login.gender')}
                rules={[{ required: true, message: t('login.genderRequired') }]}
              >
                <Select
                  placeholder={t('login.genderRequired')}
                  options={[
                    { value: 1, label: t('login.genderMale') },
                    { value: 2, label: t('login.genderFemale') },
                  ]}
                />
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
                  <Input.Password
                    autoComplete="new-password"
                    prefix={<LockOutlined />}
                    placeholder={t('login.password')}
                  />
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
                  <Input.Password
                    prefix={<LockOutlined />}
                    placeholder={t('login.confirmPassword')}
                  />
                </Form.Item>
              </div>

              <Form.Item className={styles.submitItem}>
                <Button type="primary" htmlType="submit" loading={loading} block>
                  {t('login.register')}
                </Button>
              </Form.Item>

              <div className={styles.switchTab}>
                {t('login.hasAccount')}
                <Button type="link" onClick={() => setActiveTab('login')}>
                  {t('login.goLogin')}
                </Button>
              </div>
            </Form>
          )}
        </Card>
        <FeatureProvider keys={PUBLIC_CMS_KEYS}>
          <LoginPublicFooter />
        </FeatureProvider>
      </motion.div>
    </div>
  );
};

export default Login;
