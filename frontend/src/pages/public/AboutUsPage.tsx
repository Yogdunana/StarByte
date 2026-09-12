import React, { useEffect, useState } from 'react';
import { Spin, Typography } from 'antd';
import { useParams } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { getPublicPage } from '@/api/knowledge';
import type { KnowledgeDoc } from '@/api/knowledge';
import MarkdownArticle from './MarkdownArticle';

const AboutUsPage: React.FC = () => {
  const { slug = 'about-us' } = useParams();
  const { t } = useTranslation();
  const [doc, setDoc] = useState<KnowledgeDoc | null>(null);
  const [error, setError] = useState('');

  useEffect(() => {
    getPublicPage(slug)
      .then(setDoc)
      .catch(() => setError(t('public.loadFailed')));
  }, [slug, t]);

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
