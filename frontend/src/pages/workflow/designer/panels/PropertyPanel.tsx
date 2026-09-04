import React from 'react';
import { Empty, Form, Input } from 'antd';
import type {
  ApprovalConfig,
  ConditionConfig,
  DesignerNodeData,
  NotifyConfig,
  ParallelConfig,
  TimerConfig,
} from '@/types/workflow';
import type { DesignerRFNode } from '../utils/flowConvert';
import ApprovalConfigForm from './ApprovalConfigForm';
import ConditionConfigForm from './ConditionConfigForm';
import { NotifyConfigForm, ParallelConfigForm, TimerConfigForm } from './ExtraConfigForms';

interface PropertyPanelProps {
  node: DesignerRFNode | null;
  nodeOptions: Array<{ label: string; value: string }>;
  edgeTargets: Record<string, string>;
  disabled?: boolean;
  onChangeData: (nodeId: string, data: DesignerNodeData) => void;
  onRetarget: (branchId: string, targetId: string) => void;
}

const PropertyPanel: React.FC<PropertyPanelProps> = ({
  node,
  nodeOptions,
  edgeTargets,
  disabled,
  onChangeData,
  onRetarget,
}) => {
  if (!node) {
    return (
      <div className="designer-side-panel">
        <h4>属性面板</h4>
        <Empty description="选中节点后配置属性" image={Empty.PRESENTED_IMAGE_SIMPLE} />
      </div>
    );
  }

  const patchData = (partial: Partial<DesignerNodeData>) => {
    const next = { ...node.data, ...partial };
    next.label = next.name;
    onChangeData(node.id, next);
  };

  return (
    <div className="designer-side-panel">
      <h4>属性面板</h4>
      <Form layout="vertical" size="small">
        <Form.Item label="节点名称">
          <Input
            disabled={disabled}
            value={node.data.name}
            onChange={(event) => patchData({ name: event.target.value })}
          />
        </Form.Item>
        <Form.Item label="节点说明">
          <Input.TextArea
            disabled={disabled}
            rows={2}
            value={node.data.description}
            onChange={(event) => patchData({ description: event.target.value })}
          />
        </Form.Item>
        {node.type === 'approval' && (
          <ApprovalConfigForm
            value={node.data.config as ApprovalConfig}
            disabled={disabled}
            onChange={(config) => patchData({ config })}
          />
        )}
        {node.type === 'condition' && (
          <ConditionConfigForm
            value={node.data.config as ConditionConfig}
            nodeOptions={nodeOptions}
            edgeTargets={edgeTargets}
            disabled={disabled}
            onChange={(config) => patchData({ config })}
            onRetarget={onRetarget}
          />
        )}
        {node.type === 'parallel' && (
          <ParallelConfigForm
            value={node.data.config as ParallelConfig}
            disabled={disabled}
            onChange={(config) => patchData({ config })}
          />
        )}
        {node.type === 'timer' && (
          <TimerConfigForm
            value={node.data.config as TimerConfig}
            disabled={disabled}
            onChange={(config) => patchData({ config })}
          />
        )}
        {node.type === 'notify' && (
          <NotifyConfigForm
            value={node.data.config as NotifyConfig}
            disabled={disabled}
            onChange={(config) => patchData({ config })}
          />
        )}
      </Form>
    </div>
  );
};

export default PropertyPanel;
