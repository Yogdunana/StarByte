import type {
  ApprovalConfig,
  ConditionConfig,
  FlowGraphData,
  ValidationIssue,
} from '@/types/workflow';

function outgoing(graph: FlowGraphData, nodeId: string) {
  return graph.edges.filter((e) => e.source === nodeId);
}

function reachableFromStart(graph: FlowGraphData, startId: string): Set<string> {
  const seen = new Set<string>();
  const queue = [startId];
  while (queue.length > 0) {
    const current = queue.shift();
    if (!current || seen.has(current)) continue;
    seen.add(current);
    outgoing(graph, current).forEach((edge) => queue.push(edge.target));
  }
  return seen;
}

function validateApproval(graph: FlowGraphData, issues: ValidationIssue[]): void {
  graph.nodes
    .filter((node) => node.type === 'approval')
    .forEach((node) => {
      const config = node.data.config as ApprovalConfig;
      const strategy = config?.assigneeStrategy;
      if (!strategy) {
        issues.push({ level: 'error', message: `审批节点「${node.data.name}」未配置审批人类型` });
        return;
      }
      if (strategy === 'static' && (!config.assignees || config.assignees.length === 0)) {
        issues.push({ level: 'error', message: `审批节点「${node.data.name}」未选择审批人` });
      }
      if (strategy === 'role' && !config.roleId) {
        issues.push({ level: 'error', message: `审批节点「${node.data.name}」未选择角色` });
      }
    });
}

function validateConditions(graph: FlowGraphData, issues: ValidationIssue[]): void {
  graph.nodes
    .filter((node) => node.type === 'condition')
    .forEach((node) => {
      const config = node.data.config as ConditionConfig;
      const edges = outgoing(graph, node.id);
      if (edges.length < 2) {
        issues.push({ level: 'error', message: `条件节点「${node.data.name}」至少需要两条出线` });
      }
      edges.forEach((edge) => {
        const branch = config?.branches?.find((item) => item.id === edge.sourceHandle);
        const expression = branch?.expression || edge.data?.condition || '';
        if (!expression && !branch?.is_default) {
          issues.push({
            level: 'error',
            message: `条件节点「${node.data.name}」的连线缺少条件表达式`,
          });
        }
      });
    });
}

/** 草稿：只检查 JSON 结构 */
export function validateDraftGraph(graph: FlowGraphData): ValidationIssue[] {
  if (!graph || !Array.isArray(graph.nodes) || !Array.isArray(graph.edges)) {
    return [{ level: 'error', message: '流程图数据格式无效' }];
  }
  return [];
}

/** 发布：完整合法性校验（含 timer 拦截） */
export function validatePublishGraph(graph: FlowGraphData): ValidationIssue[] {
  const issues = validateDraftGraph(graph);
  if (issues.length > 0) return issues;

  const starts = graph.nodes.filter((n) => n.type === 'start');
  const ends = graph.nodes.filter((n) => n.type === 'end');
  if (starts.length !== 1) {
    issues.push({ level: 'error', message: '必须有且仅有一个开始节点' });
  }
  if (ends.length < 1) {
    issues.push({ level: 'error', message: '至少需要一个结束节点' });
  }

  const timers = graph.nodes.filter((n) => n.type === 'timer');
  if (timers.length > 0) {
    issues.push({
      level: 'error',
      message: '定时器节点一期后端未实现，请删除后再发布',
    });
  }

  if (starts.length === 1) {
    const reachable = reachableFromStart(graph, starts[0].id);
    graph.nodes.forEach((node) => {
      if (!reachable.has(node.id)) {
        issues.push({ level: 'error', message: `节点「${node.data.name}」从开始节点不可达` });
      }
    });
  }

  validateApproval(graph, issues);
  validateConditions(graph, issues);

  const parallels = graph.nodes.filter((n) => n.type === 'parallel');
  const merges = graph.nodes.filter((n) => n.type === 'merge');
  if (parallels.length > 0 && merges.length === 0) {
    issues.push({ level: 'error', message: '并行分支必须有对应的合并节点' });
  }
  if (starts.length === 1 && parallels.length > 0) {
    parallels.forEach((node) => {
      const reachable = reachableFromStart(graph, node.id);
      const hasMerge = graph.nodes.some((item) => item.type === 'merge' && reachable.has(item.id));
      if (!hasMerge) {
        issues.push({ level: 'error', message: `并行节点「${node.data.name}」缺少对应的合并节点` });
      }
    });
  }

  return issues;
}
