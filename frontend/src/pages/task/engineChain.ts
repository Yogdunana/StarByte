export const TASK_ENGINE_STAGES = ['assignment', 'execution', 'review', 'acceptance', 'completed'] as const;

export type TaskEngineStage = (typeof TASK_ENGINE_STAGES)[number] | 'cancelled' | 'rejected';

export function taskEngineClosed(stage: string): boolean {
  return stage === 'completed' || stage === 'cancelled' || stage === 'rejected';
}

export function taskEngineStep(stage: string): number {
  const index = TASK_ENGINE_STAGES.indexOf(stage as (typeof TASK_ENGINE_STAGES)[number]);
  return index < 0 ? 0 : index;
}

export function taskEngineStatus(stage: string): 'error' | 'finish' | 'process' {
  if (stage === 'cancelled' || stage === 'rejected') return 'error';
  if (stage === 'completed') return 'finish';
  return 'process';
}
