import React from 'react';
import { Result, Button } from 'antd';
import { useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';

const Forbidden: React.FC = () => {
  const navigate = useNavigate();
  const { t } = useTranslation();

  return (
    <Result
      status="403"
      title="403"
      subTitle={t('error.forbidden')}
      extra={
        <Button type="primary" onClick={() => navigate('/dashboard', { replace: true })}>
          {t('error.backDashboard')}
        </Button>
      }
      style={{ height: '100%', display: 'flex', flexDirection: 'column', justifyContent: 'center' }}
    />
  );
};

export default Forbidden;
