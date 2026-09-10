import { useCallback, useEffect, useRef, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Alert, Button, Card, Empty, Grid, Input, Modal, Pagination, Popconfirm, Select, Space, Table, message } from 'antd';
import { PlusOutlined, ReloadOutlined } from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import StatusTag from '@/components/StatusTag/StatusTag';
import PageIntro from '@/components/PageIntro/PageIntro';
import { usePermission } from '@/hooks/usePermission';
import { cancelMeeting, createMeeting, deleteMeeting, endMeeting, getMeetingList, startMeeting, updateMeeting } from '@/api/meeting';
import type { Meeting, MeetingStatus } from '@/types/api';
import { formatDateTime } from '@/utils/format';
import { MeetingStatusMap, MeetingTypeMap } from './meta';
import FormModal from './FormModal';
import styles from './MeetingList.module.css';
export default function ListPage() {
  const nav = useNavigate();
  const screens = Grid.useBreakpoint();
  const canCreate = usePermission('meeting:create');
  const [list, setList] = useState<Meeting[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [status, setStatus] = useState<MeetingStatus>();
  const [keyword, setKeyword] = useState('');
  const [loading, setLoading] = useState(true);
  const [failed, setFailed] = useState(false);
  const [busy, setBusy] = useState(false);
  const [open, setOpen] = useState(false);
  const [editing, setEditing] = useState<Meeting | null>(null);
  const [cancelling, setCancelling] = useState<Meeting | null>(null);
  const [reason, setReason] = useState('');
  const sequence = useRef(0);
  const load = useCallback(async () => {
    const seq = ++sequence.current;
    setLoading(true);
    try {
      const result = await getMeetingList({ page, page_size: 10, status, keyword });
      if (seq !== sequence.current) return;
      setList(result.list); setTotal(result.total); setFailed(false);
      if (!result.list.length && page > 1 && result.total <= (page - 1) * 10) setPage(Math.max(1, Math.ceil(result.total / 10)));
    } catch { if (seq === sequence.current) setFailed(true); }
    finally { if (seq === sequence.current) setLoading(false); }
  }, [page, status, keyword]);
  useEffect(() => { void load(); return () => { sequence.current += 1; }; }, [load]);
  const act = async (operation: () => Promise<unknown>, success: string) => {
    setBusy(true);
    try { await operation(); message.success(success); setCancelling(null); await load(); } catch { /* Keep prior data. */ } finally { setBusy(false); }
  };
  const actions = (item: Meeting) => <Space wrap size={2}>
    <Button type="link" size="small" onClick={() => nav(`/meeting/${item.id}`)}>进入会议</Button>
    {item.can_update && item.status === 0 && <Button type="text" size="small" disabled={busy} onClick={() => { setEditing(item); setOpen(true); }}>编辑</Button>}
    {item.can_manage && item.status === 0 && <Button type="text" size="small" disabled={busy} onClick={() => void act(() => startMeeting(item.id), '会议已开始')}>开始</Button>}
    {item.can_manage && item.status === 1 && <Popconfirm title="结束这场会议？" description="所有开放投票将同时截止，已有选票会保留。" okText="结束会议" cancelText="继续会议" onConfirm={() => act(() => endMeeting(item.id), '会议已结束')}><Button type="text" size="small" disabled={busy}>结束</Button></Popconfirm>}
    {item.can_manage && item.status < 2 && <Button type="text" size="small" disabled={busy} onClick={() => { setCancelling(item); setReason(''); }}>取消</Button>}
    {item.can_delete && (item.status === 0 || item.status === 3) && <Popconfirm title="删除这场会议？" description="有投票历史的会议不能删除。" okText="删除" cancelText="保留" onConfirm={() => act(() => deleteMeeting(item.id), '会议已删除')}><Button type="text" size="small" danger disabled={busy}>删除</Button></Popconfirm>}
  </Space>;
  const columns: ColumnsType<Meeting> = [
    { title: '会议', dataIndex: 'title', render: (_, row) => <div className={styles.titleCell}><button onClick={() => nav(`/meeting/${row.id}`)}>{row.title}</button><span>{MeetingTypeMap[row.meeting_type]} · {row.location || '地点待确认'}</span></div> },
    { title: '开始时间', dataIndex: 'start_time', width: 175, render: (value: string) => formatDateTime(value) },
    { title: '组织者', width: 120, render: (_, row) => row.organizer.name || '待确认' },
    { title: '签到', width: 80, render: (_, row) => `${row.checked_in_count}/${row.attendee_count}` },
    { title: '状态', width: 95, render: (_, row) => <StatusTag status={row.status} mapping={MeetingStatusMap} /> },
    { title: '操作', width: 245, render: (_, row) => actions(row) },
  ];
  return <>
    <PageIntro eyebrow="MEETINGS / 会议协作" title="把讨论变成共同行动" description="安排议程、邀请伙伴，用投票和纪要记录每一个决定。" actions={canCreate && <Button type="primary" icon={<PlusOutlined />} onClick={() => { setEditing(null); setOpen(true); }}>新建会议</Button>} />
    <Card>
      <div className={styles.filters}><Input.Search allowClear placeholder="搜索会议标题或地点" aria-label="搜索会议" onSearch={value => { setKeyword(value); setPage(1); }} /><Select allowClear placeholder="全部状态" aria-label="会议状态" value={status} onChange={value => { setStatus(value); setPage(1); }} options={Object.entries(MeetingStatusMap).map(([key, value]) => ({ value: Number(key), label: value.text }))} /><Button icon={<ReloadOutlined />} aria-label="刷新会议列表" loading={loading} onClick={() => void load()} /></div>
      {failed && <Alert type="warning" showIcon message="会议列表暂时无法更新" description="下方保留上次读取的内容，可点击刷新重试。" style={{ marginBottom: 16 }} />}
      {screens.lg ? <Table rowKey="id" loading={loading} columns={columns} dataSource={list} scroll={{ x: 960 }} pagination={{ current: page, total, pageSize: 10, showSizeChanger: false, onChange: setPage, showTotal: count => `共 ${count} 场会议` }} /> : <>
        <div className={styles.cards} aria-busy={loading}>{list.map(item => <article key={item.id}><div className={styles.cardTop}><span>{MeetingTypeMap[item.meeting_type]}</span><StatusTag status={item.status} mapping={MeetingStatusMap} /></div><h2><button onClick={() => nav(`/meeting/${item.id}`)}>{item.title}</button></h2><p>{formatDateTime(item.start_time)}</p><p>{item.location || '地点待确认'} · {item.organizer.name}</p><div className={styles.cardBottom}><span>{item.checked_in_count}/{item.attendee_count} 人签到</span>{actions(item)}</div></article>)}</div>
        {!list.length && !loading && !failed && <Empty description="暂无符合条件的会议" />}{!list.length && loading && <Card loading />}
        <Pagination current={page} pageSize={10} total={total} showSizeChanger={false} onChange={setPage} style={{ marginTop: 20 }} />
      </>}
    </Card>
    <FormModal open={open} editing={editing} onCancel={() => setOpen(false)} onSubmit={async values => {
      if (editing) { await updateMeeting(editing.id, values); message.success('会议已更新'); } else { await createMeeting(values); message.success('会议已创建'); }
      setOpen(false); await load();
    }} />
    <Modal title="取消会议" open={Boolean(cancelling)} confirmLoading={busy} okText="确认取消" cancelText="保留会议" onCancel={() => { if (!busy) setCancelling(null); }} onOk={() => { if (cancelling) void act(() => cancelMeeting(cancelling.id, reason), '会议已取消'); }}>
      <p>取消「{cancelling?.title}」后将停止接收投票，保留已有记录。</p><Input.TextArea aria-label="取消原因" placeholder="填写取消原因（可选）" maxLength={500} rows={3} value={reason} onChange={event => setReason(event.target.value)} />
    </Modal>
  </>;
}
