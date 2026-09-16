import React from 'react';
import { useTranslation } from 'react-i18next';
import styles from './CharterPage.module.css';

const CharterPage: React.FC = () => {
  const { t } = useTranslation();
  return (
    <div className={styles.wrap}>
      <iframe className={styles.frame} title={t('public.charterTitle')} src="/charter/smbu-ca-charter.html" />
    </div>
  );
};

export default CharterPage;
