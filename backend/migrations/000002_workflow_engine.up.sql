-- ============================================================
-- 000002_workflow_engine.up.sql
-- 流程引擎相关表
-- ============================================================

-- 流程定义表
CREATE TABLE flow_definitions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    key VARCHAR(100) NOT NULL UNIQUE,
    name VARCHAR(200) NOT NULL,
    description TEXT,
    category VARCHAR(50) DEFAULT 'custom', -- interview, member, task, custom
    status SMALLINT DEFAULT 0, -- 0=草稿 1=已发布 2=已停用
    created_by UUID REFERENCES users(id),
    updated_by UUID REFERENCES users(id),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_flow_definitions_key ON flow_definitions(key);
CREATE INDEX idx_flow_definitions_category ON flow_definitions(category);
CREATE INDEX idx_flow_definitions_status ON flow_definitions(status);

-- 流程定义版本表
CREATE TABLE flow_definition_versions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    definition_id UUID NOT NULL REFERENCES flow_definitions(id) ON DELETE CASCADE,
    version INT NOT NULL,
    bpmn_data JSONB NOT NULL,
    status SMALLINT DEFAULT 0, -- 0=历史 1=当前版本
    published_by UUID REFERENCES users(id),
    published_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(definition_id, version)
);

CREATE INDEX idx_flow_def_versions_def_id ON flow_definition_versions(definition_id);
CREATE INDEX idx_flow_def_versions_status ON flow_definition_versions(status);

-- 流程实例表
CREATE TABLE flow_instances (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    definition_id UUID NOT NULL REFERENCES flow_definitions(id),
    definition_version_id UUID NOT NULL REFERENCES flow_definition_versions(id),
    business_key VARCHAR(100),
    business_type VARCHAR(50),
    initiator_id UUID NOT NULL REFERENCES users(id),
    status SMALLINT DEFAULT 0, -- 0=运行中 1=已完成 2=已终止 3=已挂起
    current_node_ids JSONB,
    started_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    ended_at TIMESTAMP,
    terminate_reason TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_flow_instances_def_id ON flow_instances(definition_id);
CREATE INDEX idx_flow_instances_initiator ON flow_instances(initiator_id);
CREATE INDEX idx_flow_instances_status ON flow_instances(status);
CREATE INDEX idx_flow_instances_business ON flow_instances(business_type, business_key);

-- 流程任务表
CREATE TABLE flow_tasks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    instance_id UUID NOT NULL REFERENCES flow_instances(id) ON DELETE CASCADE,
    node_id VARCHAR(100) NOT NULL,
    node_name VARCHAR(200),
    task_type VARCHAR(50) DEFAULT 'approval', -- approval, notification, service
    assignee_id UUID REFERENCES users(id),
    status SMALLINT DEFAULT 0, -- 0=待处理 1=已通过 2=已驳回 3=已转办 4=已撤回 5=已取消
    action VARCHAR(50),
    comment TEXT,
    form_data JSONB,
    due_date TIMESTAMP,
    claimed_at TIMESTAMP,
    completed_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_flow_tasks_instance_id ON flow_tasks(instance_id);
CREATE INDEX idx_flow_tasks_assignee ON flow_tasks(assignee_id);
CREATE INDEX idx_flow_tasks_status ON flow_tasks(status);
CREATE INDEX idx_flow_tasks_created_at ON flow_tasks(created_at);

-- 流程历史记录表
CREATE TABLE flow_histories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    instance_id UUID NOT NULL REFERENCES flow_instances(id) ON DELETE CASCADE,
    task_id UUID REFERENCES flow_tasks(id),
    node_id VARCHAR(100),
    node_name VARCHAR(200),
    node_type VARCHAR(50),
    operator_id UUID REFERENCES users(id),
    action VARCHAR(50) NOT NULL, -- start, approve, reject, transfer, rollback, complete, terminate
    comment TEXT,
    from_node_id VARCHAR(100),
    to_node_id VARCHAR(100),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_flow_histories_instance_id ON flow_histories(instance_id);
CREATE INDEX idx_flow_histories_operator ON flow_histories(operator_id);
CREATE INDEX idx_flow_histories_created_at ON flow_histories(created_at);

-- 流程变量表
CREATE TABLE flow_variables (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    instance_id UUID NOT NULL REFERENCES flow_instances(id) ON DELETE CASCADE,
    key VARCHAR(100) NOT NULL,
    value JSONB,
    scope VARCHAR(20) DEFAULT 'global', -- global, local
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(instance_id, key, scope)
);

CREATE INDEX idx_flow_variables_instance_id ON flow_variables(instance_id);
