import React, { useEffect, useState } from 'react';
import { Button, Descriptions, Drawer, Form, Input, Select, Timeline, message } from 'antd';
import {
  getApplicationHistory,
  resubmitApplication,
} from '@/api/member';
import type { MemberApplication, MemberApplicationHistory } from '@/types/api';
import StatusTag from '@/components/StatusTag/StatusTag';
import AdmissionPanel from './AdmissionPanel';
import dayjs from 'dayjs';
import { Link } from 'react-router-dom';
import { ApplicationStatusMap } from '../meta';

interface ReviewDrawerProps {
  open: boolean;
  record: MemberApplication | null;
  mode: 'view' | 'review' | 'resubmit';
  onClose: () => void;
  onDone: () => void;
}

const ReviewDrawer: React.FC<ReviewDrawerProps> = ({ open, record, mode, onClose, onDone }) => {
  const [form] = Form.useForm();
  const [history, setHistory] = useState<MemberApplicationHistory[]>([]);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (!open || !record) return;
    if (mode !== 'view') {
      form.resetFields();
    }
    let active = true;
    setHistory([]);
    getApplicationHistory(record.id)
      .then(rows => { if (active) setHistory(rows); })
      .catch(() => { if (active) setHistory([]); });
    return () => { active = false; };
  }, [open, record, form, mode]);

  const run = async (fn: () => Promise<unknown>, ok: string) => {
    setLoading(true);
    try {
      await fn();
      message.success(ok);
      onDone();
    } finally {
      setLoading(false);
    }
  };

  const handleResubmit = () =>
    run(async () => {
      const values = await form.validateFields();
      await resubmitApplication(record!.id, values);
    }, '已重新提交');

  return (
    <Drawer
      title={record ? `${record.real_name} 的申请` : '申请详情'}
      width={640}
      styles={{ wrapper: { maxWidth: '100vw' } }}
      open={open}
      onClose={onClose}
    >
      {record && (
        <>
          <Descriptions column={1} size="small" bordered>
            <Descriptions.Item label="类型">{record.applicant_type === 2 ? '干事' : '会员'}</Descriptions.Item>
            <Descriptions.Item label="学号">{record.student_no}</Descriptions.Item>
            <Descriptions.Item label="部门">{record.department_name || '-'}</Descriptions.Item>
            <Descriptions.Item label="状态">
              <StatusTag status={record.status} mapping={ApplicationStatusMap} /> {record.current_stage}
            </Descriptions.Item>
            <Descriptions.Item label="电话">{record.contact_phone}</Descriptions.Item>
            <Descriptions.Item label="邮箱">{record.contact_email}</Descriptions.Item>
            <Descriptions.Item label="技能">{(record.skills || []).join('、') || '-'}</Descriptions.Item>
            <Descriptions.Item label="理由">{record.reason}</Descriptions.Item>
            <Descriptions.Item label="经历">{record.experience || '-'}</Descriptions.Item>
            {record.review_comment && (
              <Descriptions.Item label="审核意见">{record.review_comment}</Descriptions.Item>
            )}
          </Descriptions>
          {record.flow_instance_id && <p><Link to={`/workflow/instances?instance_id=${encodeURIComponent(record.flow_instance_id)}`}>查看关联审批流程</Link></p>}

          <AdmissionPanel key={record.id} id={record.id} officer={record.applicant_type === 2} editable={mode === 'review'} onChanged={onDone} />

          {mode === 'resubmit' && (
            <Form form={form} layout="vertical" style={{ marginTop: 16 }} initialValues={record}>
              <Form.Item name="real_name" label="姓名"><Input /></Form.Item>
              <Form.Item name="student_no" label="学号"><Input /></Form.Item>
              <Form.Item name="contact_phone" label="联系电话"><Input /></Form.Item>
              <Form.Item name="contact_email" label="邮箱" rules={[{ type: 'email' }]}><Input /></Form.Item>
              <Form.Item name="experience" label="项目经历">
                <Input.TextArea rows={3} />
              </Form.Item>
              <Form.Item name="skills" label="技能">
                <Select mode="tags" />
              </Form.Item>
              <Form.Item name="reason" label="申请理由">
                <Input.TextArea rows={3} />
              </Form.Item>
              <Button type="primary" loading={loading} onClick={() => { void handleResubmit().catch(() => undefined); }}>
                重新提交
              </Button>
            </Form>
          )}

          <Timeline
            style={{ marginTop: 24 }}
            items={history.map((h) => ({
              children: `${dayjs(h.created_at).format('YYYY-MM-DD HH:mm')}：${ApplicationStatusMap[h.from_status]?.text || h.from_status} → ${ApplicationStatusMap[h.to_status]?.text || h.to_status} ${h.comment || ''}`,
            }))}
          />
        </>
      )}
    </Drawer>
  );
};

export default ReviewDrawer;
