import React from 'react';
import { Result, Button } from 'antd';
import { HomeOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { getToken } from '@/utils/storage';

const NotFound: React.FC = () => {
  const navigate = useNavigate();
  const { t } = useTranslation();

  return (
    <Result
      status="404"
      title="404"
      subTitle={t('error.notFound')}
      extra={
        <Button type="primary" icon={<HomeOutlined />} onClick={() => navigate(getToken() ? '/dashboard' : '/', { replace: true })}>
          {t('error.backHome')}
        </Button>
      }
      style={{ height: '100%', display: 'flex', flexDirection: 'column', justifyContent: 'center' }}
    />
  );
};

export default NotFound;
