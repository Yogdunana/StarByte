import React, { useEffect, useState } from 'react';
import { Card, Empty, List, Typography } from 'antd';
import { useTranslation } from 'react-i18next';
import { Link } from 'react-router-dom';
import { listCmsPages, type CmsPage } from '@/api/feature';
import FeatureEnabled from '@/components/FeatureEnabled/FeatureEnabled';

const CmsPageView: React.FC = () => {
  const { t } = useTranslation();
  const [list, setList] = useState<CmsPage[]>([]);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    setLoading(true);
    listCmsPages().then((rows) => setList(rows || [])).finally(() => setLoading(false));
  }, []);

  return (
    <FeatureEnabled flag="cms.public">
      <Card title={t('cms.title')} loading={loading}>
        <Typography.Paragraph type="secondary">{t('cms.intro')}</Typography.Paragraph>
        {list.length === 0 ? <Empty description={t('cms.empty')} /> : (
          <List
            dataSource={list}
            renderItem={(item) => (
              <List.Item>
                <List.Item.Meta
                  title={<Link to={`/forms/${item.id}/fill`}>{item.name}</Link>}
                  description={item.description}
                />
              </List.Item>
            )}
          />
        )}
      </Card>
    </FeatureEnabled>
  );
};

export default CmsPageView;
