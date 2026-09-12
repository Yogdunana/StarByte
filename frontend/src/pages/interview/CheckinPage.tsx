import { tx, useLocale } from '@/i18n/text';
import React, { useEffect, useMemo, useState } from 'react';
import { useSearchParams } from 'react-router-dom';
import { useSelector } from 'react-redux';
import { Button, Card, Result, Space, Spin, message } from 'antd';
import { checkinInterview, getMyInterviews, getSessionQRCode } from '@/api/interview';
import { selectCurrentUser } from '@/store/slices/userSlice';
import type { Interview } from '@/types/api';

const CheckinPage: React.FC = () => {
  const uiLanguage = useLocale();
  const [params] = useSearchParams();
  const sessionId = params.get('session_id') || '';
  const token = params.get('token') || '';
  const currentUser = useSelector(selectCurrentUser);
  const [loading, setLoading] = useState(true);
  const [mine, setMine] = useState<Interview | null>(null);
  const [qr, setQr] = useState<string>();
  const [done, setDone] = useState(false);

  useEffect(() => {
    Promise.all([
      getMyInterviews(),
      sessionId ? getSessionQRCode(sessionId).catch(() => undefined) : Promise.resolve(undefined),
    ])
      .then(([list, code]) => {
        const own = list.filter((i) => !currentUser || i.applicant.id === currentUser.id);
        const hit = own.find((i) => !sessionId || i.session_id === sessionId);
        setMine(hit || null);
        if (code) setQr(code.png_base64);
      })
      .finally(() => setLoading(false));
  }, [sessionId, currentUser]);

  const title = useMemo(() => {
    void uiLanguage;
    return mine ? `${mine.applicant.name} · ${mine.session_title || tx('面试')}` : tx('面试签到');
  }, [mine, uiLanguage]);

  const doCheckin = async () => {
    if (!mine) return;
    await checkinInterview(mine.id, token || undefined);
    message.success(tx('签到成功'));
    setDone(true);
  };

  if (loading) return <Spin />;

  if (done) {
    return <Result status="success" title={tx('签到成功')} subTitle={title} />;
  }

  return (
    <Card title={tx('面试签到')}>
      {!mine ? (
        <Result status="info" title={tx('没有待签到的面试')} />
      ) : (
        <Space direction="vertical" size={16} style={{ width: '100%' }}>
          <div>{title}</div>
          <div>
            {tx('地点：')}
            {mine.location || '-'}
          </div>
          <div>
            {tx('预约：')}
            {mine.scheduled_time || tx('待定')}
          </div>
          {qr && (
            <img
              alt={tx('签到二维码')}
              src={`data:image/png;base64,${qr}`}
              style={{ width: 220 }}
            />
          )}
          <Button type="primary" disabled={mine.status !== 0} onClick={() => void doCheckin()}>
            {mine.status === 0 ? tx('手动签到') : tx('已签到或不可签到')}
          </Button>
        </Space>
      )}
    </Card>
  );
};

export default CheckinPage;
