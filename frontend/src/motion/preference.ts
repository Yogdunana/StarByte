export function resolveMotionMode(userReduce: boolean): 'always' | 'user' {
  return userReduce ? 'always' : 'user';
}

export function motionDataset(userReduce: boolean): 'reduced' | 'system' {
  return userReduce ? 'reduced' : 'system';
}
