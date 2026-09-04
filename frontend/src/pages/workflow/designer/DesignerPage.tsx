import React, { useEffect, useMemo, useRef, useState } from 'react';
import { message, Modal } from 'antd';
import { useNavigate, useParams } from 'react-router-dom';
import { useDispatch, useSelector } from 'react-redux';
import { ReactFlowProvider, addEdge, applyEdgeChanges, applyNodeChanges } from 'reactflow';
import type { Connection, EdgeChange, NodeChange } from 'reactflow';
import type { CreateDefinitionPayload } from '@/types/workflow';
import type { AppDispatch } from '@/store';
import {
  resetDesigner,
  selectPreviewMode,
  selectSelectedNodeId,
  setDefinitionMeta,
  setDirty,
  setPreviewMode,
  setSelectedNodeId,
} from '@/store/slices/workflowSlice';
import DesignerToolbar from './DesignerToolbar';
import NodePanel from './NodePanel';
import PropertyPanel from './panels/PropertyPanel';
import DesignerCanvas from './DesignerCanvas';
import { CreateDefinitionModal, OpenDefinitionModal } from './DesignerModals';
import { useUndoRedo } from './hooks/useUndoRedo';
import { useDesignerHotkeys } from './hooks/useDesignerHotkeys';
import { loadDefinitionGraph, persistDraft, persistPublish } from './hooks/useDesignerPersistence';
import { downloadGraphJson, parseImportedGraph } from './utils/graphIO';
import { toFlowGraph, fromFlowGraph, type DesignerRFEdge, type DesignerRFNode } from './utils/flowConvert';
import type { ConditionConfig, DesignerNodeData } from '@/types/workflow';
import './designer.css';

const DesignerPageInner: React.FC = () => {
  const { id } = useParams<{ id?: string }>();
  const navigate = useNavigate();
  const dispatch = useDispatch<AppDispatch>();
  const previewMode = useSelector(selectPreviewMode);
  const selectedNodeId = useSelector(selectSelectedNodeId);
  const [nodes, setNodes] = useState<DesignerRFNode[]>([]);
  const [edges, setEdges] = useState<DesignerRFEdge[]>([]);
  const [saving, setSaving] = useState(false);
  const [publishing, setPublishing] = useState(false);
  const [createOpen, setCreateOpen] = useState(false);
  const [openOpen, setOpenOpen] = useState(false);
  const [pendingAction, setPendingAction] = useState<'draft' | 'publish' | null>(null);
  const fileRef = useRef<HTMLInputElement>(null);
  const { takeSnapshot, undo, redo, resetHistory, canUndo, canRedo } = useUndoRedo(
    nodes,
    edges,
    setNodes,
    setEdges,
  );

  useEffect(() => {
    dispatch(resetDesigner());
    if (!id) return;
    loadDefinitionGraph(id)
      .then((result) => {
        const converted = fromFlowGraph(result.graph);
        setNodes(converted.nodes);
        setEdges(converted.edges);
        resetHistory();
        dispatch(setDefinitionMeta({
          id,
          name: result.name,
          key: result.key,
          status: result.status,
        }));
      })
      .catch(() => message.error('加载流程失败'));
    // resetHistory 仅重置栈，不作为加载触发条件
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [id, dispatch]);

  useDesignerHotkeys(undo, redo);

  const selectedNode = nodes.find((node) => node.id === selectedNodeId) ?? null;
  const graph = useMemo(() => toFlowGraph(nodes, edges), [nodes, edges]);

  const applyGraph = (next: { nodes: DesignerRFNode[]; edges: DesignerRFEdge[] }) => {
    takeSnapshot();
    setNodes(next.nodes);
    setEdges(next.edges);
    dispatch(setDirty(true));
  };

  const handleNodesChange = (changes: NodeChange[]) => {
    setNodes((current) => applyNodeChanges(changes, current) as DesignerRFNode[]);
    dispatch(setDirty(true));
  };

  const handleEdgesChange = (changes: EdgeChange[]) => {
    const structural = changes.some((item) => item.type === 'remove' || item.type === 'add');
    if (structural) takeSnapshot();
    setEdges((current) => applyEdgeChanges(changes, current));
    dispatch(setDirty(true));
  };

  const handleConnect = (connection: Connection) => {
    takeSnapshot();
    const sourceNode = nodes.find((node) => node.id === connection.source);
    const branches = (sourceNode?.data.config as ConditionConfig | undefined)?.branches;
    const branch = branches?.find((item) => item.id === connection.sourceHandle);
    setEdges((current) =>
      addEdge(
        {
          ...connection,
          id: `e_${Date.now().toString(36)}`,
          label: branch?.expression || branch?.label,
          data: branch?.expression ? { condition: branch.expression } : undefined,
        },
        current,
      ),
    );
    dispatch(setDirty(true));
  };

  const handleChangeData = (nodeId: string, data: DesignerNodeData) => {
    takeSnapshot();
    setNodes((current) => current.map((node) => (node.id === nodeId ? { ...node, data } : node)));
    const config = data.config as ConditionConfig;
    if (config.branches) {
      setEdges((current) =>
        current.map((edge) => {
          const branch = config.branches.find((item) => item.id === edge.sourceHandle);
          if (!branch || edge.source !== nodeId) return edge;
          return { ...edge, label: branch.expression || branch.label, data: { condition: branch.expression } };
        }),
      );
    }
    dispatch(setDirty(true));
  };

  const handleRetarget = (branchId: string, targetId: string) => {
    if (!selectedNodeId) return;
    takeSnapshot();
    setEdges((current) => {
      const others = current.filter(
        (edge) => !(edge.source === selectedNodeId && edge.sourceHandle === branchId),
      );
      if (!targetId) return others;
      return [
        ...others,
        {
          id: `e_${selectedNodeId}_${branchId}`,
          source: selectedNodeId,
          target: targetId,
          sourceHandle: branchId,
        },
      ];
    });
  };

  const runSave = async (payload?: CreateDefinitionPayload) => {
    setSaving(true);
    try {
      const nextId = await persistDraft(id ?? null, graph, payload);
      dispatch(setDirty(false));
      if (!id) navigate(`/workflow/designer/${nextId}`, { replace: true });
    } catch (error) {
      if (error instanceof Error && error.message === 'NEED_CREATE') {
        setPendingAction('draft');
        setCreateOpen(true);
        return;
      }
      message.error(error instanceof Error ? error.message : '保存失败');
    } finally {
      setSaving(false);
    }
  };

  const runPublish = async (payload?: CreateDefinitionPayload) => {
    setPublishing(true);
    try {
      let defId = id;
      if (!defId) {
        defId = await persistDraft(null, graph, payload);
        navigate(`/workflow/designer/${defId}`, { replace: true });
      }
      await persistPublish(defId, graph);
      dispatch(setDirty(false));
    } catch (error) {
      if (error instanceof Error && error.message === 'NEED_CREATE') {
        setPendingAction('publish');
        setCreateOpen(true);
        return;
      }
      message.error(error instanceof Error ? error.message : '发布失败');
    } finally {
      setPublishing(false);
    }
  };

  const onCreateOk = async (payload: CreateDefinitionPayload) => {
    setCreateOpen(false);
    if (pendingAction === 'publish') await runPublish(payload);
    else await runSave(payload);
    setPendingAction(null);
  };

  const handleImportFile = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    event.target.value = '';
    if (!file) return;
    file.text().then((text) => {
      try {
        const imported = parseImportedGraph(text);
        Modal.confirm({
          title: '导入将替换当前画布',
          onOk: () => applyGraph(fromFlowGraph(imported)),
        });
      } catch (error) {
        message.error(error instanceof Error ? error.message : '导入失败');
      }
    });
  };

  const edgeTargets = useMemo(() => {
    const map: Record<string, string> = {};
    edges.forEach((edge) => {
      if (edge.source === selectedNodeId && edge.sourceHandle) {
        map[edge.sourceHandle] = edge.target;
      }
    });
    return map;
  }, [edges, selectedNodeId]);

  return (
    <div className="designer-page">
      <DesignerToolbar
        title={id ? '流程设计' : '新建流程'}
        previewMode={previewMode}
        saving={saving}
        publishing={publishing}
        canUndo={canUndo}
        canRedo={canRedo}
        onSaveDraft={() => runSave()}
        onPublish={() => runPublish()}
        onTogglePreview={() => dispatch(setPreviewMode(!previewMode))}
        onImport={() => fileRef.current?.click()}
        onExport={() => downloadGraphJson(graph, `workflow-${id || 'draft'}.json`)}
        onUndo={undo}
        onRedo={redo}
        onOpen={() => setOpenOpen(true)}
        onNew={() => navigate('/workflow/designer')}
      />
      <div className="designer-body">
        <NodePanel disabled={previewMode} />
        <DesignerCanvas
          nodes={nodes}
          edges={edges}
          previewMode={previewMode}
          onNodesChange={handleNodesChange}
          onEdgesChange={handleEdgesChange}
          onConnect={handleConnect}
          onNodeClick={(nodeId) => dispatch(setSelectedNodeId(nodeId))}
          onDropNode={(node) => applyGraph({ nodes: [...nodes, node], edges })}
          onDragStop={() => takeSnapshot()}
        />
        <PropertyPanel
          node={selectedNode}
          nodeOptions={nodes.filter((n) => n.id !== selectedNodeId).map((n) => ({ label: n.data.name, value: n.id }))}
          edgeTargets={edgeTargets}
          disabled={previewMode}
          onChangeData={handleChangeData}
          onRetarget={handleRetarget}
        />
      </div>
      <input ref={fileRef} type="file" accept="application/json" hidden onChange={handleImportFile} />
      <CreateDefinitionModal
        open={createOpen}
        confirmLoading={saving || publishing}
        onCancel={() => setCreateOpen(false)}
        onOk={onCreateOk}
      />
      <OpenDefinitionModal
        open={openOpen}
        onCancel={() => setOpenOpen(false)}
        onSelect={(nextId) => {
          setOpenOpen(false);
          navigate(`/workflow/designer/${nextId}`);
        }}
      />
    </div>
  );
};

const DesignerPage: React.FC = () => (
  <ReactFlowProvider>
    <DesignerPageInner />
  </ReactFlowProvider>
);

export default DesignerPage;
