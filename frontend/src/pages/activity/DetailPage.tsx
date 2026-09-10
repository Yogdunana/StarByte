import React, { useCallback, useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import {
  Button, Card, Descriptions, Input, Rate, Space, Statistic, Table, Tabs, Tag, message,
} from 'antd';
import { ArrowLeftOutlined } from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import type { TabsProps } from 'antd';
import { usePermission } from '@/hooks/usePermission';
import {
  getActivityDetail, registerActivity, cancelRegistration, getRegistrations,
  approveRegistration, checkinActivity, getActivityStats, submitSurvey,
} from '@/api/activity';
import type { Activity, ActivityStats, Registration, RegistrationStatus } from '@/api/activity';
import { ActivityStatusMap, RegistrationStatusMap } from './meta';

const DetailPage: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const nav = useNavigate();
  const canManage = usePermission('activity:manage');
  const [activity, setActivity] = useState<Activity | null>(null);
  const [regs, setRegs] = useState<Registration[]>([]);
  const [stats, setStats] = useState<ActivityStats | null>(null);
  const [loading, setLoading] = useState(false);
  const [rating, setRating] = useState(0);
  const [comment, setComment] = useState('');

  const load = useCallback(async () => {
    if (!id) return;
    setLoading(true);
    try {
      const [a, r, s] = await Promise.all([
        getActivityDetail(id),
        canManage ? getRegistrations(id) : Promise.resolve([]),
        canManage ? getActivityStats(id) : Promise.resolve(null),
      ]);
      setActivity(a);
      setRegs(r);
      setStats(s);
    } finally {
      setLoading(false);
    }
  }, [id, canManage]);

  useEffect(() => { void load(); }, [load]);

  const onRegister = async () => {
    if (!id) return;
    await registerActivity(id);
    message.success('报名成功');
    load();
  };

  const onCancelReg = async () => {
    if (!id) return;
    await cancelRegistration(id);
    message.success('已取消报名');
    load();
  };

  const onApprove = async (r: Registration, approve: boolean) => {
    if (!id) return;
    await approveRegistration(id, r.user.id, approve);
    message.success(approve ? '已通过' : '已拒绝');
    load();
  };

  const onCheckin = async () => {
    if (!id) return;
    await checkinActivity(id, { method: 1 });
    message.success('签到成功');
    load();
  };

  const onSurvey = async () => {
    if (!id || rating === 0) return;
    await submitSurvey(id, { rating, comment });
    message.success('评价提交成功');
    load();
  };

  if (!activity) return null;
  const statusMeta = ActivityStatusMap[activity.status];
  const myReg = regs.find((r) => r.user.id === 'me') || null; // 后端可返回当前用户报名状态

  const regColumns: ColumnsType<Registration> = [
    { title: '成员', dataIndex: ['user', 'name'], key: 'name' },
    {
      title: '报名状态', dataIndex: 'status', key: 'status', width: 100,
      render: (s: RegistrationStatus) => {
        const m = RegistrationStatusMap[s];
        return <Tag color={m.color}>{m.text}</Tag>;
      },
    },
    {
      title: '签到状态', dataIndex: 'checkin_status', key: 'checkin', width: 100,
      render: (s: number) => s === 1 ? <Tag color="success">已签到</Tag> : <Tag>未签到</Tag>,
    },
    {
      title: '操作', key: 'action', width: 160,
      render: (_, r) => (
        <Space size={4}>
          {r.status === 0 && (
            <>
              <Button size="small" type="link" onClick={() => onApprove(r, true)}>通过</Button>
              <Button size="small" type="link" danger onClick={() => onApprove(r, false)}>拒绝</Button>
            </>
          )}
          {r.status === 3 && (
            <Button size="small" type="link" onClick={() => onApprove(r, true)}>递补</Button>
          )}
        </Space>
      ),
    },
  ];

  const canRegister = activity.status === 1;
  const canCheckin = activity.status === 1 || activity.status === 2;
  const canSurvey = activity.status === 3;

  return (
    <div>
      <Button icon={<ArrowLeftOutlined />} style={{ marginBottom: 16 }} onClick={() => nav(-1)}>返回</Button>
      <Card loading={loading} title={activity.title} extra={<Tag color={statusMeta.color}>{statusMeta.text}</Tag>}>
        <Descriptions column={2} bordered size="small">
          <Descriptions.Item label="分类">{activity.category}</Descriptions.Item>
          <Descriptions.Item label="地点">{activity.location}</Descriptions.Item>
          <Descriptions.Item label="开始时间">{activity.start_time}</Descriptions.Item>
          <Descriptions.Item label="结束时间">{activity.end_time}</Descriptions.Item>
          <Descriptions.Item label="人数上限">{activity.max_participants || '不限'}</Descriptions.Item>
          <Descriptions.Item label="组织者">{activity.organizer.name}</Descriptions.Item>
          <Descriptions.Item label="标签" span={2}>
            {activity.tags.map((t) => <Tag key={t}>{t}</Tag>)}
          </Descriptions.Item>
          <Descriptions.Item label="描述" span={2}>{activity.description}</Descriptions.Item>
        </Descriptions>

        <Space style={{ marginTop: 16 }}>
          {canRegister && (
            myReg && myReg.status !== 4
              ? <Button danger onClick={onCancelReg}>取消报名</Button>
              : <Button type="primary" onClick={onRegister}>立即报名</Button>
          )}
          {canCheckin && <Button onClick={onCheckin}>签到</Button>}
        </Space>
      </Card>

      {canManage && (
        <Card style={{ marginTop: 16 }} title="统计">
          <Space size={32} wrap>
            <Statistic title="报名人数" value={stats?.approved_count || 0} />
            <Statistic title="候补人数" value={stats?.waitlist_count || 0} />
            <Statistic title="签到人数" value={stats?.checked_in_count || 0} />
            <Statistic title="报名率" value={stats?.register_rate || 0} suffix="%" precision={1} />
            <Statistic title="出席率" value={stats?.attend_rate || 0} suffix="%" precision={1} />
            <Statistic title="平均评分" value={stats?.avg_rating || 0} precision={2} />
          </Space>
        </Card>
      )}

      <Tabs
        style={{ marginTop: 16 }}
        items={buildTabs(canManage, canSurvey, regColumns, regs, rating, setRating, comment, setComment, onSurvey)}
      />
    </div>
  );
};

export default DetailPage;

interface BuildTabsProps {
  canManage: boolean;
  canSurvey: boolean;
  regColumns: ColumnsType<Registration>;
  regs: Registration[];
  rating: number;
  setRating: (v: number) => void;
  comment: string;
  setComment: (v: string) => void;
  onSurvey: () => void;
}

function buildTabs(props: BuildTabsProps): TabsProps['items'] {
  const items: NonNullable<TabsProps['items']> = [];
  if (props.canManage) {
    items.push({
      key: 'regs', label: '报名管理',
      children: (
        <Table
          rowKey="id" columns={props.regColumns} dataSource={props.regs} size="small"
          pagination={{ pageSize: 10 }}
        />
      ),
    });
  }
  if (props.canSurvey) {
    items.push({
      key: 'survey', label: '满意度评价',
      children: (
        <Card size="small">
          <Space direction="vertical">
            <Rate value={props.rating} onChange={props.setRating} />
            <Input.TextArea
              rows={3} maxLength={1000} showCount
              value={props.comment} onChange={(e) => props.setComment(e.target.value)}
              placeholder="说说你的感受（可选）"
            />
            <Button type="primary" onClick={props.onSurvey}>提交评价</Button>
          </Space>
        </Card>
      ),
    });
  }
  return items;
}
