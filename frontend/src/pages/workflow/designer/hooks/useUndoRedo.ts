import { useCallback, useRef, useState } from 'react';
import type { DesignerRFEdge, DesignerRFNode } from '../utils/flowConvert';
import { HISTORY_LIMIT } from '../constants';

interface Snapshot {
  nodes: DesignerRFNode[];
  edges: DesignerRFEdge[];
}

function cloneSnapshot(nodes: DesignerRFNode[], edges: DesignerRFEdge[]): Snapshot {
  return {
    nodes: nodes.map((node) => ({ ...node, position: { ...node.position }, data: { ...node.data } })),
    edges: edges.map((edge) => ({ ...edge })),
  };
}

export function useUndoRedo(
  nodes: DesignerRFNode[],
  edges: DesignerRFEdge[],
  setNodes: (nodes: DesignerRFNode[]) => void,
  setEdges: (edges: DesignerRFEdge[]) => void,
) {
  const pastRef = useRef<Snapshot[]>([]);
  const futureRef = useRef<Snapshot[]>([]);
  const applyingRef = useRef(false);
  const [revision, setRevision] = useState(0);

  const takeSnapshot = useCallback(() => {
    if (applyingRef.current) return;
    pastRef.current.push(cloneSnapshot(nodes, edges));
    if (pastRef.current.length > HISTORY_LIMIT) {
      pastRef.current.shift();
    }
    futureRef.current = [];
    setRevision((value) => value + 1);
  }, [nodes, edges]);

  const undo = useCallback(() => {
    const previous = pastRef.current.pop();
    if (!previous) return;
    futureRef.current.push(cloneSnapshot(nodes, edges));
    applyingRef.current = true;
    setNodes(previous.nodes);
    setEdges(previous.edges);
    applyingRef.current = false;
    setRevision((value) => value + 1);
  }, [nodes, edges, setNodes, setEdges]);

  const redo = useCallback(() => {
    const next = futureRef.current.pop();
    if (!next) return;
    pastRef.current.push(cloneSnapshot(nodes, edges));
    applyingRef.current = true;
    setNodes(next.nodes);
    setEdges(next.edges);
    applyingRef.current = false;
    setRevision((value) => value + 1);
  }, [nodes, edges, setNodes, setEdges]);

  const resetHistory = useCallback(() => {
    pastRef.current = [];
    futureRef.current = [];
    setRevision((value) => value + 1);
  }, []);

  return {
    takeSnapshot,
    undo,
    redo,
    resetHistory,
    canUndo: pastRef.current.length > 0,
    canRedo: futureRef.current.length > 0,
    revision,
  };
}
