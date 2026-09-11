import { useEffect, useState } from 'react';
import { Alert, Button, Descriptions, Skeleton } from 'antd';
import { getApplicationDetail } from '@/api/member';
import type { MemberApplication } from '@/types/api';
import { usePermission } from '@/hooks/usePermission';
import AdmissionPanel from '@/pages/member/application/AdmissionPanel';

interface Props { applicationId: string; instanceId: string; onChanged: () => void }
export default function AdmissionTask({ applicationId, instanceId, onChanged }: Props) {
  const [application, setApplication] = useState<MemberApplication | null>(null);
  const [failed, setFailed] = useState(false);
  const [retry, setRetry] = useState(0);
  const canApprove = usePermission('member:approve');
  useEffect(() => {
    let active = true;
    setApplication(null); setFailed(false);
    getApplicationDetail(applicationId).then(row => { if (active) setApplication(row); }).catch(() => { if (active) setFailed(true); });
    return () => { active = false; };
  }, [applicationId, retry]);
  if (failed) return <Alert type="warning" showIcon message="申请详情暂不可用或无查看权限" action={<Button onClick={() => setRetry(value => value + 1)}>重试</Button>} />;
  if (!application) return <Skeleton active paragraph={{ rows: 3 }} />;
  return <>
    {application.flow_instance_id !== instanceId && <Alert type="info" showIcon message="这是此前提交的流程，下面显示申请的最新状态；请从最新待办处理。" />}
    <Descriptions size="small" column={1} items={[
      { key: 'name', label: '申请人', children: application.real_name },
      { key: 'type', label: '申请身份', children: application.applicant_type === 2 ? '干事' : '会员' },
      { key: 'department', label: '意向部门', children: application.department_name || '未选择' },
      { key: 'student', label: '学号', children: application.student_no },
      { key: 'phone', label: '联系电话', children: application.contact_phone || '—' },
      { key: 'email', label: '邮箱', children: application.contact_email || '—' },
      { key: 'reason', label: '申请理由', children: application.reason },
      { key: 'skills', label: '技能', children: application.skills.join('、') || '—' },
      { key: 'experience', label: '经历', children: application.experience || '—' },
    ]} />
    <AdmissionPanel id={applicationId} officer={application.applicant_type === 2} editable={canApprove && application.flow_instance_id === instanceId} onChanged={onChanged} />
  </>;
}
