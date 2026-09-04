import React, { useEffect, useState } from 'react';
import { Descriptions, Drawer, Form, Input, Select, Table, Tabs, message, Button } from 'antd';
import { getMemberDetail, getMemberHistory, updateMember } from '@/api/member';
import type { MemberProfile, MemberProfileHistory } from '@/types/api';
import StatusTag from '@/components/StatusTag/StatusTag';
import { GenderMap } from '@/types/common';
import { MemberTypeMap, ProfileStatusMap } from '../meta';
import { usePermission } from '@/hooks/usePermission';

interface Props {
  open: boolean;
  profileId: string | null;
  onClose: () => void;
  onSaved: () => void;
}

const ProfileDetailDrawer: React.FC<Props> = ({ open, profileId, onClose, onSaved }) => {
  const canUpdate = usePermission('member:update');
  const [profile, setProfile] = useState<MemberProfile | null>(null);
  const [history, setHistory] = useState<MemberProfileHistory[]>([]);
  const [form] = Form.useForm();
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    if (!open || !profileId) return;
    getMemberDetail(profileId).then((p) => {
      setProfile(p);
      form.setFieldsValue({
        real_name: p.real_name,
        gender: p.gender,
        grade: p.grade,
        major: p.major,
        bio: p.bio,
        contact_phone: p.contact_phone,
        contact_email: p.contact_email,
        skills: p.skills,
      });
    });
    getMemberHistory(profileId).then(setHistory).catch(() => setHistory([]));
  }, [open, profileId, form]);

  const handleSave = async () => {
    if (!profileId) return;
    const values = await form.validateFields();
    setSaving(true);
    try {
      await updateMember(profileId, values);
      message.success('已保存');
      onSaved();
    } finally {
      setSaving(false);
    }
  };

  return (
    <Drawer title={profile?.real_name || '人员档案'} width={640} open={open} onClose={onClose}>
      {profile && (
        <Tabs
          items={[
            {
              key: 'basic',
              label: '基本信息',
              children: (
                <Descriptions column={1} bordered size="small">
                  <Descriptions.Item label="学号">{profile.student_no}</Descriptions.Item>
                  <Descriptions.Item label="性别">
                    <StatusTag status={profile.gender} mapping={GenderMap} />
                  </Descriptions.Item>
                  <Descriptions.Item label="年级">{profile.grade || '-'}</Descriptions.Item>
                  <Descriptions.Item label="专业">{profile.major || '-'}</Descriptions.Item>
                  <Descriptions.Item label="部门">{profile.department?.name || '-'}</Descriptions.Item>
                  <Descriptions.Item label="职位">{profile.position?.name || '-'}</Descriptions.Item>
                  <Descriptions.Item label="类型">
                    <StatusTag status={profile.member_type} mapping={MemberTypeMap} />
                  </Descriptions.Item>
                  <Descriptions.Item label="状态">
                    <StatusTag status={profile.status} mapping={ProfileStatusMap} />
                  </Descriptions.Item>
                  <Descriptions.Item label="电话">{profile.contact_phone || '-'}</Descriptions.Item>
                  <Descriptions.Item label="邮箱">{profile.contact_email || '-'}</Descriptions.Item>
                </Descriptions>
              ),
            },
            {
              key: 'skills',
              label: '技能 / 简介',
              children: (
                <Form form={form} layout="vertical" disabled={!canUpdate}>
                  <Form.Item name="real_name" label="姓名">
                    <Input />
                  </Form.Item>
                  <Form.Item name="gender" label="性别">
                    <Select
                      options={[
                        { value: 0, label: '未知' },
                        { value: 1, label: '男' },
                        { value: 2, label: '女' },
                      ]}
                    />
                  </Form.Item>
                  <Form.Item name="grade" label="年级">
                    <Input />
                  </Form.Item>
                  <Form.Item name="major" label="专业">
                    <Input />
                  </Form.Item>
                  <Form.Item name="skills" label="技能">
                    <Select mode="tags" />
                  </Form.Item>
                  <Form.Item name="bio" label="简介">
                    <Input.TextArea rows={4} />
                  </Form.Item>
                  <Form.Item name="contact_phone" label="电话">
                    <Input />
                  </Form.Item>
                  <Form.Item name="contact_email" label="邮箱">
                    <Input />
                  </Form.Item>
                  {canUpdate && (
                    <Button type="primary" loading={saving} onClick={() => void handleSave()}>
                      保存
                    </Button>
                  )}
                </Form>
              ),
            },
            {
              key: 'projects',
              label: '项目',
              children: (
                <Table
                  rowKey={(r) => `${r.name}-${r.period}`}
                  dataSource={profile.projects || []}
                  pagination={false}
                  columns={[
                    { title: '项目', dataIndex: 'name' },
                    { title: '角色', dataIndex: 'role' },
                    { title: '周期', dataIndex: 'period' },
                  ]}
                />
              ),
            },
            {
              key: 'history',
              label: '历史',
              children: (
                <Table
                  rowKey="id"
                  dataSource={history}
                  pagination={false}
                  columns={[
                    { title: '字段', dataIndex: 'field_name', width: 120 },
                    { title: '旧值', dataIndex: 'old_value' },
                    { title: '新值', dataIndex: 'new_value' },
                    { title: '原因', dataIndex: 'reason', width: 120 },
                    { title: '时间', dataIndex: 'created_at', width: 180 },
                  ]}
                />
              ),
            },
          ]}
        />
      )}
    </Drawer>
  );
};

export default ProfileDetailDrawer;
