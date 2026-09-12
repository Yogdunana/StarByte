import { tx, useLocale } from '@/i18n/text';
import React, { useEffect, useState } from 'react';
import { useSearchParams } from 'react-router-dom';
import { useSelector } from 'react-redux';
import { Button, Card, Result, Space, Spin, message } from 'antd';
import { checkinMeeting, getAttendees, getMeetingDetail, getMeetingQRCode } from '@/api/meeting';
import { selectCurrentUser } from '@/store/slices/userSlice';
import type { Meeting, MeetingAttendee } from '@/types/api';

const CheckinPage: React.FC = () => {
  useLocale();
  const [params] = useSearchParams();
  const meetingId = params.get('meeting_id') || '';
  const token = params.get('token') || '';
  const currentUser = useSelector(selectCurrentUser);
  const [loading, setLoading] = useState(true);
  const [meeting, setMeeting] = useState<Meeting | null>(null);
  const [mine, setMine] = useState<MeetingAttendee | null>(null);
  const [qr, setQr] = useState<string>();
  const [done, setDone] = useState(false);

  useEffect(() => {
    if (!meetingId) {
      setLoading(false);
      return;
    }
    Promise.all([
      getMeetingDetail(meetingId),
      getAttendees(meetingId).catch(() => []),
      getMeetingQRCode(meetingId).catch(() => undefined),
    ])
      .then(([m, list, code]) => {
        setMeeting(m);
        const own = list.find((a) => a.user_id === currentUser?.id) || null;
        setMine(own);
        if (code) setQr(code.png_base64);
      })
      .finally(() => setLoading(false));
  }, [meetingId, currentUser]);

  const doCheckin = async () => {
    if (!meetingId) return;
    await checkinMeeting(meetingId, token || undefined);
    message.success(tx('签到成功'));
    setDone(true);
  };

  if (loading) return <Spin />;
  if (done) return <Result status="success" title={tx('签到成功')} subTitle={meeting?.title} />;

  return (
    <Card title={tx('会议签到')}>
      {!meeting ? (
        <Result status="info" title={tx('未指定会议')} />
      ) : !mine ? (
        <Result status="warning" title={tx('你不在参会人名单中')} subTitle={meeting.title} />
      ) : (
        <Space direction="vertical" size={16} style={{ width: '100%' }}>
          <div>{meeting.title}</div>
          <div>
            {tx('地点：')}
            {meeting.location || '-'}
          </div>
          <div>
            {tx('开始：')}
            {meeting.start_time?.replace('T', ' ').slice(0, 16)}
          </div>
          {qr && (
            <img
              alt={tx('签到二维码')}
              src={`data:image/png;base64,${qr}`}
              style={{ width: 220 }}
            />
          )}
          <Button type="primary" disabled={mine.attended} onClick={() => void doCheckin()}>
            {mine.attended ? tx('已签到') : tx('手动签到')}
          </Button>
        </Space>
      )}
    </Card>
  );
};

export default CheckinPage;
