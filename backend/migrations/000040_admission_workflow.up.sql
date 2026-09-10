-- Append protected admission templates; preserve all existing versions and drafts.

DO $migration$
DECLARE target_definition_id uuid; next_version integer;
BEGIN
 INSERT INTO flow_definitions(id,key,name,description,category,status,created_at,updated_at)
 VALUES(uuid_generate_v4(),'member_admission','会员资料审核','资料审核与真实签字；评分不自动决定录用','member',0,NOW(),NOW())
 ON CONFLICT(key) DO NOTHING;
 SELECT id INTO target_definition_id FROM flow_definitions WHERE key='member_admission' FOR UPDATE;
 SELECT COALESCE(MAX(version),0)+1 INTO next_version FROM flow_definition_versions WHERE flow_definition_versions.definition_id=target_definition_id;
 UPDATE flow_definition_versions SET status=0 WHERE flow_definition_versions.definition_id=target_definition_id AND status=1;
 INSERT INTO flow_definition_versions(id,definition_id,version,bpmn_data,status,published_at,created_at)
 VALUES(uuid_generate_v4(),target_definition_id,next_version,'{"nodes":[{"id":"start","type":"start","position":{"x":280,"y":20},"data":{"label":"提交申请","config":{}}},{"id":"materials","type":"approval","position":{"x":280,"y":140},"data":{"label":"资料审核","config":{"assigneeStrategy":"business_role","businessType":"member_application","admissionStage":"materials","admissionRole":"materials","approvalType":"any","dueDays":1,"allowTransfer":false,"allowRollback":false}}},{"id":"end","type":"end","position":{"x":280,"y":260},"data":{"label":"正式会员","config":{}}}],"edges":[{"id":"start-materials","source":"start","target":"materials"},{"id":"materials-end","source":"materials","target":"end"}]}'::jsonb,1,NOW(),NOW());
 UPDATE flow_definitions SET status=1,updated_at=NOW() WHERE id=target_definition_id;
END $migration$;

DO $migration$
DECLARE target_definition_id uuid; next_version integer;
BEGIN
 INSERT INTO flow_definitions(id,key,name,description,category,status,created_at,updated_at)
 VALUES(uuid_generate_v4(),'officer_interview','干事面试与正式签字','资料审核与真实签字；评分不自动决定录用','member',0,NOW(),NOW())
 ON CONFLICT(key) DO NOTHING;
 SELECT id INTO target_definition_id FROM flow_definitions WHERE key='officer_interview' FOR UPDATE;
 SELECT COALESCE(MAX(version),0)+1 INTO next_version FROM flow_definition_versions WHERE flow_definition_versions.definition_id=target_definition_id;
 UPDATE flow_definition_versions SET status=0 WHERE flow_definition_versions.definition_id=target_definition_id AND status=1;
 INSERT INTO flow_definition_versions(id,definition_id,version,bpmn_data,status,published_at,created_at)
 VALUES(uuid_generate_v4(),target_definition_id,next_version,'{"nodes":[{"id":"start","type":"start","position":{"x":280,"y":20},"data":{"label":"提交申请","config":{}}},{"id":"materials","type":"approval","position":{"x":280,"y":140},"data":{"label":"资料审核","config":{"assigneeStrategy":"business_role","businessType":"member_application","admissionStage":"materials","admissionRole":"materials","approvalType":"any","dueDays":1,"allowTransfer":false,"allowRollback":false}}},{"id":"fork","type":"parallel_gateway","position":{"x":280,"y":260},"data":{"label":"一面后分别签字","config":{}}},{"id":"minister1","type":"approval","position":{"x":100,"y":380},"data":{"label":"一面 · 部长签字","config":{"assigneeStrategy":"business_role","businessType":"member_application","admissionStage":"round1","admissionRole":"minister","approvalType":"any","dueDays":1,"allowTransfer":false,"allowRollback":false}}},{"id":"center1","type":"approval","position":{"x":460,"y":380},"data":{"label":"一面 · 中心签字","config":{"assigneeStrategy":"business_role","businessType":"member_application","admissionStage":"round1","admissionRole":"center","approvalType":"any","dueDays":1,"allowTransfer":false,"allowRollback":false}}},{"id":"join","type":"parallel_gateway","position":{"x":280,"y":500},"data":{"label":"两份签字齐全","config":{}}},{"id":"center2","type":"approval","position":{"x":280,"y":620},"data":{"label":"二面 · 中心签字","config":{"assigneeStrategy":"business_role","businessType":"member_application","admissionStage":"round2","admissionRole":"center","approvalType":"any","dueDays":1,"allowTransfer":false,"allowRollback":false}}},{"id":"president","type":"approval","position":{"x":280,"y":740},"data":{"label":"会长最终确认","config":{"assigneeStrategy":"business_role","businessType":"member_application","admissionStage":"president","admissionRole":"president","approvalType":"any","dueDays":1,"allowTransfer":false,"allowRollback":false}}},{"id":"end","type":"end","position":{"x":280,"y":860},"data":{"label":"录用审批完成","config":{}}}],"edges":[{"id":"start-materials","source":"start","target":"materials"},{"id":"materials-fork","source":"materials","target":"fork"},{"id":"fork-minister1","source":"fork","target":"minister1"},{"id":"fork-center1","source":"fork","target":"center1"},{"id":"minister1-join","source":"minister1","target":"join"},{"id":"center1-join","source":"center1","target":"join"},{"id":"join-center2","source":"join","target":"center2"},{"id":"center2-president","source":"center2","target":"president"},{"id":"president-end","source":"president","target":"end"}]}'::jsonb,1,NOW(),NOW());
 UPDATE flow_definitions SET status=1,updated_at=NOW() WHERE id=target_definition_id;
END $migration$;
