-- Technical administration is separate from elected/appointed association office.
CREATE TABLE IF NOT EXISTS user_role_departments (
 user_role_id uuid NOT NULL REFERENCES user_roles(id) ON DELETE CASCADE,
 department_id uuid NOT NULL REFERENCES departments(id),
 PRIMARY KEY(user_role_id, department_id)
);
INSERT INTO roles(id,name,code,description,status,is_system,sort_order) VALUES
 (gen_random_uuid(),'中心主任','center_director','中心范围内的协会职务',0,true,2),
 (gen_random_uuid(),'候补成员','probationary','入会通过后一个自然月候补期',0,true,7)
ON CONFLICT(code) DO NOTHING;
-- Preserve existing valid office scopes instead of changing primary departments.
INSERT INTO user_role_departments(user_role_id,department_id)
SELECT ur.id,u.department_id FROM user_roles ur JOIN users u ON u.id=ur.user_id JOIN roles r ON r.id=ur.role_id
WHERE r.code IN ('minister','center_director','vice_president') AND u.department_id IS NOT NULL AND NOT EXISTS (SELECT 1 FROM user_role_departments s WHERE s.user_role_id=ur.id)
ON CONFLICT DO NOTHING;
-- Remove only the demo administrator's seeded association presidency.
DELETE FROM user_roles ur USING roles r,users u
WHERE ur.role_id=r.id AND ur.user_id=u.id AND u.username='admin' AND r.code='president'
AND EXISTS (SELECT 1 FROM user_roles x JOIN roles s ON s.id=x.role_id WHERE x.user_id=u.id AND s.code='super_admin');
UPDATE users u SET position_id=NULL WHERE u.username='admin' AND EXISTS(SELECT 1 FROM positions p WHERE p.id=u.position_id AND p.code='president');
-- A president may manage association business, not technical administration.
DELETE FROM role_permissions rp USING roles r, permissions p
WHERE rp.role_id=r.id AND rp.permission_id=p.id AND r.code='president'
AND (p.resource IN ('role','permission','config','cache','scheduler','monitor','backup','system','session') OR p.code IN ('user:create','user:update','user:delete'));
INSERT INTO role_permissions(id,role_id,permission_id,data_scope)
SELECT gen_random_uuid(),r.id,p.id,'department_and_sub' FROM roles r CROSS JOIN permissions p
WHERE r.code='center_director' AND (p.resource IN ('member','interview','interview_private','task','workflow','meeting') OR p.code IN ('user:read','department:read','position:read','announcement:read','doc:read','file:read'))
ON CONFLICT(role_id,permission_id) DO NOTHING;
INSERT INTO role_permissions(id,role_id,permission_id,data_scope)
SELECT gen_random_uuid(),r.id,p.id,'self' FROM roles r CROSS JOIN permissions p
WHERE r.code='probationary' AND p.code IN ('announcement:read','doc:read') ON CONFLICT(role_id,permission_id) DO NOTHING;
ALTER TABLE member_applications ADD COLUMN IF NOT EXISTS charter_policy boolean NOT NULL DEFAULT false;
ALTER TABLE member_applications ADD COLUMN IF NOT EXISTS review_department_id uuid REFERENCES departments(id);

UPDATE configs SET config_value='{"weights":{"president":2,"vice_president":1,"center_director":1,"minister":0.5,"vice_minister":0.5,"deputy":0.5,"officer":0.25},"default_weight":0}',updated_at=NOW()
WHERE config_key='vote_weight_config' AND config_value::jsonb = '{"weights":{"president":5,"vice_president":4,"minister":3,"deputy":2,"vice_minister":2,"officer":1},"default_weight":1}'::jsonb;
UPDATE positions SET vote_weight=CASE code WHEN 'president' THEN 2 WHEN 'vice_president' THEN 1 WHEN 'center_director' THEN 1 WHEN 'minister' THEN 0.5 WHEN 'vice_minister' THEN 0.5 WHEN 'officer' THEN 0.25 ELSE vote_weight END
WHERE code IN ('president','vice_president','center_director','minister','vice_minister','officer');

-- Only the untouched 000070 template is replaced. Existing instances and custom graphs stay intact.
DO $migration$
DECLARE target uuid; next_version integer;
BEGIN
 SELECT id INTO target FROM flow_definitions WHERE key='member_application' FOR UPDATE;
 IF EXISTS(SELECT 1 FROM flow_definition_versions WHERE definition_id=target AND status=1 AND bpmn_data=$old${"nodes": [{"id": "start", "type": "start", "position": {"x": 280, "y": 20}, "data": {"label": "提交申请", "config": {}}}, {"id": "officer", "type": "approval", "position": {"x": 280, "y": 160}, "data": {"label": "干事审批", "config": {"assigneeStrategy": "role", "roleCode": "officer", "approvalType": "any", "departmentScope": true, "allowReject": true, "allowTransfer": true, "allowRollback": false, "skipWhen": "skip_officer", "skipIfEmpty": true}}}, {"id": "minister", "type": "approval", "position": {"x": 280, "y": 300}, "data": {"label": "部长审批", "config": {"assigneeStrategy": "role", "roleCode": "minister", "approvalType": "any", "departmentScope": true, "allowReject": true, "allowTransfer": true, "allowRollback": false, "skipWhen": "skip_minister", "skipIfEmpty": true}}}, {"id": "president", "type": "approval", "position": {"x": 280, "y": 440}, "data": {"label": "社长审批", "config": {"assigneeStrategy": "role", "roleCode": "president", "approvalType": "any", "allowReject": true, "allowTransfer": true, "allowRollback": false}}}, {"id": "end", "type": "end", "position": {"x": 280, "y": 580}, "data": {"label": "结束", "config": {}}}], "edges": [{"id": "start-officer", "source": "start", "target": "officer"}, {"id": "officer-minister", "source": "officer", "target": "minister"}, {"id": "minister-president", "source": "minister", "target": "president"}, {"id": "president-end", "source": "president", "target": "end"}], "viewport": {"x": 0, "y": 0, "zoom": 1}}$old$::jsonb) THEN
  SELECT COALESCE(MAX(version),0)+1 INTO next_version FROM flow_definition_versions WHERE definition_id=target;
  UPDATE flow_definition_versions SET status=0 WHERE definition_id=target AND status=1;
  INSERT INTO flow_definition_versions(id,definition_id,version,bpmn_data,status,published_at,created_at)
  VALUES(uuid_generate_v4(),target,next_version,$graph${"nodes": [{"id": "start", "type": "start", "position": {"x": 280, "y": 20}, "data": {"label": "提交申请", "config": {}}}, {"id": "department_review", "type": "approval", "position": {"x": 280, "y": 160}, "data": {"label": "部门初审", "config": {"assigneeStrategy": "role", "roleCode": "minister", "approvalType": "any", "departmentScope": true, "requireDepartment": true, "dueDays": 1, "allowReject": true, "allowTransfer": true}}}, {"id": "center_review", "type": "approval", "position": {"x": 280, "y": 300}, "data": {"label": "中心复审", "config": {"assigneeStrategy": "role", "roleCode": "center_director", "approvalType": "any", "departmentScope": true, "departmentVariable": "center_department_id", "requireDepartment": true, "dueDays": 1, "allowReject": true, "allowTransfer": true}}}, {"id": "committee", "type": "approval", "position": {"x": 280, "y": 440}, "data": {"label": "常委会会签", "config": {"assigneeStrategy": "role", "roleCode": "standing_committee", "approvalType": "all", "dueDays": 1, "allowReject": true, "allowTransfer": false}}}, {"id": "end", "type": "end", "position": {"x": 280, "y": 580}, "data": {"label": "结束", "config": {}}}], "edges": [{"id": "start-department_review", "source": "start", "target": "department_review"}, {"id": "department_review-center_review", "source": "department_review", "target": "center_review"}, {"id": "center_review-committee", "source": "center_review", "target": "committee"}, {"id": "committee-end", "source": "committee", "target": "end"}], "viewport": {"x": 0, "y": 0, "zoom": 1}, "charterVersion": 1}$graph$::jsonb,1,NOW(),NOW());
 END IF;
END $migration$;
