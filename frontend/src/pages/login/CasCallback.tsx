import React, { useEffect, useRef, useState } from 'react';
import { Card, Spin, message } from 'antd';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { useDispatch } from 'react-redux';
import { useTranslation } from 'react-i18next';

import { exchangeCasCode } from '@/api/auth';
import { setToken } from '@/store/slices/authSlice';
import { fetchCurrentUser } from '@/store/slices/userSlice';
import type { AppDispatch } from '@/store';
import type { CASExchangeResponse } from '@/types/api';
import { draftFromExchange, saveCASRegisterDraft } from './casRegisterDraft';
import styles from './Login.module.css';

const exchangeByCode = new Map<string, Promise<CASExchangeResponse>>();

function exchangeCasCodeOnce(code: string): Promise<CASExchangeResponse> {
  let pending = exchangeByCode.get(code);
  if (!pending) {
    pending = exchangeCasCode(code);
    exchangeByCode.set(code, pending);
  }
  return pending;
}

function safeRedirect(path: string | undefined): string {
  if (!path || !path.startsWith('/') || path.startsWith('//')) {
    return '/dashboard';
  }
  return path;
}

const CasCallback: React.FC = () => {
  const { t } = useTranslation();
  const [params] = useSearchParams();
  const navigate = useNavigate();
  const dispatch = useDispatch<AppDispatch>();
  const [busy, setBusy] = useState(true);
  const applied = useRef(false);
  const code = params.get('code')?.trim() ?? '';

  useEffect(() => {
    if (!code) {
      message.error(t('login.casFail'));
      navigate('/login', { replace: true });
      setBusy(false);
      return;
    }
    let cancelled = false;
    (async () => {
      try {
        const result = await exchangeCasCodeOnce(code);
        if (cancelled || applied.current) return;
        applied.current = true;
        if (result.needs_registration) {
          const draft = draftFromExchange(result);
          if (!draft) {
            throw new Error(t('login.casFail'));
          }
          saveCASRegisterDraft(draft);
          navigate('/register/cas', { replace: true });
          return;
        }
        dispatch(
          setToken({
            accessToken: result.access_token,
            refreshToken: result.refresh_token,
          }),
        );
        await dispatch(fetchCurrentUser()).unwrap();
        message.success(t('login.success'));
        navigate(safeRedirect(result.redirect), { replace: true });
      } catch {
        if (!cancelled) {
          message.error(t('login.casFail'));
          navigate('/login', { replace: true });
        }
      } finally {
        if (!cancelled) setBusy(false);
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [code, dispatch, navigate, t]);

  return (
    <div className={styles.container}>
      <div className={styles.right} style={{ gridColumn: '1 / -1' }}>
        <Card className={styles.card}>
          <div style={{ textAlign: 'center', padding: 24 }}>
            {busy ? <Spin size="large" /> : null}
            <p className={styles.formHint} style={{ marginTop: 16 }}>
              {t('login.casProcessing')}
            </p>
          </div>
        </Card>
      </div>
    </div>
  );
};

export default CasCallback;
