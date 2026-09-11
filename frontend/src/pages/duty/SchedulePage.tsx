import { useState, useEffect, useCallback } from 'react';
import { Card, Table, Button, Modal, Form, Input, Select, DatePicker, Space, Tag, message, Tooltip } from 'antd';
import { PlusOutlined, SwapOutlined, CalendarOutlined, BarChartOutlined } from '@ant-design/icons';
import { Dayjs } from 'dayjs';
import {
  getScheduleList, createSchedule, updateSchedule, createSwapRequest, getDutyStats,
  type Schedule, type DutyStats,
} from '@/api/duty';

const { RangePicker } = DatePicker;

const statusColors: Record<number, string> = { 0: 'default', 1: 'processing', 2: 'success', 3: 'error', 4: 'warning' };
const slotLabels: Record<string, string> = { morning: '上午', afternoon: '下午', evening: '晚上', full_day: '全天' };

export default function SchedulePage() {
  const [schedules, setSchedules] = useState<Schedule[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(false);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(20);
  const [view, setView] = useState<'day' | 'week' | 'month'>('week');
  const [createModalOpen, setCreateModalOpen] = useState(false);
  const [swapModalOpen, setSwapModalOpen] = useState(false);
  const [statsModalOpen, setStatsModalOpen] = useState(false);
  const [stats, setStats] = useState<DutyStats | null>(null);
  const [selectedSchedule, setSelectedSchedule] = useState<Schedule | null>(null);
  const [createForm] = Form.useForm();
  const [swapForm] = Form.useForm();

  const fetchData = useCallback(async () => {
    setLoading(true);
    try {
      const res = await getScheduleList({ page, page_size: pageSize, view });
      setSchedules(res.data.list || []);
      setTotal(res.data.total || 0);
    } catch {
      message.error('加载排班列表失败');
    } finally {
      setLoading(false);
    }
  }, [page, pageSize, view]);

  useEffect(() => { fetchData(); }, [fetchData]);

  const handleCreate = async () => {
    try {
      const values = await createForm.validateFields();
      await createSchedule({
        ...values,
        duty_date: values.duty_date.format('YYYY-MM-DD'),
      });
      message.success('排班创建成功');
      setCreateModalOpen(false);
      createForm.resetFields();
      fetchData();
    } catch (err: any) {
      if (err?.errorFields) return;
      message.error(err?.message || '创建失败');
    }
  };

  const handleDragSchedule = async (id: string, newDate: string, newSlot: string) => {
    try {
      await updateSchedule(id, { duty_date: newDate, time_slot: newSlot });
      message.success('排班调整成功');
      fetchData();
    } catch {
      message.error('调整失败');
    }
  };

  const handleSwap = async () => {
    try {
      const values = await swapForm.validateFields();
      if (!selectedSchedule) return;
      await createSwapRequest({
        target_user_id: values.target_user_id,
        requester_schedule_id: selectedSchedule.id,
        reason: values.reason,
      });
      message.success('调班申请已提交');
      setSwapModalOpen(false);
      swapForm.resetFields();
    } catch (err: any) {
      if (err?.errorFields) return;
      message.error(err?.message || '提交失败');
    }
  };

  const handleStats = async () => {
    try {
      const res = await getDutyStats({});
      setStats(res.data);
      setStatsModalOpen(true);
    } catch {
      message.error('加载统计失败');
    }
  };

  const columns = [
    { title: '日期', dataIndex: 'duty_date', key: 'duty_date' },
    {
      title: '时段', dataIndex: 'time_slot', key: 'time_slot',
      render: (slot: string) => slotLabels[slot] || slot,
    },
    { title: '地点', dataIndex: 'location', key: 'location' },
    { title: '备注', dataIndex: 'remark', key: 'remark', ellipsis: true },
    {
      title: '状态', dataIndex: 'status', key: 'status',
      render: (status: number, record: Schedule) => (
        <Tag color={statusColors[status]}>{record.status_text}</Tag>
      ),
    },
    {
      title: '操作', key: 'action',
      render: (_: unknown, record: Schedule) => (
        <Space>
          {record.status === 0 && (
            <Button size="small" icon={<SwapOutlined />}
              onClick={() => { setSelectedSchedule(record); setSwapModalOpen(true); }}>
              调班
            </Button>
          )}
        </Space>
      ),
    },
  ];

  return (
    <div>
      <Card
        title="排班管理"
        extra={
          <Space>
            <Select value={view} onChange={setView} style={{ width: 100 }}
              options={[{ value: 'day', label: '日视图' }, { value: 'week', label: '周视图' }, { value: 'month', label: '月视图' }]}
            />
            <Button icon={<PlusOutlined />} onClick={() => setCreateModalOpen(true)}>创建排班</Button>
            <Button icon={<BarChartOutlined />} onClick={handleStats}>值班统计</Button>
          </Space>
        }
      >
        <Table
          dataSource={schedules}
          columns={columns}
          rowKey="id"
          loading={loading}
          pagination={{ current: page, pageSize, total, onChange: (p, ps) => { setPage(p); setPageSize(ps); } }}
        />
      </Card>

      <Modal title="创建排班" open={createModalOpen} onOk={handleCreate} onCancel={() => setCreateModalOpen(false)}>
        <Form form={createForm} layout="vertical">
          <Form.Item name="user_id" label="用户ID" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="duty_date" label="值班日期" rules={[{ required: true }]}>
            <DatePicker style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="time_slot" label="时段" rules={[{ required: true }]}
            initialValue="full_day">
            <Select options={Object.entries(slotLabels).map(([k, v]) => ({ value: k, label: v }))} />
          </Form.Item>
          <Form.Item name="location" label="地点">
            <Input />
          </Form.Item>
          <Form.Item name="remark" label="备注">
            <Input.TextArea />
          </Form.Item>
        </Form>
      </Modal>

      <Modal title="调班申请" open={swapModalOpen} onOk={handleSwap} onCancel={() => setSwapModalOpen(false)}>
        <Form form={swapForm} layout="vertical">
          <Form.Item name="target_user_id" label="目标用户ID" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="reason" label="调班原因" rules={[{ required: true, min: 5 }]}>
            <Input.TextArea />
          </Form.Item>
        </Form>
      </Modal>

      <Modal title="值班统计" open={statsModalOpen} onCancel={() => setStatsModalOpen(false)} footer={null} width={700}>
        {stats && (
          <div>
            <Space style={{ marginBottom: 16 }}>
              <Tag>总值班: {stats.total_duties}</Tag>
              <Tag color="success">已完成: {stats.completed_duties}</Tag>
              <Tag color="error">缺勤: {stats.absent_duties}</Tag>
              <Tag color="processing">完成率: {stats.completion_rate.toFixed(1)}%</Tag>
            </Space>
            <Table
              dataSource={stats.user_stats}
              rowKey="user_id"
              pagination={false}
              columns={[
                { title: '用户ID', dataIndex: 'user_id', key: 'user_id' },
                { title: '总次数', dataIndex: 'total_count', key: 'total_count' },
                { title: '已完成', dataIndex: 'completed', key: 'completed' },
                { title: '缺勤', dataIndex: 'absent', key: 'absent' },
                { title: '完成率', dataIndex: 'rate', key: 'rate',
                  render: (r: number) => `${r.toFixed(1)}%` },
              ]}
            />
          </div>
        )}
      </Modal>
    </div>
  );
}
