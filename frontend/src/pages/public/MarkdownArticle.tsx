import React from 'react';
import { renderMarkdown } from '@/utils/markdown';
import styles from '@/layouts/PublicLayout/PublicLayout.module.css';

const MarkdownArticle: React.FC<{ content?: string }> = ({ content }) => (
  <article className={styles.article} dangerouslySetInnerHTML={{ __html: renderMarkdown(content || '') }} />
);

export default MarkdownArticle;
