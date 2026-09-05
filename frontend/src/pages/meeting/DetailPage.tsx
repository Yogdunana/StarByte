import React, { useCallback, useEffect, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { Button, Card, Input, Select, Space, Table, Tabs, message } from 'antd';
import { ArrowDownOutlined, ArrowUpOutlined } from '@ant-design/icons';
import StatusTag from '@/components/StatusTag/StatusTag';
import { usePermission } from '@/hooks/usePermission';
import { getUserList } from '@/api/user';
import {
  addAgenda, addAttendees, deleteAgenda, getAttendees, getMeetingAgendas, getMeetingDetail,
  getMeetingQRCode, getMeetingVotes, removeAttendee, sortAgendas, updateAgenda, updateMinutes,
} from '@/api/meeting';
import type { Meeting, MeetingAgenda, MeetingAttendee, MeetingVote } from '@/types/api';
import { MeetingStatusMap, MeetingTypeMap } from './meta';
import VotePanel from './VotePanel';

const DetailPage: React.FC = () => {
  const { id = '' } = useParams();
  const nav = useNavigate();
  const canManage = usePermission('meeting:manage');
  const [meeting, setMeeting] = useState<Meeting | null>(null);
  const [agendas, setAgendas] = useState<MeetingAgenda[]>([]);
  const [attendees, setAttendees] = useState<MeetingAttendee[]>([]);
  const [votes, setVotes] = useState<MeetingVote[]>([]);
  const [qr, setQr] = useState<string>();
  const [minutes, setMinutes] = useState('');
  const [userOpts, setUserOpts] = useState<Array<{ label: string; value: string }>>([]);

  const load = useCallback(async () => {
    if (!id) return;
    const [m, a, t, v] = await Promise.all([
      getMeetingDetail(id), getMeetingAgendas(id), getAttendees(id), getMeetingVotes(id),
    ]);
    setMeeting(m);
    setAgendas(a);
    setAttendees(t);
    setVotes(v);
    setMinutes(m.minutes || '');
  }, [id]);

  useEffect(() => { void load(); }, [load]);
  useEffect(() => {
    getUserList({ page: 1, page_size: 50 }).then((res) => {
      setUserOpts(res.list.map((u) => ({ value: u.id, label: u.real_name || u.username })));
    }).catch(() => undefined);
  }, []);

  const move = async (index: number, dir: -1 | 1) => {
    const next = [...agendas];
    const j = index + dir;
    if (j < 0 || j >= next.length) return;
    [next[index], next[j]] = [next[j], next[index]];
    setAgendas(await sortAgendas(id, next.map((x) => x.id)));
  };

  if (!meeting) return <Card loading />;

  return (
    <Card
      title={meeting.title}
      extra={<Button onClick={() => nav('/meeting/list')}>返回列表</Button>}
    >
      <Space wrap style={{ marginBottom: 16 }}>
        <StatusTag status={meeting.status} mapping={MeetingStatusMap} />
        <span>{MeetingTypeMap[meeting.meeting_type]}</span>
        <span>地点：{meeting.location || '-'}</span>
        <span>组织者：{meeting.organizer?.name}</span>
        <span>时间：{meeting.start_time?.replace('T', ' ').slice(0, 16)}</span>
        {canManage && (
          <Button onClick={() => getMeetingQRCode(id).then((c) => setQr(c.png_base64))}>签到二维码</Button>
        )}
      </Space>
      {qr && <img alt="签到二维码" src={`data:image/png;base64,${qr}`} style={{ width: 180, marginBottom: 16 }} />}
      <Tabs
        items={[
          {
            key: 'agenda',
            label: '议程',
            children: (
              <Space direction="vertical" style={{ width: '100%' }}>
                {canManage && (
                  <Button onClick={() => {
                    const title = window.prompt('议程标题');
                    if (!title) return;
                    addAgenda(id, { title }).then(load);
                  }}
                  >
                    添加议程
                  </Button>
                )}
                <Table
                  rowKey="id"
                  dataSource={agendas}
                  pagination={false}
                  onRow={(_, index) => ({
                    draggable: canManage,
                    onDragStart: (e) => { e.dataTransfer.setData('text/plain', String(index)); },
                    onDragOver: (e) => e.preventDefault(),
                    onDrop: (e) => {
                      e.preventDefault();
                      const from = Number(e.dataTransfer.getData('text/plain'));
                      const to = index ?? 0;
                      if (Number.isNaN(from) || from === to) return;
                      const next = [...agendas];
                      const [item] = next.splice(from, 1);
                      next.splice(to, 0, item);
                      void sortAgendas(id, next.map((x) => x.id)).then(setAgendas);
                    },
                  })}
                  columns={[
                    { title: '顺序', dataIndex: 'sort_order', width: 70 },
                    { title: '标题', dataIndex: 'title' },
                    { title: '汇报人', dataIndex: 'presenter' },
                    { title: '时长', dataIndex: 'duration', width: 80 },
                    {
                      title: '操作',
                      width: 220,
                      render: (_, __, index) => canManage && (
                        <Space>
                          <Button size="small" icon={<ArrowUpOutlined />} onClick={() => void move(index, -1)} />
                          <Button size="small" icon={<ArrowDownOutlined />} onClick={() => void move(index, 1)} />
                          <Button size="small" onClick={() => {
                            const title = window.prompt('议程标题', agendas[index].title);
                            if (!title) return;
                            const presenter = window.prompt('汇报人', agendas[index].presenter || '') || '';
                            void updateAgenda(id, agendas[index].id, { title, presenter }).then(load);
                          }}
                          >
                            编辑
                          </Button>
                          <Button size="small" danger onClick={() => deleteAgenda(id, agendas[index].id).then(load)}>删</Button>
                        </Space>
                      ),
                    },
                  ]}
                />
              </Space>
            ),
          },
          {
            key: 'attendee',
            label: '参会人',
            children: (
              <Space direction="vertical" style={{ width: '100%' }}>
                {canManage && (
                  <Select
                    mode="multiple"
                    placeholder="添加参会人"
                    options={userOpts}
                    style={{ width: 360 }}
                    onChange={(ids: string[]) => { if (ids.length) addAttendees(id, ids).then(load); }}
                  />
                )}
                <Table
                  rowKey="id"
                  dataSource={attendees}
                  pagination={false}
                  columns={[
                    { title: '姓名', dataIndex: 'name' },
                    { title: '职务', dataIndex: 'position_code' },
                    { title: '签到', dataIndex: 'attended', render: (v: boolean) => (v ? '已签到' : '未签到') },
                    {
                      title: '操作',
                      render: (_, r) => canManage && (
                        <Button size="small" danger onClick={() => removeAttendee(id, r.user_id).then(load)}>移除</Button>
                      ),
                    },
                  ]}
                />
              </Space>
            ),
          },
          {
            key: 'vote',
            label: '投票',
            children: <VotePanel meetingId={id} votes={votes} onRefresh={load} />,
          },
          {
            key: 'minutes',
            label: '纪要',
            children: (
              <Space direction="vertical" style={{ width: '100%' }}>
                <Input.TextArea rows={8} value={minutes} onChange={(e) => setMinutes(e.target.value)} disabled={!canManage} />
                {canManage && <Button type="primary" onClick={() => updateMinutes(id, minutes).then(() => message.success('已保存'))}>保存纪要</Button>}
              </Space>
            ),
          },
        ]}
      />
    </Card>
  );
};

export default DetailPage;
