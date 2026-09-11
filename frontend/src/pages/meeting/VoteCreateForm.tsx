import { useState } from 'react';
import { Button, Form, Input, InputNumber, Radio, Switch, message } from 'antd';
import { createVote } from '@/api/meeting';
import type { CreateVoteParams } from '@/types/api';
import styles from './Voting.module.css';
interface Values extends Omit<CreateVoteParams, 'options'> { option_labels: string }
interface Props { meetingId: string; onCreated: () => Promise<void> }
export default function VoteCreateForm({ meetingId, onCreated }: Props) {
  const [form] = Form.useForm<Values>();
  const [busy, setBusy] = useState(false);
  const submit = async (values: Values) => {
    if (busy) return;
    const labels = values.option_labels.split('\n').map(label => label.trim()).filter(Boolean);
    if (labels.length < 2 || new Set(labels).size !== labels.length) { message.error('至少填写两个不重复的选项，每行一个'); return; }
    setBusy(true);
    try {
      await createVote(meetingId, { title: values.title, description: values.description, vote_type: values.vote_type, is_anonymous: values.is_anonymous, duration: values.duration, options: labels.map((label, index) => ({ key: `opt${index + 1}`, label })) });
      form.resetFields(); message.success('投票已开始'); await onCreated();
    } catch { /* The API layer displays the concrete error. Keep the draft. */ }
    finally { setBusy(false); }
  };
  return <Form form={form} layout="vertical" onFinish={values => void submit(values)} initialValues={{ vote_type: 1, duration: 300, is_anonymous: true }}>
    <Form.Item name="title" label="投票主题" rules={[{ required: true, whitespace: true }, { max: 200 }]}><Input placeholder="希望大家一起决定什么？" /></Form.Item>
    <Form.Item name="description" label="背景与说明"><Input.TextArea rows={2} /></Form.Item>
    <div className={styles.formColumns}>
      <Form.Item name="vote_type" label="计票方式"><Radio.Group options={[{ value: 1, label: '每人一票' }, { value: 2, label: '按职务加权' }]} /></Form.Item>
      <Form.Item name="is_anonymous" label="匿名投票" valuePropName="checked"><Switch /></Form.Item>
    </div>
    <Form.Item name="option_labels" label="备选方案" extra="每行一个选项，至少两个。" rules={[{ required: true, whitespace: true }]}><Input.TextArea rows={4} placeholder={'方案一\n方案二'} /></Form.Item>
    <Form.Item name="duration" label="开放时长（秒）" extra="填 0 表示由会议管理者手动截止。" rules={[{ required: true }, { type: 'number', min: 0 }]}><InputNumber precision={0} min={0} style={{ width: '100%' }} /></Form.Item>
    <p className={styles.privacy}>开始时固定参会名单和每人的权重。请先确认参会人，后添加的人员不参与本轮。</p>
    <Button type="primary" htmlType="submit" loading={busy}>发布并开始投票</Button>
  </Form>;
}
