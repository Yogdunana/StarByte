import React, { useEffect, useState } from 'react';
import { Spin, Typography } from 'antd';
import { useTranslation } from 'react-i18next';
import { getPublicPage } from '@/api/knowledge';
import type { KnowledgeDoc } from '@/api/knowledge';
import MarkdownArticle from './MarkdownArticle';

const AboutUsPage: React.FC = () => {
  const { t } = useTranslation();
  const [doc, setDoc] = useState<KnowledgeDoc | null>(null);
  const [error, setError] = useState('');

  useEffect(() => {
    getPublicPage('about-us')
      .then(setDoc)
      .catch(() => setError(t('public.loadFailed')));
  }, [t]);

  if (error) return <Typography.Paragraph>{error}</Typography.Paragraph>;
  if (!doc) return <Spin />;
  return (
    <>
      <Typography.Paragraph type="secondary">{doc.summary}</Typography.Paragraph>
      <MarkdownArticle content={doc.content} />
    </>
  );
};

export default AboutUsPage;
