UPDATE flow_definitions
SET description = '默认任务链：发布 → 认领/分配 → 执行 → 审核 → 验收 → 完成。可在流程设计器调整标签与布局；增删分支属后续迭代。',
    updated_at = NOW()
WHERE key = 'task_lifecycle';

-- Keep published transfer definitions; they may already have running instances.
DROP TABLE IF EXISTS task_transfer_signatures;
DROP TABLE IF EXISTS task_transfers;
