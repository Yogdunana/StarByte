import React from 'react';
import { NODE_PALETTE, DND_TYPE } from './constants';
import type { DesignerNodeType } from '@/types/workflow';

interface NodePanelProps {
  disabled?: boolean;
}

const NodePanel: React.FC<NodePanelProps> = ({ disabled }) => {
  const onDragStart = (event: React.DragEvent, type: DesignerNodeType) => {
    if (disabled) return;
    event.dataTransfer.setData(DND_TYPE, type);
    event.dataTransfer.effectAllowed = 'move';
  };

  return (
    <div className="designer-side-panel">
      <h4>节点面板</h4>
      <div className="node-palette">
        {NODE_PALETTE.map((item) => (
          <div
            key={item.type}
            className="node-palette-item"
            draggable={!disabled}
            onDragStart={(event) => onDragStart(event, item.type)}
            style={{ borderColor: item.color, opacity: disabled ? 0.5 : 1 }}
          >
            <span className="node-palette-dot" style={{ background: item.color }} />
            {item.label}
          </div>
        ))}
      </div>
    </div>
  );
};

export default NodePanel;
