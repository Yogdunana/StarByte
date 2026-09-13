import { tx, useLocale } from '@/i18n/text';
import { useCallback, useEffect, useRef, useState } from 'react';
import {
  Alert,
  Button,
  Form,
  Input,
  Select,
  Skeleton,
  Space,
  Steps,
  Tag,
  Timeline,
  message,
} from 'antd';
import dayjs from 'dayjs';
import {
  getAdmission,
  signAdmission,
  handleAdmissionObjection,
  type AdmissionState,
  type SignAdmissionParams,
} from '@/api/member';
import { requiredFieldOptions } from '../meta';
import styles from './AdmissionPanel.module.css';

const stages: Record<string, string> = {
  get materials() {
    return tx('资料审核');
  },
  get round1() {
    return tx('一面与会签');
  },
  get round2() {
    return tx('二面与中心审批');
  },
  get president() {
    return tx('会长最终确认');
  },
  get probation() {
    return tx('候补期');
  },
  get approved() {
    return tx('正式成员');
  },
  get rejected() {
    return tx('未通过');
  },
  get supplement() {
    return tx('补充材料');
  },
  get legacy_review() {
    return tx('历史待核验');
  },
};
const roles: Record<string, string> = {
  get materials() {
    return tx('资料审核人');
  },
  get minister() {
    return tx('部长');
  },
  get center() {
    return tx('中心负责人');
  },
  get president() {
    return tx('会长');
  },
};
const decisions: Record<string, string> = {
  get approve() {
    return tx('同意');
  },
  get reject() {
    return tx('拒绝');
  },
  get supplement() {
    return tx('要求补充材料');
  },
};
interface Props {
  id: string;
  officer: boolean;
  editable: boolean;
  onChanged: () => void;
}

export default function AdmissionPanel({ id, officer, editable, onChanged }: Props) {
  useLocale();
  const [state, setState] = useState<AdmissionState | null>(null);
  const loadSequence = useRef(0);
  const invalidate = useCallback(() => {
    loadSequence.current += 1;
  }, []);
  const [failed, setFailed] = useState(false);
  const [objectionForm] = Form.useForm<{ action: string; comment: string }>();
  const [busy, setBusy] = useState(false);
  const [form] = Form.useForm<SignAdmissionParams>();
  const decision = Form.useWatch('decision', form);
  const load = useCallback(async () => {
    const sequence = ++loadSequence.current;
    setFailed(false);
    try {
      const result = await getAdmission(id);
      if (sequence !== loadSequence.current) return;
      setState(result);
      form.setFieldsValue({ role: result.allowed_roles[0], decision: 'approve' });
    } catch {
      if (sequence === loadSequence.current) {
        setFailed(true);
        setState(null);
      }
    }
  }, [id, form]);
  useEffect(() => {
    setState(null);
    void load();
    return invalidate;
  }, [load, invalidate]);
  const submit = async (values: SignAdmissionParams) => {
    if (!state) return;
    setBusy(true);
    try {
      const result = await signAdmission(id, {
        ...values,
        stage: state.stage,
        revision: state.revision,
      });
      setState(result);
      form.resetFields();
      message.success(tx('签字已记录'));
      onChanged();
    } catch {
      await load();
    } finally {
      setBusy(false);
    }
  };
  const submitObjection = async (values: { action: string; comment: string }) => {
    setBusy(true);
    try {
      await handleAdmissionObjection(id, values.action, values.comment);
      objectionForm.resetFields();
      message.success(tx('异议处理已记录'));
      await load();
      onChanged();
    } finally {
      setBusy(false);
    }
  };
  if (failed)
    return (
      <Alert
        className={styles.panel}
        type="warning"
        showIcon
        message={tx('审批记录暂不可用或无查看权限')}
        action={<Button onClick={() => void load()}>{tx('重试')}</Button>}
      />
    );
  if (!state) return <Skeleton active paragraph={{ rows: 3 }} />;
  const path = officer
    ? ['materials', 'round1', 'round2', 'president', 'probation', 'approved']
    : state.stage === 'probation'
      ? ['probation', 'approved']
      : ['materials', 'approved'];
  return (
    <section className={styles.panel} aria-label={tx('正式审批')}>
      <div className={styles.heading}>
        <h3>{tx('正式审批')}</h3>
        <Tag color={state.historical_review_required ? 'orange' : 'green'}>
          {stages[state.stage] || state.stage}
        </Tag>
      </div>
      {state.historical_review_required ? (
        <Alert
          type="warning"
          showIcon
          message={tx('历史记录待核验')}
          description={tx('保留原有状态；缺少的签字不会自动补写。')}
        />
      ) : (
        <>
          <Steps
            direction="vertical"
            size="small"
            current={path.indexOf(state.stage)}
            items={path.map((stage) => ({ title: stages[stage] }))}
          />
          {!state.interview_completed && (
            <Alert
              type="info"
              showIcon
              message={tx('等待本轮面试完成')}
              description={tx('面试结束后才开放签字；分数不会自动替代审批。')}
            />
          )}
          {editable && state.allowed_roles.length > 0 && (
            <Form
              form={form}
              layout="vertical"
              onFinish={(values) => void submit(values)}
              className={styles.form}
            >
              <Form.Item name="role" label={tx('本次签字身份')} rules={[{ required: true }]}>
                <Select
                  options={state.allowed_roles.map((role) => ({
                    value: role,
                    label: roles[role] || role,
                  }))}
                />
              </Form.Item>
              <Form.Item name="decision" label={tx('审批决定')} rules={[{ required: true }]}>
                <Select
                  options={[
                    { value: 'approve', label: tx('同意') },
                    ...(state.stage === 'probation'
                      ? []
                      : [{ value: 'reject', label: tx('拒绝') }]),
                    ...(state.stage === 'materials'
                      ? [{ value: 'supplement', label: tx('要求补充材料') }]
                      : []),
                  ]}
                />
              </Form.Item>
              <Form.Item
                name="comment"
                label={tx('对申请人公开的审批意见')}
                rules={[
                  { required: decision !== 'approve', message: tx('请填写原因') },
                  { max: 1000 },
                ]}
              >
                <Input.TextArea rows={3} placeholder={tx('请勿在这里填写内部评分或保密评语')} />
              </Form.Item>
              {decision === 'supplement' && (
                <Form.Item name="required_fields" label={tx('需补充字段')}>
                  <Select mode="multiple" options={requiredFieldOptions} />
                </Form.Item>
              )}
              <Form.Item name="delegation_reason" label={tx('上级代签原因（本人职责签字可留空）')}>
                <Input.TextArea
                  rows={2}
                  maxLength={1000}
                  placeholder={tx('代签仅在环节超时24小时后开放')}
                />
              </Form.Item>
              <Button htmlType="submit" type="primary" loading={busy}>
                {tx('确认并记录签字')}
              </Button>
            </Form>
          )}
          {editable && !state.allowed_roles.length && state.interview_completed && (
            <p className={styles.hint}>{tx('当前没有需要你签字的环节。')}</p>
          )}
        </>
      )}
      {!!state.objections?.length && (
        <>
          <h4>{tx('候补期异议')}</h4>
          {state.objections.map((item) => (
            <Alert
              key={item.id}
              type={item.status === 'dismiss' ? 'success' : 'warning'}
              showIcon
              message={
                {
                  center_review: tx('等待中心复核'),
                  president_review: tx('等待会长裁决'),
                  uphold: tx('异议成立，终止录用'),
                  dismiss: tx('异议不成立，继续候补'),
                }[item.status] || item.status
              }
              description={
                item.reason || item.center_comment || item.final_comment ? (
                  <>
                    <p>{item.reason}</p>
                    {item.center_comment && (
                      <p>
                        {tx('中心复核：')}
                        {item.center_comment}
                      </p>
                    )}
                    {item.final_comment && (
                      <p>
                        {tx('会长意见：')}
                        {item.final_comment}
                      </p>
                    )}
                  </>
                ) : undefined
              }
            />
          ))}
        </>
      )}
      {editable && !!state.allowed_objection_actions?.length && (
        <Form
          form={objectionForm}
          layout="vertical"
          className={styles.form}
          onFinish={(values) => {
            void submitObjection(values).catch(() => undefined);
          }}
        >
          <h4>{tx('候补期异议处理')}</h4>
          <Form.Item name="action" label={tx('处理动作')} rules={[{ required: true }]}>
            <Select
              options={state.allowed_objection_actions.map((action) => ({
                value: action,
                label:
                  {
                    raise: tx('提出异议'),
                    center_review: tx('提交中心复核意见'),
                    uphold: tx('异议成立，终止录用'),
                    dismiss: tx('异议不成立，继续候补'),
                  }[action] || action,
              }))}
            />
          </Form.Item>
          <Form.Item
            name="comment"
            label={tx('具体理由')}
            rules={[{ required: true, whitespace: true }, { max: 2000 }]}
          >
            <Input.TextArea rows={3} />
          </Form.Item>
          <Button type="primary" htmlType="submit" loading={busy}>
            {tx('确认并记录')}
          </Button>
        </Form>
      )}
      <h4>{tx('签字记录')}</h4>
      {state.signatures.length ? (
        <Timeline
          items={state.signatures.map((item) => ({
            color: item.decision === 'reject' ? 'red' : 'green',
            children: (
              <>
                <Space wrap>
                  <strong>
                    {stages[item.stage] || item.stage} ·{' '}
                    {roles[item.signer_role] || item.signer_role}
                  </strong>
                  <Tag>{decisions[item.decision]}</Tag>
                </Space>
                <p>
                  {item.signer_name || item.signer_id.slice(0, 8)} ·{' '}
                  {dayjs(item.created_at).format('YYYY-MM-DD HH:mm')}
                </p>
                {item.comment && <p>{item.comment}</p>}
                {item.delegated && (
                  <p>
                    {tx('上级代签：')}
                    {item.delegation_reason}
                  </p>
                )}
              </>
            ),
          }))}
        />
      ) : (
        <p className={styles.hint}>{tx('尚无正式签字记录。')}</p>
      )}
    </section>
  );
}
