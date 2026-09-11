import { useEffect, useState } from 'react';
import { Alert, Form, Select } from 'antd';
import { getTaskAssignmentRoles } from '@/api/task';
interface Props { mode?: string; busy: boolean }
export default function AssignmentRuleFields({ mode, busy }: Props) {
  const [roles, setRoles] = useState<Array<{ id: string; name: string }>>([]);
  const [keyword, setKeyword] = useState('');
  const [failed, setFailed] = useState(false);
  useEffect(() => {
    if (mode !== 'role' && mode !== 'round_robin') return;
    let active = true;
    const timer = window.setTimeout(() => {
      void getTaskAssignmentRoles(keyword).then(rows => { if (active) { setRoles(rows); setFailed(false); } }).catch(() => { if (active) setFailed(true); });
    }, 300);
    return () => { active = false; window.clearTimeout(timer); };
  }, [mode, keyword]);
  return <>
    <Form.Item name={['workflow', 'assignment', 'mode']} label="分配方式" initialValue="manual" extra="自动分配限任务所属部门，并排除审核人、验收人和停用账号。"><Select disabled={busy} options={[{ value: 'manual', label: '手动指定 / 稍后分配' }, { value: 'department', label: '部门内 · 优先分给待办较少的人' }, { value: 'role', label: '指定角色 · 优先分给待办较少的人' }, { value: 'round_robin', label: '轮流分配 · 按队列依次接单' }]} /></Form.Item>
    {(mode === 'role' || mode === 'round_robin') && <Form.Item name={['workflow', 'assignment', 'role_id']} label="分配角色" rules={[{ required: mode === 'role', message: '请选择分配角色' }]} extra={mode === 'round_robin' ? '可选。不限制时，在整个部门的可用成员中轮流分配。' : undefined}><Select disabled={busy} allowClear showSearch filterOption={false} onSearch={setKeyword} options={roles.map(r => ({ value: r.id, label: r.name }))} placeholder="搜索角色名称" /></Form.Item>}
    {failed && <Alert type="warning" showIcon message="角色加载失败，请重新输入名称搜索" />}
  </>;
}
