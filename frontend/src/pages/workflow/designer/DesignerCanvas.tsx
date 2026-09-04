import React, { useCallback, useRef } from 'react';
import ReactFlow, {
  Background,
  Controls,
  MiniMap,
  type Connection,
  type EdgeChange,
  type NodeChange,
  type ReactFlowInstance,
  useReactFlow,
} from 'reactflow';
import 'reactflow/dist/style.css';
import { DND_TYPE } from './constants';
import { designerNodeTypes } from './nodes/nodeTypes';
import {
  createNodeId,
  defaultNodeData,
  type DesignerRFEdge,
  type DesignerRFNode,
} from './utils/flowConvert';
import type { DesignerNodeType } from '@/types/workflow';

interface DesignerCanvasProps {
  nodes: DesignerRFNode[];
  edges: DesignerRFEdge[];
  previewMode: boolean;
  onNodesChange: (changes: NodeChange[]) => void;
  onEdgesChange: (changes: EdgeChange[]) => void;
  onConnect: (connection: Connection) => void;
  onNodeClick: (nodeId: string | null) => void;
  onDropNode: (node: DesignerRFNode) => void;
  onDragStop: () => void;
}

const DesignerCanvasInner: React.FC<DesignerCanvasProps> = ({
  nodes,
  edges,
  previewMode,
  onNodesChange,
  onEdgesChange,
  onConnect,
  onNodeClick,
  onDropNode,
  onDragStop,
}) => {
  const wrapperRef = useRef<HTMLDivElement>(null);
  const instanceRef = useRef<ReactFlowInstance | null>(null);
  const { screenToFlowPosition } = useReactFlow();

  const handleDrop = useCallback(
    (event: React.DragEvent) => {
      event.preventDefault();
      if (previewMode) return;
      const type = (event.dataTransfer.getData(DND_TYPE) ||
        event.dataTransfer.getData('text/plain')) as DesignerNodeType;
      if (!type || !type.trim()) return;
      const bounds = wrapperRef.current?.getBoundingClientRect();
      const instance = instanceRef.current;
      const position = instance && bounds
        ? instance.project({
            x: event.clientX - bounds.left,
            y: event.clientY - bounds.top,
          })
        : screenToFlowPosition({ x: event.clientX, y: event.clientY });
      onDropNode({
        id: createNodeId(type),
        type,
        position,
        data: defaultNodeData(type),
      });
    },
    [onDropNode, previewMode, screenToFlowPosition],
  );

  return (
    <div className={`designer-canvas${previewMode ? ' is-preview' : ''}`} ref={wrapperRef}>
      <ReactFlow
        nodes={nodes}
        edges={edges}
        nodeTypes={designerNodeTypes}
        onInit={(instance) => {
          instanceRef.current = instance;
          instance.fitView();
        }}
        onNodesChange={onNodesChange}
        onEdgesChange={previewMode ? undefined : onEdgesChange}
        onConnect={(connection) => {
          if (!previewMode) onConnect(connection);
        }}
        onNodeClick={(_, node) => onNodeClick(node.id)}
        onPaneClick={() => onNodeClick(null)}
        onNodeDragStop={onDragStop}
        onDrop={handleDrop}
        onDragOver={(event) => {
          event.preventDefault();
          event.dataTransfer.dropEffect = 'move';
        }}
        nodesDraggable={!previewMode}
        nodesConnectable={!previewMode}
        elementsSelectable
        minZoom={0.4}
        maxZoom={1.6}
        deleteKeyCode={previewMode ? null : ['Backspace', 'Delete']}
      >
        <Background />
        <Controls />
        <MiniMap />
      </ReactFlow>
    </div>
  );
};

export default DesignerCanvasInner;
