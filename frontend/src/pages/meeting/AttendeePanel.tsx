import { useEffect, useState } from 'react';
import { Alert, Avatar, Button, Empty, Popconfirm, Select, Space, Tag, message } from 'antd';
import { addAttendees, removeAttendee } from '@/api/meeting';
import { getUserList } from '@/api/user';
import type { MeetingAttendee } from '@/types/api';
import styles from './MeetingDetail.module.css';
interface Props { meetingId: string; organizerId: string; items: MeetingAttendee[]; canManage: boolean; onChange: (items: MeetingAttendee[]) => void }
const positions: Record<string, string> = { president: '会长', vice_president: '副会长', minister: '部长', vice_minister: '副部长', deputy: '副部长', officer: '干事', member: '会员', center_director: '中心主任' };
export default function AttendeePanel({ meetingId, organizerId, items, canManage, onChange }: Props) {
  const [options, setOptions] = useState<Array<{ value: string; label: string }>>([]);
  const [selected, setSelected] = useState<string[]>([]);
  const [failed, setFailed] = useState(false);
  const [busy, setBusy] = useState(false);
  const [query, setQuery] = useState('');
  useEffect(() => {
    if (!canManage) return;
    let active = true;
    const timer = window.setTimeout(() => {
      void getUserList({ page: 1, page_size: 50, keyword: query }).then(result => {
        if (active) { setOptions(result.list.map(user => ({ value: user.id, label: user.real_name || user.username }))); setFailed(false); }
      }).catch(() => { if (active) setFailed(true); });
    }, 300);
    return () => { active = false; window.clearTimeout(timer); };
  }, [canManage, query]);
  const add = async () => {
    setBusy(true);
    try { onChange(await addAttendees(meetingId, selected)); setSelected([]); message.success('参会人已添加'); } catch { /* Keep selections. */ } finally { setBusy(false); }
  };
  const remove = async (userId: string) => {
    setBusy(true);
    try { await removeAttendee(meetingId, userId); onChange(items.filter(item => item.user_id !== userId)); } catch { /* Backend retains check-in and voting history. */ } finally { setBusy(false); }
  };
  return <section>
    <div className={styles.panelHeader}><div><h2>参会伙伴</h2><p>{items.length} 人受邀 · {items.filter(item => item.attended).length} 人已签到</p></div></div>
    {canManage && <div className={styles.invite}>{failed && <Alert type="warning" message="人员列表暂不可用或无读取权限，请联系管理员。" showIcon />}<Select mode="multiple" showSearch filterOption={false} onSearch={setQuery} aria-label="选择参会人" placeholder="搜索姓名或账号，选择参会人" value={selected} onChange={setSelected} options={options.filter(option => !items.some(item => item.user_id === option.value))} style={{ width: '100%' }} /><Button type="primary" disabled={!selected.length} loading={busy} onClick={() => void add()}>确认添加</Button></div>}
    {items.length ? <div className={styles.people}>{items.map(item => <article key={item.id}><Avatar>{item.name.slice(0, 1)}</Avatar><div className={styles.person}><strong>{item.name}</strong><span>{item.user_id === organizerId ? '会议组织者' : positions[item.position_code || ''] || '参会人'}</span></div><Space><Tag color={item.attended ? 'green' : 'default'}>{item.attended ? '已签到' : '未签到'}</Tag>{canManage && !item.attended && item.user_id !== organizerId && <Popconfirm title="移除这位参会人？" description="有投票记录的参会人将保留。" okText="移除" cancelText="保留" onConfirm={() => remove(item.user_id)}><Button type="text" danger size="small" disabled={busy}>移除</Button></Popconfirm>}</Space></article>)}</div> : <Empty description="暂无参会人" />}
  </section>;
}
