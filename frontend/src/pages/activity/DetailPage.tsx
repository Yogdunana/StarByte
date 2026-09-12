import { tx, useLocale } from '@/i18n/text';
import React, { useCallback, useEffect, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import {
  Button,
  Card,
  Descriptions,
  Form,
  Input,
  InputNumber,
  Modal,
  Rate,
  Space,
  Statistic,
  Table,
  Tabs,
  Tag,
  message,
} from 'antd';
import { QrcodeOutlined } from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import StatusTag from '@/components/StatusTag/StatusTag';
import { usePermission } from '@/hooks/usePermission';
import {
  approveRegistration,
  cancelActivity,
  cancelRegistration,
  checkinActivity,
  deleteActivity,
  endActivity,
  getActivityDetail,
  getActivityQRCode,
  getActivityStats,
  getMyRegistration,
  getRegistrations,
  registerActivity,
  startActivity,
  submitSurvey,
} from '@/api/activity';
import type { Activity, ActivityQRCode, ActivityStats, Registration } from '@/api/activity';
import { formatDateTime } from '@/utils/format';
import { ActivityStatusMap, RegistrationStatusMap, registerSuccessText } from './meta';

const DetailPage: React.FC = () => {
  useLocale();
  const { id = '' } = useParams();
  const nav = useNavigate();
  const canManage = usePermission('activity:manage');
  const canUpdate = usePermission('activity:update');
  const canDelete = usePermission('activity:delete');
  const [activity, setActivity] = useState<Activity | null>(null);
  const [myReg, setMyReg] = useState<Registration | null>(null);
  const [regs, setRegs] = useState<Registration[]>([]);
  const [stats, setStats] = useState<ActivityStats | null>(null);
  const [surveyOpen, setSurveyOpen] = useState(false);
  const [qr, setQr] = useState<ActivityQRCode | null>(null);
  const [gpsOpen, setGpsOpen] = useState(false);

  const load = useCallback(async () => {
    if (!id) return;
    const a = await getActivityDetail(id);
    setActivity(a);
    const mine = await getMyRegistration(id);
    setMyReg(mine);
    if (canManage) {
      const [r, s] = await Promise.all([getRegistrations(id), getActivityStats(id)]);
      setRegs(r);
      setStats(s);
    } else {
      setRegs([]);
      setStats(null);
    }
  }, [id, canManage]);

  useEffect(() => {
    void load();
  }, [load]);

  if (!activity) return <Card loading />;

  const approved = myReg?.status === 1;
  const cancelled = !myReg || myReg.status === 4;
  const checkedIn = myReg?.checkin_status === 1;
  const canRegister = activity.status === 1 && cancelled;
  const canCancelMine = Boolean(myReg) && !cancelled && [1, 2].includes(activity.status);
  const canQRCheckin = approved && !checkedIn && [1, 2].includes(activity.status);
  const canGPSCheckin = canQRCheckin && activity.gps_enabled;
  const canSurvey = activity.status === 3 && checkedIn;

  const regColumns: ColumnsType<Registration> = [
    { title: tx('姓名'), key: 'name', render: (_, r) => r.user?.name || '-' },
    {
      title: tx('报名状态'),
      dataIndex: 'status',
      width: 100,
      render: (v: number) => <StatusTag status={v} mapping={RegistrationStatusMap} />,
    },
    {
      title: tx('签到'),
      dataIndex: 'checkin_status',
      width: 100,
      render: (v: number) => (v === 1 ? tx('已签到') : tx('未签到')),
    },
    {
      title: tx('签到时间'),
      dataIndex: 'checked_in_at',
      width: 160,
      render: (v?: string) => formatDateTime(v, 'YYYY-MM-DD HH:mm'),
    },
    {
      title: tx('操作'),
      width: 180,
      render: (_, r) =>
        canManage && (
          <Space>
            {r.status !== 1 && (
              <Button
                size="small"
                type="link"
                onClick={() => approveRegistration(id, r.user.id, true).then(load)}
              >
                {tx('通过')}
              </Button>
            )}
            {r.status !== 2 && (
              <Button
                size="small"
                type="link"
                danger
                onClick={() => approveRegistration(id, r.user.id, false).then(load)}
              >
                {tx('拒绝')}
              </Button>
            )}
          </Space>
        ),
    },
  ];

  const tabs = [
    {
      key: 'detail',
      label: tx('活动说明'),
      children: (
        <Descriptions column={1} bordered size="small">
          <Descriptions.Item label={tx('说明')}>{activity.description || '-'}</Descriptions.Item>
          <Descriptions.Item label={tx('我的报名')}>
            {myReg ? (
              <StatusTag status={myReg.status} mapping={RegistrationStatusMap} />
            ) : (
              tx('未报名')
            )}
            {checkedIn ? tx(' · 已签到') : ''}
          </Descriptions.Item>
          <Descriptions.Item label={tx('GPS 围栏')}>
            {activity.gps_enabled
              ? tx('{{value0}}, {{value1}} / {{value2}} 米', {
                  value0: activity.latitude,
                  value1: activity.longitude,
                  value2: activity.checkin_radius_m,
                })
              : tx('未配置（GPS 签到不可用）')}
          </Descriptions.Item>
        </Descriptions>
      ),
    },
  ];
  if (canManage) {
    tabs.push({
      key: 'regs',
      label: tx('报名管理（{{value0}}）', { value0: regs.length }),
      children: <Table rowKey="id" dataSource={regs} columns={regColumns} pagination={false} />,
    });
    tabs.push({
      key: 'stats',
      label: tx('统计'),
      children: stats ? (
        <Space wrap size="large">
          <Statistic title={tx('已通过')} value={stats.approved_count} />
          <Statistic title={tx('候补')} value={stats.waitlist_count} />
          <Statistic title={tx('已签到')} value={stats.checked_in_count} />
          <Statistic title={tx('报名率')} value={stats.register_rate} suffix="%" precision={1} />
          <Statistic title={tx('出席率')} value={stats.attend_rate} suffix="%" precision={1} />
          <Statistic title={tx('评价数')} value={stats.survey_count} />
          <Statistic title={tx('平均评分')} value={stats.avg_rating} precision={1} />
        </Space>
      ) : (
        <span />
      ),
    });
  }

  return (
    <Card
      title={activity.title}
      extra={<Button onClick={() => nav('/activity/list')}>{tx('返回列表')}</Button>}
    >
      <Space wrap style={{ marginBottom: 16 }}>
        <StatusTag status={activity.status} mapping={ActivityStatusMap} />
        {activity.category && <Tag>{activity.category}</Tag>}
        <span>
          {tx('地点：')}
          {activity.location || '-'}
        </span>
        <span>
          {tx('组织者：')}
          {activity.organizer?.name || '-'}
        </span>
        <span>
          {tx('开始：')}
          {formatDateTime(activity.start_time, 'YYYY-MM-DD HH:mm')}
        </span>
        <span>
          {tx('结束：')}
          {formatDateTime(activity.end_time, 'YYYY-MM-DD HH:mm')}
        </span>
        <span>
          {tx('报名：')}
          {activity.registered_count}/
          {activity.max_participants === 0 ? tx('不限') : activity.max_participants}
        </span>
      </Space>

      <Space wrap style={{ marginBottom: 16 }}>
        {canRegister && (
          <Button
            type="primary"
            onClick={() =>
              registerActivity(id).then((reg) => {
                message.success(registerSuccessText(reg.status));
                return load();
              })
            }
          >
            {tx('我要报名')}
          </Button>
        )}
        {canCancelMine && (
          <Button
            onClick={() =>
              cancelRegistration(id).then(() => {
                message.success(tx('已取消报名'));
                return load();
              })
            }
          >
            {tx('取消报名')}
          </Button>
        )}
        {canQRCheckin && (
          <Button onClick={() => nav(`/activity/checkin?activity_id=${id}`)}>
            {tx('前往签到')}
          </Button>
        )}
        {canGPSCheckin && <Button onClick={() => setGpsOpen(true)}>{tx('GPS 签到')}</Button>}
        {canSurvey && (
          <Button type="primary" ghost onClick={() => setSurveyOpen(true)}>
            {tx('满意度评价')}
          </Button>
        )}
        {canManage && [1, 2].includes(activity.status) && (
          <Button
            icon={<QrcodeOutlined />}
            onClick={() => getActivityQRCode(id).then((code) => setQr(code))}
          >
            {tx('签到二维码')}
          </Button>
        )}
        {canUpdate && activity.status === 1 && (
          <Button
            onClick={() =>
              startActivity(id).then(() => {
                message.success(tx('活动已开始'));
                return load();
              })
            }
          >
            {tx('开始')}
          </Button>
        )}
        {canUpdate && activity.status === 2 && (
          <Button
            onClick={() =>
              endActivity(id).then(() => {
                message.success(tx('活动已结束'));
                return load();
              })
            }
          >
            {tx('结束')}
          </Button>
        )}
        {canUpdate && [0, 1, 2].includes(activity.status) && (
          <Button
            onClick={() =>
              cancelActivity(id).then(() => {
                message.success(tx('活动已取消'));
                return load();
              })
            }
          >
            {tx('取消活动')}
          </Button>
        )}
        {canDelete && [0, 3, 4].includes(activity.status) && (
          <Button
            danger
            onClick={() =>
              deleteActivity(id).then(() => {
                message.success(tx('已删除'));
                nav('/activity/list');
              })
            }
          >
            {tx('删除')}
          </Button>
        )}
      </Space>

      {activity.tags?.length > 0 && (
        <Space style={{ marginBottom: 16 }}>
          {activity.tags.map((t) => (
            <Tag key={t} color="blue">
              {t}
            </Tag>
          ))}
        </Space>
      )}

      <Tabs items={tabs} />

      <SurveyModal
        open={surveyOpen}
        onCancel={() => setSurveyOpen(false)}
        onSubmit={async (rating, comment) => {
          await submitSurvey(id, { rating, comment });
          message.success(tx('评价已提交'));
          setSurveyOpen(false);
          await load();
        }}
      />
      <Modal
        title={tx('活动签到二维码')}
        open={Boolean(qr)}
        onCancel={() => setQr(null)}
        footer={null}
      >
        {qr && (
          <Space direction="vertical" align="center" style={{ width: '100%' }}>
            <img
              alt={tx('活动签到二维码')}
              src={`data:image/png;base64,${qr.png_base64}`}
              style={{ width: 240 }}
            />
            <div>
              {tx('过期时间：')}
              {formatDateTime(qr.expires_at, 'YYYY-MM-DD HH:mm:ss')}
            </div>
          </Space>
        )}
      </Modal>
      <GPSCheckinModal
        open={gpsOpen}
        onCancel={() => setGpsOpen(false)}
        onSubmit={async (latitude, longitude) => {
          await checkinActivity(id, { method: 2, latitude, longitude });
          message.success(tx('签到成功'));
          setGpsOpen(false);
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
  useLocale();
  const [form] = Form.useForm();
  return (
    <Modal
      title={tx('满意度评价')}
      open={open}
      onCancel={onCancel}
      onOk={() => form.submit()}
      destroyOnClose
    >
      <Form
        form={form}
        layout="vertical"
        onFinish={(v: { rating: number; comment?: string }) => {
          void onSubmit(v.rating, v.comment);
        }}
      >
        <Form.Item
          name="rating"
          label={tx('评分')}
          rules={[{ required: true, message: tx('请评分') }]}
        >
          <Rate />
        </Form.Item>
        <Form.Item name="comment" label={tx('评价')}>
          <Input.TextArea rows={3} maxLength={500} showCount />
        </Form.Item>
      </Form>
    </Modal>
  );
};

const GPSCheckinModal: React.FC<{
  open: boolean;
  onCancel: () => void;
  onSubmit: (latitude: number, longitude: number) => Promise<void>;
}> = ({ open, onCancel, onSubmit }) => {
  useLocale();
  const [form] = Form.useForm();
  const fillBrowser = () => {
    if (!navigator.geolocation) {
      message.error(tx('浏览器不支持定位'));
      return;
    }
    navigator.geolocation.getCurrentPosition(
      (pos) => {
        form.setFieldsValue({ latitude: pos.coords.latitude, longitude: pos.coords.longitude });
      },
      () => message.error(tx('无法获取定位')),
    );
  };
  return (
    <Modal
      title={tx('GPS 签到')}
      open={open}
      onCancel={onCancel}
      onOk={() => form.submit()}
      destroyOnClose
    >
      <Form
        form={form}
        layout="vertical"
        onFinish={(v: { latitude: number; longitude: number }) => {
          void onSubmit(v.latitude, v.longitude);
        }}
      >
        <Button onClick={fillBrowser} style={{ marginBottom: 12 }}>
          {tx('使用当前定位')}
        </Button>
        <Form.Item name="latitude" label={tx('纬度')} rules={[{ required: true }]}>
          <InputNumber style={{ width: '100%' }} />
        </Form.Item>
        <Form.Item name="longitude" label={tx('经度')} rules={[{ required: true }]}>
          <InputNumber style={{ width: '100%' }} />
        </Form.Item>
      </Form>
    </Modal>
  );
};

export default DetailPage;
