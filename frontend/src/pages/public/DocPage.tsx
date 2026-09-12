import React, { useEffect, useState } from 'react';
import { Spin, Typography } from 'antd';
import { useNavigate, useParams } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import axios from 'axios';
import { getPublicDoc } from '@/api/knowledge';
import type { KnowledgeDoc } from '@/api/knowledge';
import { loginPath } from '@/utils/nextPath';
import MarkdownArticle from './MarkdownArticle';

const DocPage: React.FC = () => {
  const { slug = '' } = useParams();
  const { t } = useTranslation();
  const navigate = useNavigate();
  const [doc, setDoc] = useState<KnowledgeDoc | null>(null);
  const [error, setError] = useState('');

  useEffect(() => {
    if (!slug) return;
    getPublicDoc(slug)
      .then(setDoc)
      .catch((err: unknown) => {
        const status = axios.isAxiosError(err) ? err.response?.status : 0;
        const code = axios.isAxiosError(err) ? (err.response?.data as { code?: number } | undefined)?.code : 0;
        if (status === 401 || code === 33004) {
          navigate(loginPath(`/docs/${slug}`), { replace: true });
          return;
        }
        setError(t('public.loadFailed'));
      });
  }, [slug, navigate, t]);

  if (error) return <Typography.Paragraph>{error}</Typography.Paragraph>;
  if (!doc) return <Spin />;
  return (
    <>
      <Typography.Paragraph type="secondary">{doc.summary}</Typography.Paragraph>
      <MarkdownArticle content={doc.content} />
    </>
  );
};

export default DocPage;
