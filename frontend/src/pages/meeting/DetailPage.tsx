import { useCallback, useEffect, useRef, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { Alert, Button, Card, Input, Modal, Skeleton, Space, Tabs, message } from 'antd';
import { ArrowLeftOutlined, QrcodeOutlined, ReloadOutlined } from '@ant-design/icons';
import StatusTag from '@/components/StatusTag/StatusTag';
import PageIntro from '@/components/PageIntro/PageIntro';
import { getAttendees, getMeetingAgendas, getMeetingDetail, getMeetingQRCode, getMeetingVotes, updateMinutes } from '@/api/meeting';
import type { Meeting, MeetingAgenda, MeetingAttendee, MeetingVote } from '@/types/api';
import { MeetingStatusMap, MeetingTypeMap } from './meta';
import { formatDateTime } from '@/utils/format';
import AgendaPanel from './AgendaPanel';
import AttendeePanel from './AttendeePanel';
import VotePanel from './VotePanel';
import styles from './MeetingDetail.module.css';

export default function DetailPage() {
  const { id = '' } = useParams();
  const nav = useNavigate();
  const [meeting, setMeeting] = useState<Meeting | null>(null);
  const [agendas, setAgendas] = useState<MeetingAgenda[]>([]);
  const [attendees, setAttendees] = useState<MeetingAttendee[]>([]);
  const [votes, setVotes] = useState<MeetingVote[]>([]);
  const [minutes, setMinutes] = useState('');
  const [savedMinutes, setSavedMinutes] = useState('');
  const [busy, setBusy] = useState(false);
  const [loading, setLoading] = useState(true);
  const [failed, setFailed] = useState(false);
  const [qr, setQr] = useState<string>();
  const [qrLoading, setQrLoading] = useState(false);
  const [activeTab, setActiveTab] = useState('agenda');
  const requestId = useRef(0);
  const load = useCallback(async () => {
    const seq = ++requestId.current;
    setLoading(true); setFailed(false);
    try {
      const [m, a, t, v] = await Promise.all([getMeetingDetail(id), getMeetingAgendas(id), getAttendees(id), getMeetingVotes(id)]);
      if (seq !== requestId.current) return;
      setMeeting(m); setAgendas(a); setAttendees(t); setVotes(v); setMinutes(m.minutes || ''); setSavedMinutes(m.minutes || '');
    } catch { if (seq === requestId.current) setFailed(true); }
    finally { if (seq === requestId.current) setLoading(false); }
  }, [id]);
  useEffect(() => { setMeeting(null); setQr(undefined); setActiveTab('agenda'); void load(); return () => { requestId.current += 1; }; }, [load]);
  const refreshVotes = useCallback(async () => { setVotes(await getMeetingVotes(id)); }, [id]);
  const save = async () => {
    setBusy(true);
    const draft = minutes;
    try { await updateMinutes(id, draft); setSavedMinutes(draft); message.success('会议纪要已保存'); } catch { /* Keep unsaved text. */ } finally { setBusy(false); }
  };
  const showQR = async () => {
    setQrLoading(true);
    try { setQr((await getMeetingQRCode(id)).png_base64); } catch { /* API displays error. */ } finally { setQrLoading(false); }
  };
  if (loading) return <Card><Skeleton active paragraph={{ rows: 6 }} /></Card>;
  if (failed || !meeting) return <Alert showIcon type="error" message="会议详情暂不可用" description="可能是网络问题、会议已删除，或你没有访问权限。" action={<Space><Button onClick={() => nav('/meeting/list')}>返回列表</Button><Button icon={<ReloadOutlined />} onClick={() => void load()}>重试</Button></Space>} />;
  const canManage = meeting.can_manage === true;
  const open = meeting.status === 0 || meeting.status === 1;
  return <>
    <PageIntro eyebrow="MEETINGS / 会议协作" title={meeting.title} description={meeting.description || '围绕共同议题，记录讨论、参与投票，形成可追溯的会议纪要。'} actions={<Button icon={<ArrowLeftOutlined />} onClick={() => nav('/meeting/list')}>返回会议</Button>} />
    <Card>
      <Space wrap><StatusTag status={meeting.status} mapping={MeetingStatusMap} /><span>{MeetingTypeMap[meeting.meeting_type]}</span>{canManage && open && <Button icon={<QrcodeOutlined />} loading={qrLoading} onClick={() => void showQR()}>签到二维码</Button>}</Space>
      <div className={styles.summary}><div><span>会议时间</span><strong>{formatDateTime(meeting.start_time)}</strong><p>至 {formatDateTime(meeting.end_time)}</p></div><div><span>会议地点</span><strong>{meeting.location || '地点待确认'}</strong>{meeting.online_link && /^https?:\/\//i.test(meeting.online_link) && <p><a href={meeting.online_link} target="_blank" rel="noreferrer">进入线上会议 ↗</a></p>}</div><div><span>组织者</span><strong>{meeting.organizer.name || '待确认'}</strong><p>{attendees.length} 人受邀 · {attendees.filter(item => item.attended).length} 人已签到</p></div></div>
      {meeting.cancel_reason && <Alert type="info" showIcon message={`取消原因：${meeting.cancel_reason}`} />}
      <Tabs activeKey={activeTab} onChange={setActiveTab} items={[
        { key: 'agenda', label: `议程 (${agendas.length})`, children: <AgendaPanel meetingId={id} items={agendas} canManage={canManage && open} onChange={setAgendas} /> },
        { key: 'attendee', label: `参会人 (${attendees.length})`, children: <AttendeePanel meetingId={id} organizerId={meeting.organizer.id} items={attendees} canManage={canManage && open} onChange={setAttendees} /> },
        { key: 'vote', label: `投票 (${votes.length})`, children: activeTab === 'vote' ? <VotePanel meetingId={id} votes={votes} canManage={canManage && open} onRefresh={refreshVotes} /> : null },
        { key: 'minutes', label: '会议纪要', children: <section><div className={styles.panelHeader}><div><h2>让讨论有记录</h2><p>记录结论、待办事项与负责人，便于会后跟进。</p></div></div>{canManage ? <><Input.TextArea aria-label="会议纪要" rows={12} value={minutes} onChange={event => setMinutes(event.target.value)} placeholder="讨论结论、后续行动、负责人及完成时间…" /><div className={styles.minutesHeader}><span>{minutes !== savedMinutes ? '有尚未保存的修改' : '内容已保存'}</span><Button type="primary" loading={busy} disabled={minutes === savedMinutes} onClick={() => void save()}>保存纪要</Button></div></> : <div className={styles.text}>{minutes || '会议纪要尚未发布。'}</div>}</section> },
      ]} />
    </Card>
    <Modal open={Boolean(qr)} onCancel={() => setQr(undefined)} footer={null} title="会议签到"><div className={styles.qr}><img src={`data:image/png;base64,${qr || ''}`} alt="会议签到二维码" /><p>已受邀的参会人登录后扫码签到。</p></div></Modal>
  </>;
}
