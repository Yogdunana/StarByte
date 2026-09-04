import React, { useEffect, useState } from 'react';
import { Button, Descriptions, Drawer, Form, Input, Select, Space, Timeline, message } from 'antd';
import {
  approveApplication,
  getApplicationHistory,
  rejectApplication,
  resubmitApplication,
  supplementApplication,
} from '@/api/member';
import type { MemberApplication, MemberApplicationHistory } from '@/types/api';
import StatusTag from '@/components/StatusTag/StatusTag';
import { ApplicationStatusMap, requiredFieldOptions } from '../meta';

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
    getApplicationHistory(record.id)
      .then(setHistory)
      .catch(() => setHistory([]));
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

  const handleApprove = () =>
    run(async () => {
      const { comment } = await form.validateFields(['comment']);
      await approveApplication(record!.id, comment || '');
    }, '已通过');

  const handleReject = () =>
    run(async () => {
      const { comment } = await form.validateFields(['comment']);
      await rejectApplication(record!.id, comment || '');
    }, '已拒绝');

  const handleSupplement = () =>
    run(async () => {
      const values = await form.validateFields(['comment', 'required_fields']);
      await supplementApplication(record!.id, values.comment, values.required_fields || []);
    }, '已要求补充材料');

  const handleResubmit = () =>
    run(async () => {
      const values = await form.validateFields();
      await resubmitApplication(record!.id, values);
    }, '已重新提交');

  return (
    <Drawer
      title={record ? `${record.real_name} 的申请` : '申请详情'}
      width={560}
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
              <StatusTag status={record.status} mapping={ApplicationStatusMap} />
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

          {mode === 'review' && (
            <Form form={form} layout="vertical" style={{ marginTop: 16 }}>
              <Form.Item name="comment" label="审核意见">
                <Input.TextArea rows={3} />
              </Form.Item>
              <Form.Item name="required_fields" label="需补充字段">
                <Select mode="multiple" options={requiredFieldOptions} placeholder="要求补充时选择" />
              </Form.Item>
              <Space wrap>
                <Button type="primary" loading={loading} onClick={() => void handleApprove()}>
                  通过
                </Button>
                <Button danger loading={loading} onClick={() => void handleReject()}>
                  拒绝
                </Button>
                <Button loading={loading} onClick={() => void handleSupplement()}>
                  补充材料
                </Button>
              </Space>
            </Form>
          )}

          {mode === 'resubmit' && (
            <Form form={form} layout="vertical" style={{ marginTop: 16 }} initialValues={record}>
              <Form.Item name="experience" label="项目经历">
                <Input.TextArea rows={3} />
              </Form.Item>
              <Form.Item name="skills" label="技能">
                <Select mode="tags" />
              </Form.Item>
              <Form.Item name="reason" label="申请理由">
                <Input.TextArea rows={3} />
              </Form.Item>
              <Button type="primary" loading={loading} onClick={() => void handleResubmit()}>
                重新提交
              </Button>
            </Form>
          )}

          <Timeline
            style={{ marginTop: 24 }}
            items={history.map((h) => ({
              children: `${h.created_at}：${h.from_status} → ${h.to_status} ${h.comment || ''}`,
            }))}
          />
        </>
      )}
    </Drawer>
  );
};

export default ReviewDrawer;
