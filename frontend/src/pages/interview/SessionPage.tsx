import React, { useCallback, useEffect, useState } from 'react';
import {
  Button, Card, Image, Input, Modal, Select, Space, Table, Tabs, message,
} from 'antd';
import { PlusOutlined } from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import StatusTag from '@/components/StatusTag/StatusTag';
import { usePermission } from '@/hooks/usePermission';
import { getMemberDepartments, getApplicationList } from '@/api/member';
import { getUserList } from '@/api/user';
import {
  assignEvaluators, createInterview, createSession, deleteSession, endSession,
  getInterviewList, getSessionList, getSessionQRCode, startSession, updateSession,
} from '@/api/interview';
import type {
  Interview, InterviewSession, InterviewSessionStatus, MemberApplication, MemberDepartmentOption,
} from '@/types/api';
import { InterviewStatusMap, ResultMap, SessionStatusMap } from './meta';
import SessionFormModal from './SessionFormModal';

const SessionPage: React.FC = () => {
  const canManage = usePermission('interview:manage');
  const [sessions, setSessions] = useState<InterviewSession[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [status, setStatus] = useState<InterviewSessionStatus | undefined>();
  const [loading, setLoading] = useState(false);
  const [formOpen, setFormOpen] = useState(false);
  const [editing, setEditing] = useState<InterviewSession | null>(null);
  const [departments, setDepartments] = useState<MemberDepartmentOption[]>([]);
  const [current, setCurrent] = useState<InterviewSession | null>(null);
  const [records, setRecords] = useState<Interview[]>([]);
  const [qr, setQr] = useState<string>();

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const res = await getSessionList({ page, page_size: 10, status });
      setSessions(res.list);
      setTotal(res.total);
    } finally {
      setLoading(false);
    }
  }, [page, status]);

  useEffect(() => {
    void load();
    void getMemberDepartments().then(setDepartments).catch(() => undefined);
  }, [load]);

  const loadRecords = async (session: InterviewSession) => {
    setCurrent(session);
    const res = await getInterviewList({ session_id: session.id, page: 1, page_size: 50 });
    setRecords(res.list);
  };

  const sessionColumns: ColumnsType<InterviewSession> = [
    { title: '名称', dataIndex: 'title' },
    { title: '轮次', dataIndex: 'round', width: 70 },
    { title: '部门', dataIndex: 'department_name', width: 100, render: (v?: string) => v || '-' },
    { title: '地点', dataIndex: 'location' },
    { title: '开始', dataIndex: 'start_time', width: 170 },
    { title: '人数', key: 'n', width: 90, render: (_, r) => `${r.candidate_count}/${r.max_candidates}` },
    { title: '状态', dataIndex: 'status', width: 90, render: (v: number) => <StatusTag status={v} mapping={SessionStatusMap} /> },
    {
      title: '操作',
      width: 280,
      render: (_, record) => (
        <Space wrap>
          <Button type="link" size="small" onClick={() => void loadRecords(record)}>面试者</Button>
          {canManage && record.status === 0 && (
            <Button type="link" size="small" onClick={() => { setEditing(record); setFormOpen(true); }}>编辑</Button>
          )}
          {canManage && record.status === 0 && (
            <Button type="link" size="small" onClick={() => startSession(record.id).then(load)}>开始</Button>
          )}
          {canManage && record.status === 1 && (
            <Button type="link" size="small" onClick={() => endSession(record.id).then(load)}>结束</Button>
          )}
          {canManage && (
            <Button type="link" size="small" onClick={() => getSessionQRCode(record.id).then((q) => setQr(q.png_base64))}>二维码</Button>
          )}
          {canManage && (record.status === 0 || record.status === 3) && (
            <Button type="link" size="small" danger onClick={() => {
              Modal.confirm({ title: '删除该场次？', onOk: () => deleteSession(record.id).then(load) });
            }}>删除</Button>
          )}
        </Space>
      ),
    },
  ];

  const importApplicant = async () => {
    if (!current) return;
    const apps = await getApplicationList({ page: 1, page_size: 50, status: 2 });
    let selected = '';
    Modal.confirm({
      title: '从入会申请导入',
      content: (
        <Select
          style={{ width: '100%', marginTop: 12 }}
          placeholder="选择面试中的申请"
          options={apps.list.map((a: MemberApplication) => ({
            label: `${a.real_name}（${a.student_no}）`,
            value: a.id,
          }))}
          onChange={(v) => { selected = v; }}
        />
      ),
      onOk: async () => {
        if (!selected) return;
        await createInterview({ session_id: current.id, application_id: selected });
        message.success('已导入');
        await loadRecords(current);
        await load();
      },
    });
  };

  const assign = async (record: Interview) => {
    const users = await getUserList({ page: 1, page_size: 50 });
    let ids: string[] = record.evaluators.map((e) => e.id);
    Modal.confirm({
      title: '分配面试官',
      content: (
        <Select
          mode="multiple"
          style={{ width: '100%', marginTop: 12 }}
          defaultValue={ids}
          options={users.list.map((u) => ({ label: u.real_name || u.username, value: u.id }))}
          onChange={(v) => { ids = v; }}
        />
      ),
      onOk: async () => {
        await assignEvaluators(record.id, ids);
        message.success('已分配并通知');
        if (current) await loadRecords(current);
      },
    });
  };

  return (
    <Card title="面试安排">
      <Tabs
        items={[
          {
            key: 'sessions',
            label: '场次管理',
            children: (
              <>
                <Space style={{ marginBottom: 16 }}>
                  <Select
                    allowClear
                    placeholder="状态"
                    style={{ width: 140 }}
                    value={status}
                    onChange={setStatus}
                    options={Object.entries(SessionStatusMap).map(([k, v]) => ({ value: Number(k), label: v.text }))}
                  />
                  {canManage && (
                    <Button type="primary" icon={<PlusOutlined />} onClick={() => { setEditing(null); setFormOpen(true); }}>
                      新建场次
                    </Button>
                  )}
                </Space>
                <Table
                  rowKey="id"
                  loading={loading}
                  columns={sessionColumns}
                  dataSource={sessions}
                  pagination={{ current: page, total, onChange: setPage }}
                />
              </>
            ),
          },
          {
            key: 'records',
            label: current ? `面试者 · ${current.title}` : '面试者',
            children: (
              <>
                <Space style={{ marginBottom: 16 }}>
                  <Input disabled value={current ? current.title : '请先在场次中选择'} style={{ width: 240 }} />
                  {canManage && current && <Button onClick={() => void importApplicant()}>导入申请</Button>}
                </Space>
                <Table
                  rowKey="id"
                  dataSource={records}
                  columns={[
                    { title: '面试者', render: (_, r) => r.applicant.name },
                    { title: '学号', dataIndex: 'student_no' },
                    { title: '面试官', render: (_, r) => r.evaluators.map((e) => e.name).join('、') || '-' },
                    { title: '状态', dataIndex: 'status', render: (v: number) => <StatusTag status={v} mapping={InterviewStatusMap} /> },
                    { title: '结果', dataIndex: 'result', render: (v: number) => <StatusTag status={v} mapping={ResultMap} /> },
                    ...(canManage ? [{
                      title: '操作',
                      render: (_: unknown, r: Interview) => (
                        <Button type="link" size="small" onClick={() => void assign(r)}>分配面试官</Button>
                      ),
                    }] : []),
                  ]}
                />
              </>
            ),
          },
        ]}
      />
      <SessionFormModal
        open={formOpen}
        editing={editing}
        departments={departments}
        onCancel={() => setFormOpen(false)}
        onSubmit={async (values) => {
          if (editing) await updateSession(editing.id, values);
          else await createSession(values as never);
          setFormOpen(false);
          await load();
        }}
      />
      <Modal title="签到二维码" open={!!qr} footer={null} onCancel={() => setQr(undefined)}>
        {qr && <Image src={`data:image/png;base64,${qr}`} alt="签到二维码" />}
      </Modal>
    </Card>
  );
};

export default SessionPage;
