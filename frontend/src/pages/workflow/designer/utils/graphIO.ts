import type { FlowGraphData } from '@/types/workflow';
import { validateDraftGraph } from './graphValidate';

export function parseImportedGraph(raw: string): FlowGraphData {
  const parsed: unknown = JSON.parse(raw);
  if (!parsed || typeof parsed !== 'object') {
    throw new Error('导入文件不是有效的 JSON 对象');
  }
  const graph = parsed as FlowGraphData;
  const issues = validateDraftGraph(graph);
  if (issues.length > 0) {
    throw new Error(issues[0].message);
  }
  if (graph.nodes.some((node) => !node.id || !node.type || !node.position || !node.data)) {
    throw new Error('导入文件缺少节点必要字段');
  }
  return graph;
}

export function downloadGraphJson(graph: FlowGraphData, filename: string): void {
  const blob = new Blob([JSON.stringify(graph, null, 2)], { type: 'application/json' });
  const url = URL.createObjectURL(blob);
  const link = document.createElement('a');
  link.href = url;
  link.download = filename;
  link.click();
  URL.revokeObjectURL(url);
}
