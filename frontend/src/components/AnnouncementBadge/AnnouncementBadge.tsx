import React from 'react';
import { Badge, Button, Tooltip } from 'antd';
import { NotificationOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { useAnnouncementUnread } from '@/hooks/useAnnouncementUnread';

const AnnouncementBadge: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { count, canRead } = useAnnouncementUnread();

  if (!canRead) return null;

  return (
    <Tooltip title={t('announcement.unreadHint', { count })}>
      <Badge count={count} size="small" offset={[-2, 2]}>
        <Button
          type="text"
          aria-label={t('announcement.menu')}
          icon={<NotificationOutlined />}
          onClick={() => navigate('/announcement/list')}
        />
      </Badge>
    </Tooltip>
  );
};

export default AnnouncementBadge;
