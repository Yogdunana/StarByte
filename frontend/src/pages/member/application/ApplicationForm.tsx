import { tx, useLocale } from '@/i18n/text';
import React, { useEffect, useState } from 'react';
import { Button, Form, Input, Select, Steps, Tag, Modal, message } from 'antd';
import { getCurrentUser } from '@/api/auth';
import { useTranslation } from 'react-i18next';
import { getMemberDepartments, submitApplication } from '@/api/member';
import type { CreateMemberApplicationParams, MemberDepartmentOption } from '@/types/api';
import { isCnMobile, nationalMobileDigits } from '@/utils/phone';

interface ApplicationFormProps {
  onSubmitted?: () => void;
}

type LockedField = 'real_name' | 'student_no' | 'gender';

const ApplicationForm: React.FC<ApplicationFormProps> = ({ onSubmitted }) => {
  useLocale();
  const { t } = useTranslation();
  const [identity, setIdentity] = useState<Partial<CreateMemberApplicationParams>>({});
  const [locked, setLocked] = useState({ real_name: false, student_no: false, gender: false });
  const [canApplyOfficer, setCanApplyOfficer] = useState(false);
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
        const locks = { real_name: false, student_no: false, gender: false };
        for (const field of ['real_name', 'student_no'] as const) {
          if (user[field]?.trim() && !form.isFieldTouched(field)) {
            values[field] = user[field];
            locks[field] = true;
          }
        }
        if ((user.gender === 1 || user.gender === 2) && !form.isFieldTouched('gender')) {
          values.gender = user.gender;
          locks.gender = true;
        }
        const phone = nationalMobileDigits(user.phone);
        if (isCnMobile(phone) && !form.isFieldTouched('contact_phone')) {
          values.contact_phone = phone;
        }
        if (user.email?.trim() && !form.isFieldTouched('contact_email')) {
          values.contact_email = user.email.trim();
        }
        const member = (user.roles || []).includes('member');
        setCanApplyOfficer(member);
        if (!member) {
          values.applicant_type = 1;
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

  useEffect(() => {
    if (applicantType !== 2) {
      form.setFieldValue('department_id', undefined);
    }
  }, [applicantType, form]);

  const unlock = (field: LockedField) => {
    if (!locked[field]) return;
    Modal.confirm({
      title: t('application.confirmIdentityEdit', '是否确定修改'),
      content: t('application.identityEditHint', '该信息是从系统自动获取的，是否要更改？'),
      onOk: () => setLocked((current) => ({ ...current, [field]: false })),
    });
  };

  const next = async () => {
    const fields: Array<keyof CreateMemberApplicationParams> =
      applicantType === 2
        ? ['applicant_type', 'real_name', 'student_no', 'gender', 'department_id']
        : ['applicant_type', 'real_name', 'student_no', 'gender'];
    await form.validateFields([...fields]);
    setStep(1);
  };

  const handleSubmit = async () => {
    const values = await form.validateFields();
    setSubmitting(true);
    try {
      await submitApplication({
        ...values,
        applicant_type: canApplyOfficer ? values.applicant_type : 1,
        department_id: values.applicant_type === 2 && canApplyOfficer ? values.department_id : undefined,
        contact_phone: nationalMobileDigits(values.contact_phone),
      });
      message.success(tx('申请已提交'));
      form.resetFields();
      form.setFieldsValue(identity);
      setLocked({
        real_name: !!identity.real_name,
        student_no: !!identity.student_no,
        gender: identity.gender === 1 || identity.gender === 2,
      });
      setStep(0);
      onSubmitted?.();
    } finally {
      setSubmitting(false);
    }
  };

  const lockedInputStyle = {
    background: 'var(--ant-color-fill-tertiary, #f5f5f5)',
    color: 'var(--ant-color-text-secondary, #666)',
    cursor: 'pointer',
  } as const;

  return (
    <>
      <Steps
        current={step}
        style={{ marginBottom: 24 }}
        items={[{ title: tx('基本信息') }, { title: tx('联系方式') }]}
      />
      <Form form={form} layout="vertical" initialValues={{ applicant_type: 1 }}>
        <div style={{ display: step === 0 ? 'block' : 'none' }}>
          <Form.Item name="applicant_type" label={tx('申请类型')} rules={[{ required: true }]}>
            <Select
              options={[
                { value: 1, label: tx('会员') },
                { value: 2, label: tx('干事（需面试）'), disabled: !canApplyOfficer },
              ]}
            />
          </Form.Item>
          {!canApplyOfficer ? (
            <p style={{ margin: '0 0 16px', color: 'var(--sb-muted)', fontSize: 13 }}>
              {tx('须先成为会员后再申请干事或干部职务')}
            </p>
          ) : null}
          <Form.Item
            name="real_name"
            label={tx('姓名')}
            rules={[{ required: true, whitespace: true, max: 50 }]}
          >
            <Input
              placeholder={tx('真实姓名')}
              readOnly={locked.real_name}
              style={locked.real_name ? lockedInputStyle : undefined}
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
            label={tx('学号')}
            rules={[{ required: true, whitespace: true, max: 30 }]}
          >
            <Input
              placeholder={tx('学号')}
              readOnly={locked.student_no}
              style={locked.student_no ? lockedInputStyle : undefined}
              onClick={() => unlock('student_no')}
              onKeyDown={(event) => {
                if (locked.student_no && (event.key === 'Enter' || event.key === ' ')) {
                  event.preventDefault();
                  unlock('student_no');
                }
              }}
            />
          </Form.Item>
          <Form.Item name="gender" label={tx('性别')} rules={[{ required: true, message: tx('请选择性别') }]}>
            <Select
              placeholder={tx('请选择性别')}
              open={locked.gender ? false : undefined}
              onClick={() => unlock('gender')}
              options={[
                { value: 1, label: tx('男') },
                { value: 2, label: tx('女') },
              ]}
            />
          </Form.Item>
          {applicantType === 2 && canApplyOfficer ? (
            <Form.Item
              name="department_id"
              label={tx('意向部门')}
              rules={[{ required: true, message: tx('干事申请请选择意向部门') }]}
            >
              <Select
                allowClear
                placeholder={tx('选择部门')}
                options={departments.map((d) => ({ value: d.id, label: d.name }))}
              />
            </Form.Item>
          ) : (
            <p style={{ margin: '0 0 16px', color: 'var(--sb-muted)', fontSize: 13 }}>
              {tx('会员不隶属任何部门，无需选择意向部门')}
            </p>
          )}
        </div>
        <div style={{ display: step === 1 ? 'block' : 'none' }}>
          <Form.Item
            name="contact_phone"
            label={tx('手机号')}
            getValueFromEvent={(event: { target: { value: string } }) =>
              nationalMobileDigits(event.target.value).slice(0, 11)
            }
            rules={[
              { required: true, message: tx('请输入手机号') },
              {
                validator: async (_, value) => {
                  if (!isCnMobile(value)) {
                    throw new Error(tx('请输入 11 位中国大陆手机号'));
                  }
                },
              },
            ]}
          >
            <Input prefix="+86" placeholder={tx('11 位手机号')} maxLength={13} inputMode="numeric" />
          </Form.Item>
          <Form.Item
            name="contact_email"
            label={tx('邮箱')}
            rules={[{ required: true, type: 'email', max: 100 }]}
          >
            <Input placeholder={tx('联系邮箱')} />
          </Form.Item>
        </div>
      </Form>
      <div style={{ display: 'flex', justifyContent: 'space-between' }}>
        <Button disabled={step === 0} onClick={() => setStep(0)}>
          {tx('上一步')}
        </Button>
        {step < 1 ? (
          <Button
            type="primary"
            onClick={() => {
              void next().catch(() => undefined);
            }}
          >
            {tx('下一步')}
          </Button>
        ) : (
          <Button
            type="primary"
            loading={submitting}
            onClick={() => {
              void handleSubmit().catch(() => undefined);
            }}
          >
            {tx('提交申请')}
          </Button>
        )}
      </div>
      {step === 1 && (
        <div style={{ marginTop: 16 }}>
          <Tag color="blue">{tx('会员：资料审核后即为会员')}</Tag>
          <Tag color="green">{tx('干事须先成为会员，再另行申请')}</Tag>
        </div>
      )}
    </>
  );
};

export default ApplicationForm;
