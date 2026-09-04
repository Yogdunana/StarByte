import React, { useEffect, useState } from 'react';
import { Button, Form, Input, Select, Steps, Tag, message } from 'antd';
import { getMemberDepartments, submitApplication } from '@/api/member';
import type { CreateMemberApplicationParams, MemberDepartmentOption } from '@/types/api';

const { TextArea } = Input;

interface ApplicationFormProps {
  onSubmitted?: () => void;
}

const ApplicationForm: React.FC<ApplicationFormProps> = ({ onSubmitted }) => {
  const [form] = Form.useForm<CreateMemberApplicationParams>();
  const [step, setStep] = useState(0);
  const [submitting, setSubmitting] = useState(false);
  const [departments, setDepartments] = useState<MemberDepartmentOption[]>([]);

  useEffect(() => {
    getMemberDepartments()
      .then(setDepartments)
      .catch(() => undefined);
  }, []);

  const next = async () => {
    const fields =
      step === 0
        ? (['applicant_type', 'real_name', 'student_no', 'department_id'] as const)
        : (['contact_phone', 'contact_email'] as const);
    await form.validateFields([...fields]);
    setStep((s) => s + 1);
  };

  const handleSubmit = async () => {
    const values = await form.validateFields();
    setSubmitting(true);
    try {
      await submitApplication({
        ...values,
        skills: values.skills || [],
      });
      message.success('申请已提交');
      form.resetFields();
      setStep(0);
      onSubmitted?.();
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <>
      <Steps
        current={step}
        style={{ marginBottom: 24 }}
        items={[{ title: '基本信息' }, { title: '联系方式' }, { title: '申请材料' }]}
      />
      <Form form={form} layout="vertical" initialValues={{ applicant_type: 1, skills: [] }}>
        <div style={{ display: step === 0 ? 'block' : 'none' }}>
          <Form.Item name="applicant_type" label="申请类型" rules={[{ required: true }]}>
            <Select
              options={[
                { value: 1, label: '会员' },
                { value: 2, label: '干事（需面试）' },
              ]}
            />
          </Form.Item>
          <Form.Item name="real_name" label="姓名" rules={[{ required: true, max: 50 }]}>
            <Input placeholder="真实姓名" />
          </Form.Item>
          <Form.Item name="student_no" label="学号" rules={[{ required: true, max: 30 }]}>
            <Input placeholder="学号" />
          </Form.Item>
          <Form.Item name="department_id" label="意向部门">
            <Select
              allowClear
              placeholder="选择部门"
              options={departments.map((d) => ({ value: d.id, label: d.name }))}
            />
          </Form.Item>
        </div>
        <div style={{ display: step === 1 ? 'block' : 'none' }}>
          <Form.Item name="contact_phone" label="手机号" rules={[{ required: true, max: 20 }]}>
            <Input placeholder="11 位手机号" />
          </Form.Item>
          <Form.Item
            name="contact_email"
            label="邮箱"
            rules={[{ required: true, type: 'email', max: 100 }]}
          >
            <Input placeholder="联系邮箱" />
          </Form.Item>
        </div>
        <div style={{ display: step === 2 ? 'block' : 'none' }}>
          <Form.Item name="reason" label="申请理由" rules={[{ required: true, max: 2000 }]}>
            <TextArea rows={4} placeholder="为什么想加入协会" />
          </Form.Item>
          <Form.Item name="skills" label="技能标签">
            <Select mode="tags" placeholder="输入后回车，如 Go / React" />
          </Form.Item>
          <Form.Item name="experience" label="项目经历">
            <TextArea rows={4} placeholder="过往项目、社团经历" />
          </Form.Item>
        </div>
      </Form>
      <div style={{ display: 'flex', justifyContent: 'space-between' }}>
        <Button disabled={step === 0} onClick={() => setStep((s) => s - 1)}>
          上一步
        </Button>
        {step < 2 ? (
          <Button type="primary" onClick={() => void next()}>
            下一步
          </Button>
        ) : (
          <Button type="primary" loading={submitting} onClick={() => void handleSubmit()}>
            提交申请
          </Button>
        )}
      </div>
      {step === 2 && (
        <div style={{ marginTop: 16 }}>
          <Tag color="blue">会员：资料审核后直接通过/拒绝</Tag>
          <Tag color="purple">干事：审核后进入面试流程</Tag>
        </div>
      )}
    </>
  );
};

export default ApplicationForm;
