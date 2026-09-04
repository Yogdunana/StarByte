import React, { useEffect, useState } from 'react';
import { AutoComplete, Form, InputNumber, Select } from 'antd';
import { getUserList } from '@/api/user';
import { getWorkflowRoleList } from '@/api/workflow';
import type { ApprovalConfig, ApprovalType, AssigneeStrategy } from '@/types/workflow';

interface OptionItem {
  label: string;
  value: string;
}

interface ApprovalConfigFormProps {
  value: ApprovalConfig;
  disabled?: boolean;
  onChange: (next: ApprovalConfig) => void;
}

const ApprovalConfigForm: React.FC<ApprovalConfigFormProps> = ({ value, disabled, onChange }) => {
  const [userOptions, setUserOptions] = useState<OptionItem[]>([]);
  const [roleOptions, setRoleOptions] = useState<OptionItem[]>([]);

  useEffect(() => {
    getWorkflowRoleList()
      .then((res) =>
        setRoleOptions((res.list ?? []).map((item) => ({ label: item.name, value: item.id }))),
      )
      .catch(() => setRoleOptions([]));
  }, []);

  const searchUsers = (keyword: string) => {
    getUserList({ page: 1, page_size: 20, keyword })
      .then((res) =>
        setUserOptions(
          res.list.map((user) => ({
            label: `${user.real_name || user.username} (${user.username})`,
            value: user.id,
          })),
        ),
      )
      .catch(() => setUserOptions([]));
  };

  const patch = (partial: Partial<ApprovalConfig>) => onChange({ ...value, ...partial });

  return (
    <>
      <Form.Item label="审批人类型">
        <Select
          disabled={disabled}
          value={value.assigneeStrategy}
          onChange={(assigneeStrategy: AssigneeStrategy) => patch({ assigneeStrategy })}
          options={[
            { label: '指定用户', value: 'static' },
            { label: '指定角色', value: 'role' },
            { label: '部门负责人', value: 'dept_leader' },
            { label: '发起人', value: 'initiator' },
          ]}
        />
      </Form.Item>
      {value.assigneeStrategy === 'static' && (
        <Form.Item label="审批人">
          <Select
            mode="multiple"
            disabled={disabled}
            value={value.assignees}
            options={userOptions}
            showSearch
            filterOption={false}
            onSearch={searchUsers}
            onDropdownVisibleChange={(open) => open && searchUsers('')}
            onChange={(assignees: string[]) => patch({ assignees })}
            placeholder="搜索并选择用户"
          />
        </Form.Item>
      )}
      {value.assigneeStrategy === 'role' && (
        <Form.Item label="角色">
          <AutoComplete
            disabled={disabled}
            value={value.roleId}
            options={roleOptions}
            onChange={(roleId: string) => patch({ roleId })}
            placeholder="选择角色"
          />
        </Form.Item>
      )}
      <Form.Item label="多人策略">
        <Select
          disabled={disabled}
          value={value.approvalType ?? 'single'}
          onChange={(approvalType: ApprovalType) => patch({ approvalType })}
          options={[
            { label: '单人', value: 'single' },
            { label: '会签（全部）', value: 'all' },
            { label: '或签（任一）', value: 'any' },
            { label: '比例通过', value: 'ratio' },
          ]}
        />
      </Form.Item>
      {value.approvalType === 'ratio' && (
        <Form.Item label="比例阈值">
          <InputNumber
            disabled={disabled}
            min={1}
            max={100}
            value={value.passRatio}
            onChange={(passRatio) => patch({ passRatio: passRatio ?? 0 })}
            addonAfter="%"
          />
        </Form.Item>
      )}
    </>
  );
};

export default ApprovalConfigForm;
