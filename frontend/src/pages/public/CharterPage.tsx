import React from 'react';
import styles from './CharterPage.module.css';

const CharterPage: React.FC = () => (
  <div className={styles.wrap}>
    <iframe
      className={styles.frame}
      title="深圳北理莫斯科大学计算机协会章程"
      src="/charter/smbu-ca-charter.html"
    />
  </div>
);

export default CharterPage;
