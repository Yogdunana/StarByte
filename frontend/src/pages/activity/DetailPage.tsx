import React, { useCallback, useEffect, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import {
  Button, Card, Descriptions, Form, Input, Modal, Rate, Space, Statistic, Table, Tabs, Tag, message,
} from 'antd';
import type { ColumnsType } from 'antd/es/table';
import StatusTag from '@/components/StatusTag/StatusTag';
import { usePermission } from '@/hooks/usePermission';
import {
  approveRegistration, cancelRegistration, checkinActivity, getActivityDetail, getActivityStats,
  getRegistrations, registerActivity, submitSurvey,
} from '@/api/activity';
import type { Activity, ActivityStats, Registration } from '@/api/activity';
import { ActivityStatusMap, RegistrationStatusMap } from './meta';

const DetailPage: React.FC = () => {
  const { id = '' } = useParams();
  const nav = useNavigate();
  const canManage = usePermission('activity:manage');
  const [activity, setActivity] = useState<Activity | null>(null);
  const [regs, setRegs] = useState<Registration[]>([]);
  const [stats, setStats] = useState<ActivityStats | null>(null);
  const [surveyOpen, setSurveyOpen] = useState(false);

  const load = useCallback(async () => {
    if (!id) return;
    const [a, r, s] = await Promise.all([
      getActivityDetail(id), getRegistrations(id), getActivityStats(id),
    ]);
    setActivity(a);
    setRegs(r);
    setStats(s);
  }, [id]);

  useEffect(() => { void load(); }, [load]);

  if (!activity) return <Card loading />;

  const myReg = regs.find(() => false);
  void myReg;

  const regColumns: ColumnsType<Registration> = [
    { title: '姓名', key: 'name', render: (_, r) => r.user?.name || '-' },
    { title: '报名状态', dataIndex: 'status', width: 100, render: (v: number) => <StatusTag status={v} mapping={RegistrationStatusMap} /> },
    {
      title: '签到',
      dataIndex: 'checkin_status',
      width: 100,
      render: (v: number) => (v === 1 ? '已签到' : '未签到'),
    },
    { title: '签到时间', dataIndex: 'checked_in_at', width: 160, render: (v?: string) => v?.replace('T', ' ').slice(0, 16) || '-' },
    {
      title: '操作',
      width: 180,
      render: (_, r) => canManage && (
        <Space>
          {r.status !== 1 && (
            <Button size="small" type="link" onClick={() => approveRegistration(id, r.user.id, true).then(load)}>通过</Button>
          )}
          {r.status !== 2 && (
            <Button size="small" type="link" danger onClick={() => approveRegistration(id, r.user.id, false).then(load)}>拒绝</Button>
          )}
        </Space>
      ),
    },
  ];

  return (
    <Card
      title={activity.title}
      extra={<Button onClick={() => nav('/activity/list')}>返回列表</Button>}
    >
      <Space wrap style={{ marginBottom: 16 }}>
        <StatusTag status={activity.status} mapping={ActivityStatusMap} />
        {activity.category && <Tag>{activity.category}</Tag>}
        <span>地点：{activity.location || '-'}</span>
        <span>组织者：{activity.organizer?.name || '-'}</span>
        <span>开始：{activity.start_time?.replace('T', ' ').slice(0, 16)}</span>
        <span>结束：{activity.end_time?.replace('T', ' ').slice(0, 16)}</span>
        <span>
          报名：{activity.registered_count}/{activity.max_participants === 0 ? '不限' : activity.max_participants}
        </span>
      </Space>

      <Space style={{ marginBottom: 16 }}>
        {activity.status === 1 && (
          <Button type="primary" onClick={() => registerActivity(id).then(() => { message.success('报名成功'); return load(); })}>
            我要报名
          </Button>
        )}
        {[1, 2].includes(activity.status) && (
          <Button onClick={() => cancelRegistration(id).then(() => { message.success('已取消报名'); return load(); })}>
            取消报名
          </Button>
        )}
        {activity.status === 2 && (
          <Button onClick={() => checkinActivity(id, { method: 1 }).then(() => { message.success('签到成功'); return load(); })}>
            二维码签到
          </Button>
        )}
        {activity.status === 3 && (
          <Button type="primary" ghost onClick={() => setSurveyOpen(true)}>满意度评价</Button>
        )}
      </Space>

      {activity.tags?.length > 0 && (
        <Space style={{ marginBottom: 16 }}>
          {activity.tags.map((t) => <Tag key={t} color="blue">{t}</Tag>)}
        </Space>
      )}

      <Tabs
        items={[
          {
            key: 'detail',
            label: '活动说明',
            children: (
              <Descriptions column={1} bordered size="small">
                <Descriptions.Item label="说明">{activity.description || '-'}</Descriptions.Item>
              </Descriptions>
            ),
          },
          {
            key: 'regs',
            label: `报名管理（${regs.length}）`,
            children: (
              <Table
                rowKey="id"
                dataSource={regs}
                columns={regColumns}
                pagination={false}
              />
            ),
          },
          {
            key: 'stats',
            label: '统计',
            children: stats && (
              <Space wrap size="large">
                <Statistic title="已通过" value={stats.approved_count} />
                <Statistic title="候补" value={stats.waitlist_count} />
                <Statistic title="已签到" value={stats.checked_in_count} />
                <Statistic title="报名率" value={stats.register_rate} suffix="%" precision={1} />
                <Statistic title="出席率" value={stats.attend_rate} suffix="%" precision={1} />
                <Statistic title="评价数" value={stats.survey_count} />
                <Statistic title="平均评分" value={stats.avg_rating} precision={1} />
              </Space>
            ),
          },
        ]}
      />

      <SurveyModal
        open={surveyOpen}
        onCancel={() => setSurveyOpen(false)}
        onSubmit={async (rating, comment) => {
          await submitSurvey(id, { rating, comment });
          message.success('评价已提交');
          setSurveyOpen(false);
          await load();
        }}
      />
    </Card>
  );
};

const SurveyModal: React.FC<{
  open: boolean;
  onCancel: () => void;
  onSubmit: (rating: number, comment?: string) => Promise<void>;
}> = ({ open, onCancel, onSubmit }) => {
  const [form] = Form.useForm();
  return (
    <Modal title="满意度评价" open={open} onCancel={onCancel} onOk={() => form.submit()} destroyOnClose>
      <Form
        form={form}
        layout="vertical"
        onFinish={(v) => ononSubmitWrap(v, onSubmit)}
      >
        <Form.Item name="rating" label="评分" rules={[{ required: true, message: '请评分' }]}>
          <Rate />
        </Form.Item>
        <Form.Item name="comment" label="评价">
          <Input.TextArea rows={3} maxLength={500} showCount />
        </Form.Item>
      </Form>
    </Modal>
  );
};

function ononSubmitWrap(
  v: { rating: number; comment?: string },
  onSubmit: (rating: number, comment?: string) => Promise<void>,
) {
  void onSubmit(v.rating, v.comment);
}

void InputNumber;

export default DetailPage;
