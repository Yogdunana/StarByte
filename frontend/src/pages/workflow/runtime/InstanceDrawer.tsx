import { tx, useLocale } from '@/i18n/text';
import TaskWorkflowPanel from '@/pages/task/WorkflowPanel';
import { useCallback, useEffect, useRef, useState } from 'react';
import {
  Alert,
  Button,
  Descriptions,
  Drawer,
  Form,
  Input,
  Select,
  Skeleton,
  Tag,
  message,
} from 'antd';
import { useSelector } from 'react-redux';
import { selectCurrentUser } from '@/store/slices/userSlice';
import { usePermission } from '@/hooks/usePermission';
import {
  changeWorkflowInstance,
  getWorkflowHistory,
  getWorkflowInstance,
  type WorkflowHistory,
  type WorkflowInstance,
} from '@/api/workflowRuntime';
import HistoryTimeline from './HistoryTimeline';
import AdmissionTask from './AdmissionTask';
import HandoverPanel from '@/pages/task/HandoverPanel';
import { dateLabel, instanceLabels } from './meta';
import styles from './Runtime.module.css';
interface Props {
  id: string | null;
  onClose: () => void;
  onChanged: () => void;
}
interface Operation {
  action: 'suspend' | 'resume' | 'terminate';
  reason: string;
}
export default function InstanceDrawer({ id, onClose, onChanged }: Props) {
  useLocale();
  const [instance, setInstance] = useState<WorkflowInstance | null>(null);
  const [history, setHistory] = useState<WorkflowHistory[]>([]);
  const [failed, setFailed] = useState(false);
  const [busy, setBusy] = useState(false);
  const [loading, setLoading] = useState(true);
  const sequence = useRef(0);
  const invalidate = useCallback(() => {
    sequence.current += 1;
  }, []);
  const [form] = Form.useForm<Operation>();
  const action = Form.useWatch('action', form);
  const user = useSelector(selectCurrentUser);
  const canUpdate = usePermission('workflow:update');
  const load = useCallback(async () => {
    if (!id) return;
    const seq = ++sequence.current;
    setLoading(true);
    setFailed(false);
    try {
      const [detail, records] = await Promise.all([
        getWorkflowInstance(id),
        getWorkflowHistory(id),
      ]);
      if (sequence.current !== seq) return;
      setInstance(detail);
      setHistory(records);
    } catch {
      if (sequence.current === seq) setFailed(true);
    } finally {
      if (sequence.current === seq) setLoading(false);
    }
  }, [id]);
  useEffect(() => {
    setInstance(null);
    form.resetFields();
    void load();
    return invalidate;
  }, [form, load, invalidate]);
  const submit = async (values: Operation) => {
    if (!id || busy) return;
    const seq = sequence.current;
    setBusy(true);
    try {
      await changeWorkflowInstance(id, values.action, values.reason || '');
      if (sequence.current !== seq) return;
      message.success(tx('流程状态已更新'));
      onChanged();
      await load();
    } catch {
      /* API layer shows the concrete reason, including scope restrictions. */
    } finally {
      setBusy(false);
    }
  };
  return (
    <Drawer
      open={!!id}
      onClose={onClose}
      title={instance?.definition_name || tx('流程进度')}
      width="min(600px, 100vw)"
      destroyOnClose
    >
      {loading ? (
        <Skeleton active paragraph={{ rows: 8 }} />
      ) : failed ? (
        <Alert
          type="warning"
          showIcon
          message={tx('流程或操作记录暂不可用')}
          action={<Button onClick={() => void load()}>{tx('重试')}</Button>}
        />
      ) : (
        instance && (
          <>
            <Tag
              color={
                instance.status === 0 ? 'processing' : instance.status === 1 ? 'success' : 'default'
              }
            >
              {instanceLabels[instance.status]}
            </Tag>
            <Descriptions
              column={1}
              size="small"
              style={{ marginTop: 20 }}
              items={[
                { key: 'initiator', label: tx('发起人'), children: instance.initiator_name || '—' },
                { key: 'started', label: tx('发起时间'), children: dateLabel(instance.started_at) },
                { key: 'ended', label: tx('结束时间'), children: dateLabel(instance.ended_at) },
                {
                  key: 'current',
                  label: tx('当前待处理环节'),
                  children: instance.current_node_ids.length
                    ? tx('{{value0}} 个', { value0: instance.current_node_ids.length })
                    : tx('无'),
                },
              ]}
            />
            {instance.terminate_reason && (
              <Alert type="info" showIcon message={instance.terminate_reason} />
            )}
            {instance.business_type === 'collaboration_task' && (
              <TaskWorkflowPanel
                taskId={instance.business_key}
                onChanged={() => {
                  onChanged();
                  void load();
                }}
              />
            )}
            {instance.business_type === 'task_transfer' && (
              <HandoverPanel
                transferId={instance.business_key}
                onChanged={() => {
                  onChanged();
                  void load();
                }}
              />
            )}
            {instance.business_type === 'member_application' && (
              <AdmissionTask
                applicationId={instance.business_key}
                instanceId={instance.id}
                onChanged={() => {
                  onChanged();
                  void load();
                }}
              />
            )}
            {instance.business_type !== 'member_application' &&
              instance.business_type !== 'collaboration_task' &&
              instance.business_type !== 'task_transfer' &&
              [0, 3].includes(instance.status) &&
              (canUpdate || instance.initiator_id === user?.id) && (
                <Form
                  form={form}
                  layout="vertical"
                  onFinish={(values) => void submit(values)}
                  className={styles.form}
                >
                  <Form.Item name="action" label={tx('流程管理')} rules={[{ required: true }]}>
                    <Select
                      options={[
                        {
                          value: instance.status === 3 ? 'resume' : 'suspend',
                          label: instance.status === 3 ? tx('恢复流转') : tx('暂时挂起'),
                        },
                        { value: 'terminate', label: tx('终止流程') },
                      ]}
                    />
                  </Form.Item>
                  {action === 'terminate' && (
                    <Alert
                      type="warning"
                      showIcon
                      message={tx('终止后将关闭所有未处理待办，已有操作记录会保留。')}
                    />
                  )}
                  {action !== 'resume' && (
                    <Form.Item
                      name="reason"
                      label={tx('具体原因')}
                      rules={[{ required: true, whitespace: true }, { max: 500 }]}
                    >
                      <Input.TextArea rows={3} />
                    </Form.Item>
                  )}
                  <Button
                    htmlType="submit"
                    type="primary"
                    danger={action === 'terminate'}
                    loading={busy}
                  >
                    {tx('确认操作')}
                  </Button>
                </Form>
              )}
            <h3 className={styles.sectionTitle}>{tx('操作记录')}</h3>
            <HistoryTimeline items={history} />
          </>
        )
      )}
    </Drawer>
  );
}
