import { useCallback, useEffect, useRef, useState } from 'react';
import { Alert, Button, Descriptions, Drawer, Empty, Input, List, Modal, Space, Spin, Tabs, Tag, Timeline, Upload, message } from 'antd';
import { UploadOutlined } from '@ant-design/icons';
import { useSelector } from 'react-redux';
import dayjs from 'dayjs';
import StatusTag from '@/components/StatusTag/StatusTag';
import { selectCurrentUser } from '@/store/slices/userSlice';
import { addTaskComment, deleteTaskAttachment, deleteTaskComment, downloadTaskAttachment, getTaskAttachments, getTaskComments, getTaskDetail, getTaskLogs, updateTaskStatus, uploadTaskAttachment, urgeTask } from '@/api/task';
import type { Task, TaskAttachment, TaskComment, TaskLog } from '@/types/api';
import { TaskPriorityMap, TaskStatusMap } from './meta';
import { taskDate } from './TaskCollection';
import AssignmentModal from './AssignmentModal';
import WorkflowPanel from './WorkflowPanel';
import styles from './TaskWorkspace.module.css';
interface Props { taskId: string | null; open: boolean; onClose: () => void; onChanged: () => void }
const actionNames: Record<string, string> = { auto_assign: '按规则自动分配', create: '创建任务', assign: '分配负责人', transfer: '转办任务', status_change: '更新状态', comment: '添加评论', urge: '催办任务', delete: '删除任务', attachment_add: '添加附件', attachment_remove: '删除附件', attachment_remove_requested: '申请删除附件', workflow_delegate: '委托任务', workflow_handover: '申请转办', workflow_handover_approve: '转办签字通过', workflow_handover_reject: '转办被拒绝', workflow_escalate: '超时升级' };
const historyText = (log: TaskLog) => log.action_type === 'status_change' ? `${TaskStatusMap[Number(log.old_value)]?.text || '原状态'} → ${TaskStatusMap[Number(log.new_value)]?.text || '新状态'}` : log.action_type === 'create' ? log.new_value : '';
export default function DetailDrawer({ taskId, open, onClose, onChanged }: Props) {
  const me = useSelector(selectCurrentUser);
  const [modal, holder] = Modal.useModal();
  const [task, setTask] = useState<Task | null>(null);
  const [comments, setComments] = useState<TaskComment[]>([]);
  const [logs, setLogs] = useState<TaskLog[]>([]);
  const [files, setFiles] = useState<TaskAttachment[]>([]);
  const [content, setContent] = useState('');
  const [loading, setLoading] = useState(false);
  const [busy, setBusy] = useState(false);
  const [failed, setFailed] = useState(false);
  const [assignment, setAssignment] = useState<'assign' | 'transfer' | null>(null);
  const seq = useRef(0);
  const invalidate = useCallback(() => { seq.current++; }, []);
  const load = useCallback(async () => {
    if (!taskId) return;
    const current = ++seq.current; setLoading(true);
    try {
      const [t, c, l, a] = await Promise.all([getTaskDetail(taskId), getTaskComments(taskId), getTaskLogs(taskId), getTaskAttachments(taskId)]);
      if (seq.current !== current) return;
      setTask(t); setComments(c || []); setLogs(l || []); setFiles(a || []); setFailed(false);
    } catch { if (seq.current === current) setFailed(true); } finally { if (seq.current === current) setLoading(false); }
  }, [taskId]);
  useEffect(() => { setTask(null); setContent(''); setFailed(false); setAssignment(null); if (open) void load(); return invalidate; }, [open, load, invalidate]);
  const run = async (work: () => Promise<unknown>, success: string) => {
    if (busy) return; setBusy(true);
    try { await work(); message.success(success); await load(); onChanged(); } catch { /* API interceptor reports failure. */ } finally { setBusy(false); }
  };
  const changeStatus = (status: number) => {
    if (!task) return;
    let reason = '';
    const work = () => updateTaskStatus(task.id, status, reason);
    if ([2, 3].includes(status)) modal.confirm({ title: status === 2 ? '确认这项任务已经完成？' : '取消这项任务？', content: status === 3 ? <Input.TextArea placeholder="请填写取消原因" onChange={e => { reason = e.target.value; }} /> : '关闭后将保留历史记录，无法继续编辑或转办。', okText: status === 2 ? '确认完成' : '取消任务', cancelText: '继续处理', okButtonProps: { danger: status === 3 }, onOk: async () => { if (status === 3 && !reason.trim()) { message.warning('请填写取消原因'); throw new Error('reason required'); } await work(); message.success('状态已更新'); await load(); onChanged(); } });
    else void run(work, '状态已更新');
  };
  return <Drawer title="任务详情" width={720} open={open} onClose={() => { if (!busy) onClose(); }} styles={{ body: { padding: '20px clamp(16px, 4vw, 28px)' } }}>{holder}
    {loading && !task && <Spin tip="正在加载任务"><div style={{ minHeight: 180 }} /></Spin>}
    {failed && <Alert type="error" showIcon message="任务详情加载失败" description="任务可能已转办、删除，或暂时无法连接。" action={<Button onClick={() => void load()}>重试</Button>} style={{ marginBottom: 16 }} />}
    {task && <><div className={styles.detailHeading}><Space><StatusTag status={task.status} mapping={TaskStatusMap} /><StatusTag status={task.priority} mapping={TaskPriorityMap} /></Space><h2>{task.title}</h2><p>{task.creator.name} 创建于 {dayjs(task.created_at).format('YYYY-MM-DD HH:mm')}</p></div>
      <Descriptions column={{ xs: 1, sm: 2 }} size="small"><Descriptions.Item label="负责人">{task.assignee?.name || '待分配'}</Descriptions.Item><Descriptions.Item label="所属部门">{task.department?.name || '个人协作'}</Descriptions.Item><Descriptions.Item label="截止时间">{taskDate(task.due_date)}</Descriptions.Item><Descriptions.Item label="标签">{task.tags.length ? task.tags.map(tag => <Tag key={tag}>{tag}</Tag>) : '无标签'}</Descriptions.Item></Descriptions>
      <div className={styles.description}>{task.description || '暂未填写任务说明。'}</div>
      {task.children.length > 0 && <Alert type="info" message="关联子任务" description={task.children.map(child => <div key={child.id}>{child.title} · {TaskStatusMap[child.status]?.text}</div>)} />}
      <div className={styles.actions}>
        {task.can_update && (!task.workflow_stage || task.assignee?.id === me?.id) && task.status === 0 && <Button type="primary" disabled={busy} onClick={() => changeStatus(1)}>开始处理</Button>}
        {task.can_update && (!task.workflow_stage || task.assignee?.id === me?.id) && task.status === 1 && <>{!task.workflow_stage && <Button type="primary" disabled={busy} onClick={() => changeStatus(2)}>完成任务</Button>}<Button disabled={busy} onClick={() => changeStatus(4)}>暂时挂起</Button></>}
        {task.can_update && (!task.workflow_stage || task.assignee?.id === me?.id) && task.status === 4 && <Button type="primary" disabled={busy} onClick={() => changeStatus(1)}>恢复处理</Button>}
        {task.can_assign && <Button disabled={busy} onClick={() => setAssignment('assign')}>分配负责人</Button>}
        {task.can_transfer && <Button disabled={busy} onClick={() => setAssignment('transfer')}>转办</Button>}
        {task.can_urge && <Button disabled={busy} onClick={() => void run(() => urgeTask(task.id, '请查看任务进度并及时处理'), '催办已发送')}>提醒负责人</Button>}
        {task.can_cancel && <Button danger disabled={busy} onClick={() => changeStatus(3)}>取消任务</Button>}
      </div>
      <Tabs items={[
 ...(task.workflow_instance_id ? [{ key: 'workflow', label: '审核验收', children: <WorkflowPanel taskId={task.id} onChanged={() => { onChanged(); void load(); }} /> }] : []),
        { key: 'comments', label: `讨论 · ${comments.length}`, children: <>
          {!comments.length && <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="还没有讨论，写下进展或需要协助的事" />}
          {comments.map(item => <article className={styles.comment} key={item.id}><div className={styles.commentHead}><strong>{item.user?.name || '协作成员'}</strong><span>{dayjs(item.created_at).format('MM-DD HH:mm')}</span></div><p>{item.content}</p>{task.can_comment && item.user_id === me?.id && <Button size="small" type="text" danger disabled={busy} onClick={() => modal.confirm({ title: '删除这条评论？', okText: '删除评论', cancelText: '保留', onOk: async () => { await deleteTaskComment(task.id, item.id); await load(); onChanged(); } })}>删除评论</Button>}</article>)}
          {task.can_comment && <div className={styles.composer}><Input.TextArea rows={3} value={content} onChange={e => setContent(e.target.value)} placeholder="记录进展、问题或交接说明" /><div><small>@ 提及会发送通知；无权参与的人不会收到讨论内容。</small><Button type="primary" loading={busy} disabled={!content.trim()} onClick={() => void run(async () => { await addTaskComment(task.id, content.trim()); setContent(''); }, '评论已发布')}>发布评论</Button></div></div>}
        </> },
        { key: 'files', label: `附件 · ${files.length}`, children: <><List dataSource={files} locale={{ emptyText: '暂无附件' }} renderItem={item => <List.Item actions={[<Button key="download" type="text" onClick={() => void run(() => downloadTaskAttachment(task.id, item.id, item.file_name), '下载已开始')}>下载</Button>, task.can_update && <Button key="delete" type="text" danger onClick={() => modal.confirm({ title: '删除这个附件及存储文件？', content: '原文件也会删除；若存储暂不可用，系统会记录请求并重试。', okText: '删除附件', cancelText: '保留', onOk: async () => { await deleteTaskAttachment(task.id, item.id); await load(); onChanged(); } })}>删除</Button>]}><List.Item.Meta title={item.file_name} description={`${Math.max(1, Math.round(item.file_size / 1024))} KB · ${taskDate(item.created_at)}`} /></List.Item>} />{task.can_update && <Upload accept=".pdf,.doc,.docx,.xls,.xlsx,.ppt,.pptx" showUploadList={false} disabled={busy} beforeUpload={async file => { await run(() => uploadTaskAttachment(task.id, file), '附件已上传'); return false; }}><Button icon={<UploadOutlined />} loading={busy}>上传文档（最大 50 MB）</Button></Upload>}</> },
        { key: 'logs', label: '流转历史', children: <Timeline className={styles.history} items={logs.map(log => ({ children: <><strong>{log.operator?.name || '协作成员'} · {actionNames[log.action_type] || '更新任务'}</strong>{historyText(log) && <p>{historyText(log)}</p>}{log.comment && <p>{log.comment}</p>}<small>{dayjs(log.created_at).format('YYYY-MM-DD HH:mm')}</small></> }))} /> },
      ]} />
      <AssignmentModal task={task} mode={assignment} onClose={() => setAssignment(null)} onSaved={async () => { onChanged(); await load(); }} />
    </>}
  </Drawer>;
}
