import React, { useEffect, useState } from 'react';
import { Button, Card, Empty, Form, Input, InputNumber, Progress, Radio, Space, Switch, message } from 'antd';
import ReactECharts from 'echarts-for-react';
import { usePermission } from '@/hooks/usePermission';
import { castVote, closeVote, createVote, getVoteResult } from '@/api/meeting';
import type { MeetingVote, VoteResult } from '@/types/api';
import { VoteStatusMap, VoteTypeMap } from './meta';
import StatusTag from '@/components/StatusTag/StatusTag';

interface Props {
  meetingId: string;
  votes: MeetingVote[];
  onRefresh: () => Promise<void>;
}

const VotePanel: React.FC<Props> = ({ meetingId, votes, onRefresh }) => {
  const canManage = usePermission('meeting:manage');
  const [form] = Form.useForm();
  const [results, setResults] = useState<Record<string, VoteResult>>({});
  const [now, setNow] = useState(Date.now());

  useEffect(() => {
    const t = window.setInterval(() => setNow(Date.now()), 1000);
    return () => window.clearInterval(t);
  }, []);

  useEffect(() => {
    votes.forEach((v) => {
      getVoteResult(v.id).then((r) => setResults((prev) => ({ ...prev, [v.id]: r }))).catch(() => undefined);
    });
  }, [votes]);

  const remain = (v: MeetingVote) => {
    if (!v.end_time || v.status !== 1) return '';
    const ms = new Date(v.end_time).getTime() - now;
    if (ms <= 0) return '已到期';
    const s = Math.floor(ms / 1000);
    return `剩余 ${Math.floor(s / 60)}:${String(s % 60).padStart(2, '0')}`;
  };

  return (
    <Space direction="vertical" style={{ width: '100%' }} size={16}>
      {canManage && (
        <Card size="small" title="发起投票">
          <Form
            form={form}
            layout="vertical"
            initialValues={{ vote_type: 1, duration: 300, is_anonymous: false }}
            onFinish={async (values) => {
              const labels = String(values.options || '').split('\n').map((s) => s.trim()).filter(Boolean);
              if (labels.length < 2) {
                message.error('至少两个选项，每行一个');
                return;
              }
              await createVote(meetingId, {
                title: values.title,
                description: values.description,
                vote_type: values.vote_type,
                is_anonymous: values.is_anonymous,
                duration: values.duration,
                options: labels.map((label, i) => ({ key: `opt${i + 1}`, label })),
              });
              message.success('投票已开始');
              form.resetFields();
              await onRefresh();
            }}
          >
            <Form.Item name="title" label="标题" rules={[{ required: true }]}><Input /></Form.Item>
            <Form.Item name="description" label="说明"><Input.TextArea rows={2} /></Form.Item>
            <Form.Item name="vote_type" label="类型">
              <Radio.Group options={[{ value: 1, label: '等权' }, { value: 2, label: '加权' }]} />
            </Form.Item>
            <Form.Item name="is_anonymous" label="匿名" valuePropName="checked"><Switch /></Form.Item>
            <Form.Item name="duration" label="时长（秒，0 为手动关闭）"><InputNumber min={0} style={{ width: '100%' }} /></Form.Item>
            <Form.Item name="options" label="选项（每行一个）" rules={[{ required: true }]}>
              <Input.TextArea rows={3} placeholder={'Web 开发\nAI/ML'} />
            </Form.Item>
            <Button type="primary" htmlType="submit">创建并开始</Button>
          </Form>
        </Card>
      )}
      {votes.length === 0 && <Empty description="暂无投票" />}
      {votes.map((v) => {
        const result = results[v.id];
        const max = result ? Math.max(...result.results.map((r) => r.weight_total), 1) : 1;
        return (
          <Card
            key={v.id}
            size="small"
            title={`${v.title} · ${VoteTypeMap[v.vote_type]}`}
            extra={<Space><StatusTag status={v.status} mapping={VoteStatusMap} /><span>{remain(v)}</span></Space>}
          >
            <div style={{ marginBottom: 8 }}>{v.description}</div>
            {v.status === 1 && !v.has_voted && (
              <Space wrap>
                {v.options.map((o) => (
                  <Button key={o.key} onClick={() => castVote(v.id, o.key).then(() => { message.success('已投票'); void onRefresh(); })}>
                    {o.label}
                  </Button>
                ))}
              </Space>
            )}
            {v.has_voted && <div style={{ marginBottom: 8 }}>你已投票{v.is_anonymous ? '（匿名）' : ''}</div>}
            {canManage && v.status === 1 && (
              <Button size="small" onClick={() => closeVote(v.id).then(onRefresh)}>关闭投票</Button>
            )}
            {result && (
              <>
                {result.results.map((r) => (
                  <div key={r.option_key} style={{ marginTop: 8 }}>
                    <div>{r.option_label} · {r.count} 票 · 权重 {r.weight_total}</div>
                    <Progress percent={Math.round((r.weight_total / max) * 100)} />
                  </div>
                ))}
                <ReactECharts
                  style={{ height: 260, marginTop: 12 }}
                  option={{
                    tooltip: { trigger: 'item' },
                    legend: { bottom: 0 },
                    series: [{
                      type: 'pie',
                      radius: ['35%', '65%'],
                      data: result.results.map((r) => ({ name: r.option_label, value: r.weight_total })),
                    }],
                  }}
                />
              </>
            )}
          </Card>
        );
      })}
    </Space>
  );
};

export default VotePanel;
