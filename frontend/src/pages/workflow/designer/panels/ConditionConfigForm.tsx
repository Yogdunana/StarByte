import React from 'react';
import { AutoComplete, Button, Checkbox, Form, Input, Select, Space } from 'antd';
import { MinusCircleOutlined, PlusOutlined } from '@ant-design/icons';
import type { ConditionBranch, ConditionConfig, ConditionOperator } from '@/types/workflow';
import { CONDITION_OPERATORS, CONDITION_VARIABLES } from '../constants';

interface ConditionConfigFormProps {
  value: ConditionConfig;
  nodeOptions: Array<{ label: string; value: string }>;
  edgeTargets: Record<string, string>;
  disabled?: boolean;
  onChange: (next: ConditionConfig) => void;
  onRetarget: (branchId: string, targetId: string) => void;
}

function parseExpression(expression: string): {
  variable: string;
  operator: ConditionOperator;
  value: string;
} {
  const matched = expression.match(/^(.+?)\s*(==|!=|>=|<=|>|<)\s*(.+)$/);
  if (!matched) {
    return { variable: expression, operator: '==', value: '' };
  }
  return {
    variable: matched[1].trim(),
    operator: matched[2] as ConditionOperator,
    value: matched[3].trim(),
  };
}

const ConditionConfigForm: React.FC<ConditionConfigFormProps> = ({
  value,
  nodeOptions,
  edgeTargets,
  disabled,
  onChange,
  onRetarget,
}) => {
  const updateBranch = (index: number, next: ConditionBranch) => {
    const branches = value.branches.map((branch, i) => (i === index ? next : branch));
    onChange({ branches });
  };

  const addBranch = () => {
    const id = `branch_${Date.now().toString(36)}`;
    onChange({
      branches: [
        ...value.branches,
        { id, label: `条件${value.branches.length + 1}`, expression: '', is_default: false },
      ],
    });
  };

  const removeBranch = (index: number) => {
    onChange({ branches: value.branches.filter((_, i) => i !== index) });
  };

  return (
    <>
      {value.branches.map((branch, index) => {
        const parsed = parseExpression(branch.expression);
        return (
          <div key={branch.id} style={{ marginBottom: 12, padding: 8, background: '#fafafa' }}>
            <Space style={{ width: '100%', justifyContent: 'space-between' }}>
              <span>分支 {index + 1}</span>
              {!disabled && (
                <MinusCircleOutlined onClick={() => removeBranch(index)} />
              )}
            </Space>
            <Form.Item label="变量">
              <AutoComplete
                disabled={disabled}
                value={parsed.variable}
                options={CONDITION_VARIABLES.map((item) => ({ value: item }))}
                onChange={(variable) =>
                  updateBranch(index, {
                    ...branch,
                    expression: `${variable} ${parsed.operator} ${parsed.value}`.trim(),
                    label: `${variable} ${parsed.operator} ${parsed.value}`.trim(),
                  })
                }
              />
            </Form.Item>
            <Form.Item label="运算符">
              <Select
                disabled={disabled}
                value={parsed.operator}
                options={CONDITION_OPERATORS.map((op) => ({ label: op, value: op }))}
                onChange={(operator: ConditionOperator) =>
                  updateBranch(index, {
                    ...branch,
                    expression: `${parsed.variable} ${operator} ${parsed.value}`.trim(),
                    label: `${parsed.variable} ${operator} ${parsed.value}`.trim(),
                  })
                }
              />
            </Form.Item>
            <Form.Item label="值">
              <Input
                disabled={disabled}
                value={parsed.value}
                onChange={(event) =>
                  updateBranch(index, {
                    ...branch,
                    expression: `${parsed.variable} ${parsed.operator} ${event.target.value}`.trim(),
                    label: `${parsed.variable} ${parsed.operator} ${event.target.value}`.trim(),
                  })
                }
              />
            </Form.Item>
            <Form.Item label="目标节点">
              <Select
                disabled={disabled}
                allowClear
                value={edgeTargets[branch.id]}
                options={nodeOptions}
                onChange={(targetId: string) => onRetarget(branch.id, targetId)}
              />
            </Form.Item>
            <Checkbox
              disabled={disabled}
              checked={branch.is_default}
              onChange={(event) =>
                updateBranch(index, { ...branch, is_default: event.target.checked })
              }
            >
              默认分支
            </Checkbox>
          </div>
        );
      })}
      {!disabled && (
        <Button type="dashed" block icon={<PlusOutlined />} onClick={addBranch}>
          添加条件
        </Button>
      )}
    </>
  );
};

export default ConditionConfigForm;
