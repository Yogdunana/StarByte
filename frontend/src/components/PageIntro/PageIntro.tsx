import type { ReactNode } from 'react';
import styles from './PageIntro.module.css';

interface PageIntroProps { eyebrow?: string; title: string; description?: string; actions?: ReactNode }
export default function PageIntro({ eyebrow, title, description, actions }: PageIntroProps) {
  return <header className={styles.header}>
    <div>{eyebrow && <span className={styles.eyebrow}>{eyebrow}</span>}<h1>{title}</h1>{description && <p>{description}</p>}</div>
    {actions && <div className={styles.actions}>{actions}</div>}
  </header>;
}
