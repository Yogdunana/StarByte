import React, { useEffect, useState } from 'react';
import { Button, Card, Descriptions, Form, Input, Modal, Select, Tag, message } from 'antd';
import { useDispatch, useSelector } from 'react-redux';
import { useTranslation } from 'react-i18next';
import { fetchCurrentUser, selectCurrentUser, selectUserLoading } from '@/store/slices/userSlice';
import { updateMyProfile, type ProfileUpdate } from '@/api/user';
import type { AppDispatch } from '@/store';

const ProfileMePage: React.FC = () => {
  const { t } = useTranslation();
  const dispatch = useDispatch<AppDispatch>();
  const user = useSelector(selectCurrentUser);
  const loading = useSelector(selectUserLoading);
  const [open, setOpen] = useState(false);
  const [saving, setSaving] = useState(false);
  const [form] = Form.useForm<ProfileUpdate>();
  useEffect(() => {
    void dispatch(fetchCurrentUser());
  }, [dispatch]);
  const dash = (value?: string) => value?.trim() || t('profile.empty', '未填写');
  const edit = () => {
    form.resetFields();
    form.setFieldsValue({
      real_name: user?.real_name,
      email: user?.email,
      phone: user?.phone,
      gender: user?.gender,
    });
    setOpen(true);
  };
  const save = async () => {
    const values = await form.validateFields();
    setSaving(true);
    try {
      await updateMyProfile(values);
      await dispatch(fetchCurrentUser()).unwrap();
      setOpen(false);
      message.success(t('common.saveSuccess', '保存成功'));
    } finally {
      setSaving(false);
    }
  };
  return (
    <Card
      title={t('profile.title', '个人中心')}
      loading={loading && !user}
      extra={
        <Button disabled={!user} onClick={edit}>
          {t('profile.edit', '编辑资料')}
        </Button>
      }
    >
      <Descriptions column={2} bordered size="small">
        <Descriptions.Item label={t('profile.name', '姓名')}>
          {dash(user?.real_name)}
        </Descriptions.Item>
        <Descriptions.Item label={t('profile.studentNo', '学号')}>
          {dash(user?.student_no)}
        </Descriptions.Item>
        <Descriptions.Item label={t('profile.username', '用户名')}>
          {dash(user?.username)}
        </Descriptions.Item>
        <Descriptions.Item label={t('profile.gender', '性别')}>
          {t(`profile.gender${user?.gender || 0}`, ['未知', '男', '女'][user?.gender || 0])}
        </Descriptions.Item>
        <Descriptions.Item label={t('profile.grade', '年级')}>
          {dash(user?.grade)}
        </Descriptions.Item>
        <Descriptions.Item label={t('profile.major', '专业')}>
          {dash(user?.major)}
        </Descriptions.Item>
        <Descriptions.Item label={t('profile.department', '部门')}>
          {dash(user?.department_name)}
        </Descriptions.Item>
        <Descriptions.Item label={t('profile.position', '职位')}>
          {dash(user?.position_name)}
        </Descriptions.Item>
        <Descriptions.Item label={t('profile.email', '邮箱')}>
          {dash(user?.email)}
        </Descriptions.Item>
        <Descriptions.Item label={t('profile.phone', '手机')}>
          {dash(user?.phone)}
        </Descriptions.Item>
        <Descriptions.Item label={t('common.status', '状态')}>
          <Tag>
            {t(`profile.status${user?.status || 0}`, ['正常', '禁用', '锁定'][user?.status || 0])}
          </Tag>
        </Descriptions.Item>
        <Descriptions.Item label={t('profile.roles', '角色')}>
          {user?.roles?.length
            ? user.roles.map((role) => <Tag key={role}>{role}</Tag>)
            : t('profile.noRoles', '未分配')}
        </Descriptions.Item>
      </Descriptions>
      <Modal
        title={t('profile.edit', '编辑资料')}
        open={open}
        onCancel={() => setOpen(false)}
        onOk={() => void save().catch(() => undefined)}
        confirmLoading={saving}
      >
        <Form form={form} layout="vertical">
          <Form.Item
            name="real_name"
            label={t('profile.name', '姓名')}
            rules={[{ required: true, whitespace: true, max: 50 }]}
          >
            <Input />
          </Form.Item>
          <Form.Item
            name="email"
            label={t('profile.email', '邮箱')}
            rules={[{ type: 'email', max: 100 }]}
          >
            <Input />
          </Form.Item>
          <Form.Item name="phone" label={t('profile.phone', '手机')} rules={[{ max: 20 }]}>
            <Input />
          </Form.Item>
          <Form.Item name="gender" label={t('profile.gender', '性别')}>
            <Select
              options={[0, 1, 2].map((value) => ({
                value,
                label: t(`profile.gender${value}`, ['未知', '男', '女'][value]),
              }))}
            />
          </Form.Item>
        </Form>
      </Modal>
    </Card>
  );
};
export default ProfileMePage;
