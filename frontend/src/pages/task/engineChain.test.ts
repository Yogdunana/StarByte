import { describe, expect, it } from 'vitest';
import { taskEngineClosed, taskEngineStatus, taskEngineStep } from './engineChain';

describe('task engine helpers', () => {
  it('hides actions for completed, cancelled, and rejected', () => {
    expect(taskEngineClosed('assignment')).toBe(false);
    expect(taskEngineClosed('review')).toBe(false);
    expect(taskEngineClosed('completed')).toBe(true);
    expect(taskEngineClosed('cancelled')).toBe(true);
    expect(taskEngineClosed('rejected')).toBe(true);
  });

  it('maps stages to step index and status', () => {
    expect(taskEngineStep('assignment')).toBe(0);
    expect(taskEngineStep('acceptance')).toBe(3);
    expect(taskEngineStep('rejected')).toBe(0);
    expect(taskEngineStatus('rejected')).toBe('error');
    expect(taskEngineStatus('completed')).toBe('finish');
    expect(taskEngineStatus('review')).toBe('process');
  });
});
