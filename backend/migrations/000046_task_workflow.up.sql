-- Opt-in lifecycle; legacy task states and history are unchanged.
ALTER TABLE tasks
 ADD COLUMN workflow_revision bigint NOT NULL DEFAULT 0,
 ADD COLUMN workflow_instance_id uuid REFERENCES flow_instances(id),
 ADD COLUMN workflow_stage varchar(32) NOT NULL DEFAULT '',
 ADD COLUMN reviewer_id uuid REFERENCES users(id),
 ADD COLUMN acceptor_id uuid REFERENCES users(id),
 ADD COLUMN submission text NOT NULL DEFAULT '';
CREATE UNIQUE INDEX tasks_workflow_instance_unique ON tasks(workflow_instance_id) WHERE workflow_instance_id IS NOT NULL;
CREATE INDEX tasks_reviewer_active ON tasks(reviewer_id,workflow_stage) WHERE deleted_at IS NULL;
CREATE INDEX tasks_acceptor_active ON tasks(acceptor_id,workflow_stage) WHERE deleted_at IS NULL;
DO $migration$
DECLARE definition_id_value uuid; next_version integer;
BEGIN
 INSERT INTO flow_definitions(id,key,name,description,category,status,created_at,updated_at)
 VALUES(uuid_generate_v4(),'task_lifecycle','任务交付与验收','发布、分配、执行、审核、验收；交付不能自动代替签字','task',0,NOW(),NOW())
 ON CONFLICT(key) DO NOTHING;
 SELECT id INTO definition_id_value FROM flow_definitions WHERE key='task_lifecycle' FOR UPDATE;
 SELECT COALESCE(MAX(version),0)+1 INTO next_version FROM flow_definition_versions WHERE definition_id=definition_id_value;
 UPDATE flow_definition_versions SET status=0 WHERE definition_id=definition_id_value AND status=1;
 INSERT INTO flow_definition_versions(id,definition_id,version,bpmn_data,status,published_at,created_at)
 VALUES(uuid_generate_v4(),definition_id_value,next_version,'{"nodes":[{"id":"start","type":"start","position":{"x":280,"y":0},"data":{"label":"发布任务","config":{}}},{"id":"assignment","type":"approval","position":{"x":280,"y":140},"data":{"label":"认领 / 分配","config":{"assigneeStrategy":"business_role","businessType":"collaboration_task","taskStage":"assignment","approvalType":"single","dueDays":1,"allowTransfer":false,"allowRollback":false}}},{"id":"execution","type":"approval","position":{"x":280,"y":280},"data":{"label":"执行与提交交付","config":{"assigneeStrategy":"business_role","businessType":"collaboration_task","taskStage":"execution","approvalType":"single","dueDays":1,"allowTransfer":false,"allowRollback":false}}},{"id":"review","type":"approval","position":{"x":280,"y":420},"data":{"label":"负责人审核","config":{"assigneeStrategy":"business_role","businessType":"collaboration_task","taskStage":"review","approvalType":"single","dueDays":1,"allowTransfer":false,"allowRollback":false}}},{"id":"acceptance","type":"approval","position":{"x":280,"y":560},"data":{"label":"正式验收","config":{"assigneeStrategy":"business_role","businessType":"collaboration_task","taskStage":"acceptance","approvalType":"single","dueDays":1,"allowTransfer":false,"allowRollback":false}}},{"id":"end","type":"end","position":{"x":280,"y":700},"data":{"label":"任务完成","config":{}}}],"edges":[{"id":"start-assignment","source":"start","target":"assignment"},{"id":"assignment-execution","source":"assignment","target":"execution"},{"id":"execution-review","source":"execution","target":"review"},{"id":"review-acceptance","source":"review","target":"acceptance"},{"id":"acceptance-end","source":"acceptance","target":"end"}]}'::jsonb,1,NOW(),NOW());
 UPDATE flow_definitions SET status=1,updated_at=NOW() WHERE id=definition_id_value;
END $migration$;
