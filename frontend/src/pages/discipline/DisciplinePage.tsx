import React, { useCallback, useEffect, useState } from 'react';
import { Button, Card, Descriptions, Drawer, Form, Input, Modal, Select, Space, Table, message } from 'antd';
import { PlusOutlined } from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import { useTranslation } from 'react-i18next';
import { useSelector } from 'react-redux';
import { usePermission } from '@/hooks/usePermission';
import { selectCurrentUser } from '@/store/slices/userSlice';
import { getUserList, type UserListItem } from '@/api/user';
import {
  appealDiscipline,
  approveDiscipline,
  createDisciplineRecord,
  getDisciplineRecord,
  getDisciplineRecords,
  revokeDiscipline,
  type DisciplineRecord,
} from '@/api/discipline';

const DisciplinePage: React.FC = () => {
  const { t } = useTranslation();
  const me = useSelector(selectCurrentUser);
  const canCreate = usePermission('discipline:create');
  const canApprove = usePermission('discipline:approve');
  const canRevoke = usePermission('discipline:revoke');
  const [list, setList] = useState<DisciplineRecord[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(false);
  const [open, setOpen] = useState(false);
  const [detail, setDetail] = useState<DisciplineRecord | null>(null);
  const [users, setUsers] = useState<UserListItem[]>([]);
  const [form] = Form.useForm();

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const res = await getDisciplineRecords({ page, page_size: 10 });
      setList(res.list || []);
      setTotal(res.total);
    } finally {
      setLoading(false);
    }
  }, [page]);

  useEffect(() => { void load(); }, [load]);
  useEffect(() => {
    if (!open) return;
    void getUserList({ page: 1, page_size: 50 })
      .then((res) => setUsers(res.list || []))
      .catch(() => undefined);
  }, [open]);

  const levelLabel = (v: number) => t(`discipline.level.${v}`);
  const statusLabel = (v: number) => t(`discipline.status.${v}`);

  const columns: ColumnsType<DisciplineRecord> = [
    { title: t('discipline.title'), dataIndex: 'title', render: (v: string, r) => <Button type="link" onClick={() => { void getDisciplineRecord(r.id).then(setDetail); }}>{v}</Button> },
    { title: t('discipline.member'), render: (_, r) => r.user?.name || '-' },
    { title: t('discipline.levelLabel'), dataIndex: 'level', width: 120, render: levelLabel },
    { title: t('discipline.statusLabel'), dataIndex: 'status', width: 100, render: statusLabel },
    {
      title: t('common.actions'),
      width: 220,
      render: (_, record) => (
        <Space>
          {canApprove && (record.status === 0 || record.status === 3) && (
            <Button type="link" size="small" onClick={() => { void approveDiscipline(record.id).then(() => { message.success(t('common.saved')); void load(); }); }}>{t('discipline.approve')}</Button>
          )}
          {canRevoke && record.status !== 2 && (
            <Button type="link" size="small" danger onClick={() => {
              Modal.confirm({
                title: t('discipline.revoke'),
                content: <Input.TextArea id="revoke-reason" rows={3} />,
                onOk: () => {
                  const reason = (document.getElementById('revoke-reason') as HTMLTextAreaElement | null)?.value || t('discipline.revoke');
                  return revokeDiscipline(record.id, reason).then(() => { message.success(t('common.saved')); void load(); });
                },
              });
            }}
            >{t('discipline.revoke')}</Button>
          )}
          {me?.id === record.user?.id && record.status === 1 && (
            <Button type="link" size="small" onClick={() => {
              Modal.confirm({
                title: t('discipline.appeal'),
                content: <Input.TextArea id="appeal-reason" rows={3} />,
                onOk: () => {
                  const reason = (document.getElementById('appeal-reason') as HTMLTextAreaElement | null)?.value || '';
                  if (!reason) return Promise.reject();
                  return appealDiscipline(record.id, reason).then(() => { message.success(t('common.saved')); void load(); });
                },
              });
            }}
            >{t('discipline.appeal')}</Button>
          )}
        </Space>
      ),
    },
  ];

  return (
    <Card
      title={t('discipline.records')}
      extra={canCreate && (
        <Button type="primary" icon={<PlusOutlined />} onClick={() => { form.resetFields(); setOpen(true); }}>
          {t('discipline.create')}
        </Button>
      )}
    >
      <Table rowKey="id" loading={loading} columns={columns} dataSource={list} pagination={{ current: page, pageSize: 10, total, onChange: setPage }} />
      <Modal open={open} title={t('discipline.create')} onCancel={() => setOpen(false)} onOk={() => form.submit()} destroyOnClose>
        <Form form={form} layout="vertical" onFinish={(values: { user_id: string; title: string; level: number; description?: string }) => {
          void createDisciplineRecord(values).then(() => { message.success(t('common.saved')); setOpen(false); void load(); });
        }}
        >
          <Form.Item name="user_id" label={t('discipline.member')} rules={[{ required: true }]}>
            <Select
              showSearch
              optionFilterProp="label"
              placeholder={t('discipline.pickMember')}
              options={users.map((u) => ({
                value: u.id,
                label: `${u.real_name || u.username} (${u.username})`,
              }))}
            />
          </Form.Item>
          <Form.Item name="title" label={t('discipline.title')} rules={[{ required: true }]}><Input /></Form.Item>
          <Form.Item name="level" label={t('discipline.levelLabel')} rules={[{ required: true }]}>
            <Select options={[1, 2, 3, 4, 5].map((v) => ({ value: v, label: t(`discipline.level.${v}`) }))} />
          </Form.Item>
          <Form.Item name="description" label={t('discipline.description')}><Input.TextArea rows={3} /></Form.Item>
        </Form>
      </Modal>
      <Drawer open={!!detail} title={detail?.title} onClose={() => setDetail(null)} width={480}>
        {detail && (
          <Descriptions column={1} bordered size="small">
            <Descriptions.Item label={t('discipline.member')}>{detail.user?.name}</Descriptions.Item>
            <Descriptions.Item label={t('discipline.levelLabel')}>{levelLabel(detail.level)}</Descriptions.Item>
            <Descriptions.Item label={t('discipline.statusLabel')}>{statusLabel(detail.status)}</Descriptions.Item>
            <Descriptions.Item label={t('discipline.description')}>{detail.description || '-'}</Descriptions.Item>
            {detail.flow_instance_id && <Descriptions.Item label={t('discipline.flow')}>{detail.flow_instance_id}</Descriptions.Item>}
          </Descriptions>
        )}
      </Drawer>
    </Card>
  );
};

export default DisciplinePage;
