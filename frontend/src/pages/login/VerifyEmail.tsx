import React, { useEffect, useState } from 'react';
import { Button, Card, Result, Spin } from 'antd';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { useTranslation } from 'react-i18next';

import { verifyEmail } from '@/api/auth';
import styles from './Login.module.css';

const VerifyEmail: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const [params] = useSearchParams();
  const token = params.get('token')?.trim() ?? '';
  const [status, setStatus] = useState<'loading' | 'ok' | 'fail'>(token ? 'loading' : 'fail');

  useEffect(() => {
    if (!token) return;
    let cancelled = false;
    void verifyEmail(token)
      .then(() => {
        if (!cancelled) setStatus('ok');
      })
      .catch(() => {
        if (!cancelled) setStatus('fail');
      });
    return () => {
      cancelled = true;
    };
  }, [token]);

  return (
    <div className={styles.container}>
      <div className={styles.right} style={{ gridColumn: '1 / -1' }}>
        <Card className={styles.card}>
          {status === 'loading' ? (
            <div style={{ textAlign: 'center', padding: 24 }}>
              <Spin size="large" />
              <p className={styles.formHint} style={{ marginTop: 16 }}>
                {t('login.verifyChecking')}
              </p>
            </div>
          ) : (
            <Result
              status={status === 'ok' ? 'success' : 'error'}
              title={status === 'ok' ? t('login.verifyOk') : t('login.verifyFail')}
              extra={
                <Button type="primary" onClick={() => navigate('/login', { replace: true })}>
                  {t('login.verifyGoLogin')}
                </Button>
              }
            />
          )}
        </Card>
      </div>
    </div>
  );
};

export default VerifyEmail;
