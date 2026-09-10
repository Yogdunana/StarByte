import { useCallback, useEffect, useRef, useState } from 'react';
import { Alert, Button, Descriptions, Input, Modal, Space, Spin, Steps, Timeline, Typography, message } from 'antd';
import dayjs from 'dayjs';
import { actTaskWorkflow, getTaskWorkflow, type TaskWorkflow } from '@/api/taskWorkflow';
import styles from './TaskWorkspace.module.css';
interface Props { taskId: string; onChanged: () => void }
const stages = ['assignment', 'execution', 'review', 'acceptance', 'completed'];
export const workflowStageNames: Record<string, string> = { assignment: '待分配', execution: '执行中', review: '待审核', acceptance: '待验收', completed: '已验收', cancelled: '已取消' };
const actionNames: Record<string, string> = { workflow_start: '开始处理', workflow_pause: '暂停处理', workflow_resume: '恢复处理', workflow_submit: '提交交付', workflow_approve: '正式签字', workflow_return: '退回返工' };
export default function WorkflowPanel({ taskId, onChanged }: Props) {
  const [flow, setFlow] = useState<TaskWorkflow | null>(null);
  const [error, setError] = useState(false);
  const [busy, setBusy] = useState(false);
  const [comment, setComment] = useState('');
  const [modal, holder] = Modal.useModal();
  const sequence = useRef(0);
  const invalidate = useCallback(() => { sequence.current++; }, []);
  const load = useCallback(async () => {
    const current = ++sequence.current;
    try { const result = await getTaskWorkflow(taskId); if (sequence.current === current) { setFlow(result); setError(false); } }
    catch { if (sequence.current === current) setError(true); }
  }, [taskId]);
  useEffect(() => { setFlow(null); setComment(''); void load(); return invalidate; }, [load, invalidate]);
  const move = async (action: 'start' | 'pause' | 'resume') => {
    if (!flow || busy) return; setBusy(true);
    try { const result = await actTaskWorkflow(taskId, action, ({ start: '执行人开始处理', pause: '执行人暂时暂停', resume: '执行人恢复处理' })[action], flow.revision); setFlow(result); onChanged(); } finally { setBusy(false); }
  };
  const act = (action: 'submit' | 'approve' | 'return') => {
    if (!flow || !comment.trim() || busy) return;
    modal.confirm({ title: action === 'submit' ? '提交这份交付说明？' : action === 'return' ? '退回给执行人返工？' : '确认以本人身份签字？', content: action === 'approve' ? '本次意见将记入审批历史。审核通过后还需完成正式验收。' : comment.trim(), okText: action === 'return' ? '退回返工' : '确认提交', cancelText: '继续检查', onOk: async () => {
      setBusy(true);
      try { const result = await actTaskWorkflow(taskId, action, comment.trim(), flow.revision); setFlow(result); setComment(''); message.success('任务流程已更新'); onChanged(); } finally { setBusy(false); }
    } });
  };
  return <section aria-label="任务审核验收">{holder}
    {error && <Alert type="error" showIcon message="审批进度暂时无法读取" description="仅指定参与人可查看和处理本任务审批。" action={<Button onClick={() => void load()}>重试</Button>} />}
    {!flow && !error && <Spin tip="正在读取审批进度"><div style={{ minHeight: 120 }} /></Spin>}
    {flow && <><Typography.Title level={4}>{flow.title}</Typography.Title><Button size="small" disabled={busy} onClick={() => void load()} style={{ marginBottom: 16 }}>刷新审批进度</Button><Steps size="small" direction="vertical" current={Math.max(0, stages.indexOf(flow.stage))} status={flow.stage === 'cancelled' ? 'error' : flow.stage === 'completed' ? 'finish' : 'process'} items={stages.map(stage => ({ title: workflowStageNames[stage], description: stage === 'assignment' ? flow.creator.name : stage === 'execution' ? flow.assignee?.name || '等待指定执行人' : stage === 'review' ? flow.reviewer.name : stage === 'acceptance' ? flow.acceptor.name : undefined }))} />
      <Space wrap style={{ marginBlock: 16 }}>{flow.can_start && <Button type="primary" loading={busy} onClick={() => void move('start')}>开始处理</Button>}{flow.can_pause && <Button disabled={busy} onClick={() => void move('pause')}>暂时挂起</Button>}{flow.can_resume && <Button type="primary" loading={busy} onClick={() => void move('resume')}>恢复处理</Button>}</Space>
      {flow.stage === 'cancelled' && <Alert type="warning" message="该任务流程已取消，历史记录仍保留" />}
      <Descriptions column={1} size="small" style={{ marginTop: 20 }}><Descriptions.Item label="分配方式">{({ department: "部门内按待办量分配", role: "部门内按角色及待办量分配", round_robin: "部门内轮流分配" } as Record<string, string>)[flow.assignment_mode] || "手动指定"}</Descriptions.Item><Descriptions.Item label="交付说明"><span style={{ whiteSpace: 'pre-wrap' }}>{flow.submission || '执行人尚未提交交付说明'}</span></Descriptions.Item></Descriptions>
      {(flow.can_submit || flow.can_approve) && <div className={styles.composer}><label htmlFor={`task-workflow-comment-${taskId}`}>{flow.can_submit ? '本次交付内容' : '审核 / 验收意见'}</label><Input.TextArea id={`task-workflow-comment-${taskId}`} value={comment} onChange={e => setComment(e.target.value)} rows={4} maxLength={5000} showCount disabled={busy} placeholder={flow.can_submit ? '写明完成情况、成果位置和需要确认的事项' : '写明检查结果；退回时说明需要修改的内容'} /><Space wrap style={{ marginTop: 16 }}>{flow.can_submit && <Button type="primary" disabled={!comment.trim()} loading={busy} onClick={() => act('submit')}>提交交付</Button>}{flow.can_approve && <Button type="primary" disabled={!comment.trim()} loading={busy} onClick={() => act('approve')}>{flow.stage === 'acceptance' ? '签字验收' : '签字通过审核'}</Button>}{flow.can_return && <Button danger disabled={!comment.trim() || busy} onClick={() => act('return')}>退回返工</Button>}</Space></div>}
      <Typography.Title level={5}>审批记录</Typography.Title>{!flow.history.length && <Typography.Text type="secondary">尚未提交交付或签字。</Typography.Text>}<Timeline items={flow.history.map(log => ({ children: <><strong>{log.operator.name} · {actionNames[log.action_type] || '流程更新'}</strong><p style={{ whiteSpace: 'pre-wrap' }}>{log.comment}</p><small>{dayjs(log.created_at).format('YYYY-MM-DD HH:mm')}</small></> }))} />
    </>}
  </section>;
}
