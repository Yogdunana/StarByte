import TaskWorkflowPanel from '@/pages/task/WorkflowPanel';
import { useCallback, useEffect, useRef, useState } from 'react';
import { Alert, Button, Descriptions, Drawer, Form, Input, Select, Skeleton, Space, Tag, message } from 'antd';
import { useSelector } from 'react-redux';
import { selectCurrentUser } from '@/store/slices/userSlice';
import { decideWorkflowTask, getWorkflowCandidates, getWorkflowHistory, getWorkflowTask, rollbackWorkflowTask, transferWorkflowTask, type WorkflowHistory, type WorkflowPerson, type WorkflowTask } from '@/api/workflowRuntime';
import HistoryTimeline from './HistoryTimeline';
import AdmissionTask from './AdmissionTask';
import { dateLabel, taskLabels } from './meta';
import styles from './Runtime.module.css';
interface Props { id: string | null; onClose: () => void; onChanged: () => void }
interface Decision { action: 'approve' | 'reject' | 'transfer' | 'rollback'; comment: string; target_user_id?: string; target_node_id?: string }
export default function TaskDrawer({ id, onClose, onChanged }: Props) {
  const user = useSelector(selectCurrentUser);
  const [task, setTask] = useState<WorkflowTask | null>(null);
  const [history, setHistory] = useState<WorkflowHistory[]>([]);
  const [loading, setLoading] = useState(false);
  const [failed, setFailed] = useState(false);
  const [busy, setBusy] = useState(false);
  const [people, setPeople] = useState<WorkflowPerson[]>([]);
  const [peopleFailed, setPeopleFailed] = useState(false);
  const [keyword, setKeyword] = useState('');
  const sequence = useRef(0);
  const invalidate = useCallback(() => { sequence.current += 1; }, []);
  const [form] = Form.useForm<Decision>();
  const action = Form.useWatch('action', form);
  const load = useCallback(async () => {
    if (!id) return;
    const seq = ++sequence.current;
    setLoading(true); setFailed(false);
    try {
      const next = await getWorkflowTask(id);
      const records = await getWorkflowHistory(next.instance_id);
      if (sequence.current !== seq) return;
      setTask(next); setHistory(records);
    } catch { if (sequence.current === seq) setFailed(true); }
    finally { if (sequence.current === seq) setLoading(false); }
  }, [id]);
  useEffect(() => { setPeople([]); setKeyword(''); setTask(null); setHistory([]); setBusy(false); form.resetFields(); void load(); return invalidate; }, [form, load, invalidate]);
  useEffect(() => {
    if (!id || action !== 'transfer') return;
    let current = true;
    setPeopleFailed(false);
    const timer = window.setTimeout(() => {
      getWorkflowCandidates(id, keyword).then(result => { if (current) setPeople(result); }).catch(() => { if (current) { setPeople([]); setPeopleFailed(true); } });
    }, 250);
    return () => { current = false; window.clearTimeout(timer); };
  }, [id, keyword, action]);
  const submit = async (values: Decision) => {
    if (!task || busy) return;
    const seq = sequence.current;
    setBusy(true);
    try {
      if (values.action === 'transfer') await transferWorkflowTask(task.id, values.target_user_id!, values.comment);
      else if (values.action === 'rollback') await rollbackWorkflowTask(task.id, values.target_node_id!, values.comment);
      else await decideWorkflowTask(task.id, values.action, values.comment || '');
      if (sequence.current !== seq) return;
      message.success('操作已记录'); onChanged(); onClose();
    } catch { if (sequence.current === seq) await load(); }
    finally { setBusy(false); }
  };
  const targets = [...new Map(history.filter(item => item.action === 'approve' && item.node_id !== task?.node_id).map(item => [item.node_id, { value: item.node_id, label: item.node_name }])).values()];
  const canAct = task?.business_type !== 'member_application' && task?.business_type !== 'collaboration_task' && task?.status === 0 && task.instance_status === 0 && task.assignee_id === user?.id;
  return <Drawer open={!!id} title={task?.node_name || '审批待办'} onClose={onClose} width="min(600px, 100vw)" destroyOnClose>
    {loading ? <Skeleton active paragraph={{ rows: 8 }} /> : failed ? <Alert type="warning" showIcon message="任务或操作记录暂不可用" action={<Button onClick={() => void load()}>重试</Button>} /> : task && <>
      <Space wrap><Tag color={task.status === 0 ? 'gold' : 'green'}>{taskLabels[task.status]}</Tag><strong>{task.definition_name || '审批流程'}</strong></Space>
      <Descriptions size="small" column={1} style={{ marginTop: 20 }} items={[
        { key: 'assignee', label: '处理人', children: task.assignee_name || '待指定' },
        { key: 'created', label: '到达时间', children: dateLabel(task.created_at) },
        { key: 'due', label: '截止时间', children: dateLabel(task.due_date) },
        ...(task.completed_at ? [{ key: 'done', label: '处理时间', children: dateLabel(task.completed_at) }] : []),
      ]} />
      {task.instance_status === 3 && <Alert showIcon type="info" message="流程已挂起，恢复后才能继续处理" />}
      {task.business_type === 'collaboration_task' && <TaskWorkflowPanel taskId={task.business_key} onChanged={() => { onChanged(); onClose(); }} />}
      {task.business_type === 'member_application' && <AdmissionTask applicationId={task.business_key} instanceId={task.instance_id} onChanged={() => { onChanged(); onClose(); }} />}
      {canAct && <Form form={form} layout="vertical" initialValues={{ action: 'approve' }} onFinish={values => void submit(values)} className={styles.form}>
        <Form.Item name="action" label="处理方式" rules={[{ required: true }]}><Select options={[{ value: 'approve', label: '同意' }, { value: 'reject', label: '拒绝' }, { value: 'transfer', label: '转办给其他人' }, ...(targets.length ? [{ value: 'rollback', label: '退回已审批环节' }] : [])]} /></Form.Item>
        {action === 'transfer' && <><Form.Item name="target_user_id" label="接收人" rules={[{ required: true, message: '请选择接收人' }]}><Select showSearch filterOption={false} onSearch={setKeyword} options={people.map(person => ({ value: person.id, label: `${person.name}${person.department_name ? ` · ${person.department_name}` : ''}` }))} placeholder="搜索姓名或账号" /></Form.Item>{peopleFailed && <Alert type="warning" message="人员加载失败，请重新搜索" />}</>}
        {action === 'rollback' && <Form.Item name="target_node_id" label="退回环节" rules={[{ required: true }]}><Select options={targets} /></Form.Item>}
        <Form.Item name="comment" label={action === 'approve' ? '审批意见' : '具体原因'} rules={[{ required: action !== 'approve', whitespace: true, message: '请填写具体原因' }, { max: 1000 }]}><Input.TextArea rows={3} maxLength={1000} showCount /></Form.Item>
        <Button type="primary" htmlType="submit" loading={busy}>确认并记录</Button>
      </Form>}
      <h3 className={styles.sectionTitle}>流转记录</h3><HistoryTimeline items={history} />
    </>}
  </Drawer>;
}
