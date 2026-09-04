import React, { useCallback, useEffect, useState } from 'react';
import { Button, Card, Input, Modal, Select, Space, Table, message } from 'antd';
import { exportMemberProfiles, getMemberList, updateMemberStatus } from '@/api/member';
import type { MemberProfile } from '@/types/api';
import { usePermission } from '@/hooks/usePermission';
import { downloadBlob } from '@/utils/download';
import { buildProfileColumns } from './profileColumns';
import ProfileDetailDrawer from './ProfileDetailDrawer';

const ProfilePage: React.FC = () => {
  const canExport = usePermission('member:export');
  const canManage = usePermission('member:manage');
  const [loading, setLoading] = useState(false);
  const [list, setList] = useState<MemberProfile[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [keyword, setKeyword] = useState('');
  const [detailId, setDetailId] = useState<string | null>(null);
  const [statusRecord, setStatusRecord] = useState<MemberProfile | null>(null);
  const [newStatus, setNewStatus] = useState<number>(1);
  const [reason, setReason] = useState('');

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const res = await getMemberList({
        page,
        page_size: pageSize,
        keyword: keyword || undefined,
      });
      setList(res.list);
      setTotal(res.total);
    } finally {
      setLoading(false);
    }
  }, [page, pageSize, keyword]);

  useEffect(() => {
    void load();
  }, [load]);

  const handleExport = async () => {
    const res = await exportMemberProfiles({
      page: 1,
      page_size: 200,
      keyword: keyword || undefined,
    });
    downloadBlob(res.data, 'member-profiles.pdf');
    message.success('已开始下载');
  };

  const handleStatusOk = async () => {
    if (!statusRecord || !reason.trim()) {
      message.warning('请填写原因');
      return;
    }
    await updateMemberStatus(statusRecord.id, newStatus, reason);
    message.success('状态已更新');
    setStatusRecord(null);
    setReason('');
    void load();
  };

  return (
    <Card
      title="人员档案"
      extra={
        canExport && (
          <Button onClick={() => void handleExport()}>导出 PDF</Button>
        )
      }
    >
      <Space style={{ marginBottom: 16 }}>
        <Input.Search
          placeholder="姓名 / 学号 / 技能 / 部门"
          allowClear
          onSearch={(v) => {
            setKeyword(v);
            setPage(1);
          }}
          style={{ width: 280 }}
        />
      </Space>
      <Table
        rowKey="id"
        loading={loading}
        dataSource={list}
        columns={buildProfileColumns({
          onView: (r) => setDetailId(r.id),
          onStatus: canManage ? (r) => setStatusRecord(r) : undefined,
        })}
        pagination={{
          current: page,
          pageSize,
          total,
          onChange: (p, ps) => {
            setPage(p);
            setPageSize(ps);
          },
        }}
      />
      <ProfileDetailDrawer
        open={!!detailId}
        profileId={detailId}
        onClose={() => setDetailId(null)}
        onSaved={() => {
          setDetailId(null);
          void load();
        }}
      />
      <Modal
        title="变更档案状态"
        open={!!statusRecord}
        onCancel={() => setStatusRecord(null)}
        onOk={() => void handleStatusOk()}
      >
        <Select
          style={{ width: '100%', marginBottom: 12 }}
          value={newStatus}
          onChange={setNewStatus}
          options={[
            { value: 0, label: '正常' },
            { value: 1, label: '禁用' },
            { value: 2, label: '已退出' },
          ]}
        />
        <Input.TextArea
          rows={3}
          placeholder="变更原因"
          value={reason}
          onChange={(e) => setReason(e.target.value)}
        />
      </Modal>
    </Card>
  );
};

export default ProfilePage;
