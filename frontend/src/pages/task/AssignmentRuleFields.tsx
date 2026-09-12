import { useEffect, useState } from 'react';
import { Alert, Form, Select } from 'antd';
import { useTranslation } from 'react-i18next';
import { getTaskAssignmentRoles } from '@/api/task';
interface Props { mode?: string; busy: boolean }
export default function AssignmentRuleFields({ mode, busy }: Props) {
  const { t } = useTranslation();
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
    <Form.Item name={['workflow', 'assignment', 'mode']} label={t('task.form.assignmentMode')} initialValue="manual" extra={t('task.form.assignmentExtra')}>
      <Select disabled={busy} options={[
        { value: 'manual', label: t('task.form.manual') },
        { value: 'department', label: t('task.form.department') },
        { value: 'role', label: t('task.form.role') },
        { value: 'round_robin', label: t('task.form.roundRobin') },
      ]} />
    </Form.Item>
    {(mode === 'role' || mode === 'round_robin') && (
      <Form.Item name={['workflow', 'assignment', 'role_id']} label={t('task.form.roleLabel')} rules={[{ required: mode === 'role', message: t('task.form.roleRequired') }]} extra={mode === 'round_robin' ? t('task.form.roundRobinExtra') : undefined}>
        <Select disabled={busy} allowClear showSearch filterOption={false} onSearch={setKeyword} options={roles.map(r => ({ value: r.id, label: r.name }))} placeholder={t('task.form.roleSearch')} />
      </Form.Item>
    )}
    {failed && <Alert type="warning" showIcon message={t('task.form.roleFailed')} />}
  </>;
}
