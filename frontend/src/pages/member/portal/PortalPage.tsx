import React, { useEffect, useState } from 'react';
import { Card, Empty, List, Tag, Typography } from 'antd';
import { useTranslation } from 'react-i18next';
import { Link } from 'react-router-dom';
import { getMembershipPortal, type MembershipPortal } from '@/api/feature';
import FeatureEnabled from '@/components/FeatureEnabled/FeatureEnabled';

const statusColor: Record<number, string> = {
  0: 'default', 1: 'processing', 2: 'processing', 3: 'success', 4: 'error', 5: 'warning',
};

const PortalPage: React.FC = () => {
  const { t } = useTranslation();
  const [data, setData] = useState<MembershipPortal | null>(null);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    setLoading(true);
    getMembershipPortal().then(setData).finally(() => setLoading(false));
  }, []);

  return (
    <FeatureEnabled flag="membership.portal">
      <Card title={t('portal.title')} loading={loading}>
        <Typography.Paragraph type="secondary">{t('portal.intro')}</Typography.Paragraph>
        <Typography.Paragraph>
          <Link to="/member/application">{t('portal.goApply')}</Link>
        </Typography.Paragraph>
        {!data?.applications?.length ? <Empty description={t('portal.empty')} /> : (
          <List
            dataSource={data.applications}
            renderItem={(item) => (
              <List.Item extra={<Tag color={statusColor[item.status] || 'default'}>{t(`portal.status.${item.status}`, { defaultValue: String(item.status) })}</Tag>}>
                <List.Item.Meta title={item.real_name} description={item.submitted_at} />
              </List.Item>
            )}
          />
        )}
      </Card>
    </FeatureEnabled>
  );
};

export default PortalPage;
