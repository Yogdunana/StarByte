import { tx, useLocale } from '@/i18n/text';
import { useEffect, useState } from 'react';
import { Alert, Button, Descriptions, Skeleton } from 'antd';
import { getApplicationDetail } from '@/api/member';
import type { MemberApplication } from '@/types/api';
import { usePermission } from '@/hooks/usePermission';
import AdmissionPanel from '@/pages/member/application/AdmissionPanel';
import EngineChainPanel from '@/pages/member/application/EngineChainPanel';

interface Props {
  applicationId: string;
  instanceId: string;
  onChanged: () => void;
}
export default function AdmissionTask({ applicationId, instanceId, onChanged }: Props) {
  useLocale();
  const [application, setApplication] = useState<MemberApplication | null>(null);
  const [failed, setFailed] = useState(false);
  const [retry, setRetry] = useState(0);
  const canApprove = usePermission('member:approve');
  useEffect(() => {
    let active = true;
    setApplication(null);
    setFailed(false);
    getApplicationDetail(applicationId)
      .then((row) => {
        if (active) setApplication(row);
      })
      .catch(() => {
        if (active) setFailed(true);
      });
    return () => {
      active = false;
    };
  }, [applicationId, retry]);
  if (failed)
    return (
      <Alert
        type="warning"
        showIcon
        message={tx('申请详情暂不可用或无查看权限')}
        action={<Button onClick={() => setRetry((value) => value + 1)}>{tx('重试')}</Button>}
      />
    );
  if (!application) return <Skeleton active paragraph={{ rows: 3 }} />;
  return (
    <>
      {application.flow_instance_id !== instanceId && (
        <Alert
          type="info"
          showIcon
          message={tx('这是此前提交的流程，下面显示申请的最新状态；请从最新待办处理。')}
        />
      )}
      <Descriptions
        size="small"
        column={1}
        items={[
          { key: 'name', label: tx('申请人'), children: application.real_name },
          {
            key: 'type',
            label: tx('申请身份'),
            children: application.applicant_type === 2 ? tx('干事') : tx('会员'),
          },
          {
            key: 'department',
            label: tx('意向部门'),
            children: application.department_name || tx('未选择'),
          },
          { key: 'student', label: tx('学号'), children: application.student_no },
          { key: 'phone', label: tx('联系电话'), children: application.contact_phone || '—' },
          { key: 'email', label: tx('邮箱'), children: application.contact_email || '—' },
          { key: 'reason', label: tx('申请理由'), children: application.reason },
          { key: 'skills', label: tx('技能'), children: application.skills.join('、') || '—' },
          { key: 'experience', label: tx('经历'), children: application.experience || '—' },
        ]}
      />
      {application.workflow_key === 'member_application' ? (
        <EngineChainPanel
          record={application}
          editable={canApprove && application.flow_instance_id === instanceId}
          onChanged={onChanged}
        />
      ) : (
        <AdmissionPanel
          id={applicationId}
          officer={application.applicant_type === 2}
          editable={canApprove && application.flow_instance_id === instanceId}
          onChanged={onChanged}
        />
      )}
    </>
  );
}
