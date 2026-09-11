import React, { useEffect, useState } from 'react';
import { useSearchParams } from 'react-router-dom';
import { Button, Card, Input, Result, Space, Spin, message } from 'antd';
import { checkinActivity, getActivityDetail, getMyRegistration } from '@/api/activity';
import type { Activity, Registration } from '@/api/activity';
import { formatDateTime } from '@/utils/format';

const CheckinPage: React.FC = () => {
  const [params] = useSearchParams();
  const activityId = params.get('activity_id') || '';
  const [token, setToken] = useState(params.get('token') || '');
  const [loading, setLoading] = useState(true);
  const [activity, setActivity] = useState<Activity | null>(null);
  const [mine, setMine] = useState<Registration | null>(null);
  const [done, setDone] = useState(false);

  useEffect(() => {
    if (!activityId) {
      setLoading(false);
      return;
    }
    Promise.all([
      getActivityDetail(activityId),
      getMyRegistration(activityId),
    ])
      .then(([a, reg]) => {
        setActivity(a);
        setMine(reg);
      })
      .finally(() => setLoading(false));
  }, [activityId]);

  const doCheckin = async () => {
    if (!activityId) return;
    if (!token.trim()) {
      message.error('请填写签到令牌');
      return;
    }
    await checkinActivity(activityId, { method: 1, token: token.trim() });
    message.success('签到成功');
    setDone(true);
  };

  if (loading) return <Spin />;
  if (done) return <Result status="success" title="签到成功" subTitle={activity?.title} />;

  return (
    <Card title="活动签到">
      {!activity ? (
        <Result status="info" title="未指定活动" />
      ) : !mine || mine.status !== 1 ? (
        <Result status="warning" title="你尚未通过该活动报名" subTitle={activity.title} />
      ) : mine.checkin_status === 1 ? (
        <Result status="success" title="你已签到" subTitle={activity.title} />
      ) : (
        <Space direction="vertical" size={16} style={{ width: '100%' }}>
          <div>{activity.title}</div>
          <div>地点：{activity.location || '-'}</div>
          <div>开始：{formatDateTime(activity.start_time, 'YYYY-MM-DD HH:mm')}</div>
          <Input
            placeholder="扫码后自动填入，或粘贴签到令牌"
            value={token}
            onChange={(e) => setToken(e.target.value)}
          />
          <Button type="primary" onClick={() => void doCheckin()}>确认签到</Button>
        </Space>
      )}
    </Card>
  );
};

export default CheckinPage;
