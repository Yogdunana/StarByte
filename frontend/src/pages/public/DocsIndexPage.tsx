import React, { useEffect, useState } from 'react';
import { Input, List, Spin, Typography } from 'antd';
import { Link } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { listPublicDocs, searchPublicDocs } from '@/api/knowledge';
import type { KnowledgeDoc } from '@/api/knowledge';

const DocsIndexPage: React.FC = () => {
  const { t } = useTranslation();
  const [docs, setDocs] = useState<KnowledgeDoc[]>([]);
  const [loading, setLoading] = useState(true);

  const load = async (q?: string) => {
    setLoading(true);
    try {
      const res = q ? await searchPublicDocs(q) : await listPublicDocs();
      setDocs(res.list || []);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { void load(); }, []);

  return (
    <>
      <Typography.Title level={2}>{t('public.docsTitle')}</Typography.Title>
      <Input.Search
        allowClear
        placeholder={t('public.search')}
        onSearch={(v) => void load(v.trim() || undefined)}
        style={{ margin: '16px 0 24px', maxWidth: 420 }}
      />
      {loading ? <Spin /> : (
        <List
          dataSource={docs}
          locale={{ emptyText: t('public.emptyDocs') }}
          renderItem={(item) => (
            <List.Item>
              <List.Item.Meta
                title={<Link to={item.path || `/docs/${item.slug}`}>{item.title}</Link>}
                description={item.summary}
              />
            </List.Item>
          )}
        />
      )}
    </>
  );
};

export default DocsIndexPage;
