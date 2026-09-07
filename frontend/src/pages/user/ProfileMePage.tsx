import React, { useEffect } from 'react';
import { Card, Descriptions, Tag } from 'antd';
import { useDispatch, useSelector } from 'react-redux';
import { fetchCurrentUser, selectCurrentUser, selectUserLoading } from '@/store/slices/userSlice';
import type { AppDispatch } from '@/store';

const genderLabel = (gender?: number): string => {
  if (gender === 1) return '男';
  if (gender === 2) return '女';
  return '未知';
};

const statusLabel = (status?: number): { text: string; color: string } => {
  if (status === 1) return { text: '禁用', color: 'red' };
  if (status === 2) return { text: '锁定', color: 'orange' };
  return { text: '正常', color: 'green' };
};

const dash = (value?: string): string => (value && value.trim() ? value : '未填写');

const ProfileMePage: React.FC = () => {
  const dispatch = useDispatch<AppDispatch>();
  const user = useSelector(selectCurrentUser);
  const loading = useSelector(selectUserLoading);

  useEffect(() => {
    void dispatch(fetchCurrentUser());
  }, [dispatch]);

  const status = statusLabel(user?.status);

  return (
    <Card title="个人中心" loading={loading && !user}>
      <Descriptions column={2} bordered size="small">
        <Descriptions.Item label="姓名">{dash(user?.real_name)}</Descriptions.Item>
        <Descriptions.Item label="学号">{dash(user?.student_no)}</Descriptions.Item>
        <Descriptions.Item label="用户名">{dash(user?.username)}</Descriptions.Item>
        <Descriptions.Item label="性别">{genderLabel(user?.gender)}</Descriptions.Item>
        <Descriptions.Item label="年级">{dash(user?.grade)}</Descriptions.Item>
        <Descriptions.Item label="专业">{dash(user?.major)}</Descriptions.Item>
        <Descriptions.Item label="部门">{dash(user?.department_name)}</Descriptions.Item>
        <Descriptions.Item label="职位">{dash(user?.position_name)}</Descriptions.Item>
        <Descriptions.Item label="邮箱">{dash(user?.email)}</Descriptions.Item>
        <Descriptions.Item label="手机">{dash(user?.phone)}</Descriptions.Item>
        <Descriptions.Item label="状态">
          <Tag color={status.color}>{status.text}</Tag>
        </Descriptions.Item>
        <Descriptions.Item label="角色">
          {(user?.roles || []).length > 0
            ? user?.roles.map((role) => <Tag key={role}>{role}</Tag>)
            : '未分配'}
        </Descriptions.Item>
      </Descriptions>
    </Card>
  );
};

export default ProfileMePage;
