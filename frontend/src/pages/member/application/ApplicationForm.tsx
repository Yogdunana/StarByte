import React, { useEffect, useState } from 'react';
import { Button, Form, Input, Select, Steps, Tag, Modal, message } from 'antd';
import { getCurrentUser } from '@/api/auth';
import { useTranslation } from 'react-i18next';
import { getMemberDepartments, submitApplication } from '@/api/member';
import type { CreateMemberApplicationParams, MemberDepartmentOption } from '@/types/api';

const { TextArea } = Input;

interface ApplicationFormProps {
  onSubmitted?: () => void;
}

const ApplicationForm: React.FC<ApplicationFormProps> = ({ onSubmitted }) => {
  const { t } = useTranslation();
  const [identity, setIdentity] = useState<Partial<CreateMemberApplicationParams>>({});
  const [locked, setLocked] = useState({ real_name: false, student_no: false });
  const [form] = Form.useForm<CreateMemberApplicationParams>();
  const applicantType = Form.useWatch('applicant_type', form);
  const [step, setStep] = useState(0);
  const [submitting, setSubmitting] = useState(false);
  const [departments, setDepartments] = useState<MemberDepartmentOption[]>([]);

  useEffect(() => {
    getMemberDepartments()
      .then(setDepartments)
      .catch(() => undefined);
  }, []);

  useEffect(() => {
    let active = true;
    void getCurrentUser()
      .then((user) => {
        if (!active) return;
        const values: Partial<CreateMemberApplicationParams> = {};
        const locks = { real_name: false, student_no: false };
        for (const field of ['real_name', 'student_no'] as const) {
          if (user[field]?.trim() && !form.isFieldTouched(field)) {
            values[field] = user[field];
            locks[field] = true;
          }
        }
        form.setFieldsValue(values);
        setIdentity(values);
        setLocked(locks);
      })
      .catch(() => undefined);
    return () => {
      active = false;
    };
  }, [form]);

  const unlock = (field: 'real_name' | 'student_no') => {
    if (!locked[field]) return;
    Modal.confirm({
      title: t('application.confirmIdentityEdit', '是否确定修改'),
      content: t('application.identityEditHint', '此信息已自动获取，请确认修改后的信息准确。'),
      onOk: () => setLocked((current) => ({ ...current, [field]: false })),
    });
  };

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
      form.setFieldsValue(identity);
      setLocked({ real_name: !!identity.real_name, student_no: !!identity.student_no });
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
          <Form.Item
            name="real_name"
            label="姓名"
            rules={[{ required: true, whitespace: true, max: 50 }]}
          >
            <Input
              placeholder="真实姓名"
              readOnly={locked.real_name}
              style={
                locked.real_name
                  ? {
                      background: 'var(--ant-color-fill-tertiary, #f5f5f5)',
                      color: 'var(--ant-color-text-secondary, #666)',
                      cursor: 'pointer',
                    }
                  : undefined
              }
              onClick={() => unlock('real_name')}
              onKeyDown={(event) => {
                if (locked.real_name && (event.key === 'Enter' || event.key === ' ')) {
                  event.preventDefault();
                  unlock('real_name');
                }
              }}
            />
          </Form.Item>
          <Form.Item
            name="student_no"
            label="学号"
            rules={[{ required: true, whitespace: true, max: 30 }]}
          >
            <Input
              placeholder="学号"
              readOnly={locked.student_no}
              style={
                locked.student_no
                  ? {
                      background: 'var(--ant-color-fill-tertiary, #f5f5f5)',
                      color: 'var(--ant-color-text-secondary, #666)',
                      cursor: 'pointer',
                    }
                  : undefined
              }
              onClick={() => unlock('student_no')}
              onKeyDown={(event) => {
                if (locked.student_no && (event.key === 'Enter' || event.key === ' ')) {
                  event.preventDefault();
                  unlock('student_no');
                }
              }}
            />
          </Form.Item>
          <Form.Item
            name="department_id"
            label="意向部门"
            rules={[{ required: applicantType === 2, message: '干事申请请选择意向部门' }]}
          >
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
          <Form.Item
            name="reason"
            label="申请理由"
            rules={[{ required: true, whitespace: true, max: 2000 }]}
          >
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
          <Button
            type="primary"
            onClick={() => {
              void next().catch(() => undefined);
            }}
          >
            下一步
          </Button>
        ) : (
          <Button
            type="primary"
            loading={submitting}
            onClick={() => {
              void handleSubmit().catch(() => undefined);
            }}
          >
            提交申请
          </Button>
        )}
      </div>
      {step === 2 && (
        <div style={{ marginTop: 16 }}>
          <Tag color="blue">会员：资料审核后直接通过/拒绝</Tag>
          <Tag color="green">干事：面试 → 正式签字 → 候补期</Tag>
        </div>
      )}
    </>
  );
};

export default ApplicationForm;
