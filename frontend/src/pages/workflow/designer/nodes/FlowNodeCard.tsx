import React from 'react';
import { Handle, Position, type NodeProps } from 'reactflow';
import type { ConditionConfig, DesignerNodeData, DesignerNodeType } from '@/types/workflow';
import { NODE_META } from '../constants';

const ICONS: Record<DesignerNodeType, string> = {
  start: '▶',
  end: '■',
  approval: '✓',
  condition: '◇',
  parallel: '⑂',
  merge: '⫻',
  timer: '⏱',
  notify: '🔔',
};

const FlowNodeCard: React.FC<NodeProps<DesignerNodeData>> = ({ data, selected, type }) => {
  const nodeType = (type ?? 'approval') as DesignerNodeType;
  const meta = NODE_META[nodeType];
  const isCircle = meta.shape === 'circle';
  const isDiamond = meta.shape === 'diamond';
  const size = isCircle ? 72 : undefined;

  const boxStyle: React.CSSProperties = isDiamond
    ? {
        width: 88,
        height: 88,
        transform: 'rotate(45deg)',
        background: '#fff',
        border: `2px solid ${meta.color}`,
        boxShadow: selected ? `0 0 0 3px ${meta.color}33` : '0 2px 6px rgba(0,0,0,0.08)',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
      }
    : {
        minWidth: isCircle ? size : 132,
        height: isCircle ? size : 56,
        borderRadius: isCircle ? '50%' : 8,
        background: '#fff',
        border: `2px solid ${meta.color}`,
        boxShadow: selected ? `0 0 0 3px ${meta.color}33` : '0 2px 6px rgba(0,0,0,0.08)',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        gap: 6,
        padding: isCircle ? 0 : '0 12px',
      };

  const innerStyle: React.CSSProperties = isDiamond
    ? { transform: 'rotate(-45deg)', textAlign: 'center', fontSize: 12, fontWeight: 600 }
    : { fontSize: 13, fontWeight: 600, color: '#1f1f1f' };

  return (
    <div style={boxStyle}>
      {nodeType !== 'start' && (
        <Handle type="target" position={Position.Top} style={{ background: meta.color }} />
      )}
      <div style={innerStyle}>
        <span style={{ color: meta.color, marginRight: 4 }}>{ICONS[nodeType]}</span>
        {data.name || meta.label}
      </div>
      {nodeType !== 'end' && nodeType !== 'condition' && (
        <Handle type="source" position={Position.Bottom} style={{ background: meta.color }} />
      )}
      {nodeType === 'condition' &&
        ((data.config as ConditionConfig).branches ?? []).map((branch, index) => (
          <Handle
            key={branch.id}
            id={branch.id}
            type="source"
            position={index % 2 === 0 ? Position.Right : Position.Bottom}
            style={{ background: meta.color, top: index % 2 === 0 ? 24 + index * 8 : undefined }}
          />
        ))}
    </div>
  );
};

export default FlowNodeCard;
