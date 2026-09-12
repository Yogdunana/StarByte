package engine

import (
	"encoding/json"
	"fmt"
)

const TaskBusinessType = "collaboration_task"
const TaskDefinitionKey = "task_lifecycle"

func IsProtectedBusiness(kind string) bool {
	return kind == "member_application" || kind == TaskBusinessType || kind == TaskTransferBusinessType
}

// TaskLifecycleBPMN is the default published graph seeded by 000046 / 000062.
// 发布 → 认领/分配 → 执行 → 审核 → 验收 → 完成
func TaskLifecycleBPMN() []byte {
	stage := func(id, label string, y int) map[string]interface{} {
		return node(id, "approval", label, y, map[string]interface{}{
			"assigneeStrategy": "business_role", "businessType": TaskBusinessType,
			"taskStage": id, "approvalType": "single", "dueDays": 1,
			"allowReject": true, "allowTransfer": false, "allowRollback": false,
		})
	}
	raw, err := json.Marshal(map[string]interface{}{
		"nodes": []map[string]interface{}{
			node("start", "start", "发布任务", 0, nil),
			stage("assignment", "认领 / 分配", 140),
			stage("execution", "执行与提交交付", 280),
			stage("review", "负责人审核", 420),
			stage("acceptance", "正式验收", 560),
			node("end", "end", "任务完成", 700, nil),
		},
		"edges": []map[string]string{
			{"id": "start-assignment", "source": "start", "target": "assignment"},
			{"id": "assignment-execution", "source": "assignment", "target": "execution"},
			{"id": "execution-review", "source": "execution", "target": "review"},
			{"id": "review-acceptance", "source": "review", "target": "acceptance"},
			{"id": "acceptance-end", "source": "acceptance", "target": "end"},
		},
		"viewport": map[string]interface{}{"x": 0, "y": 0, "zoom": 1},
	})
	if err != nil {
		return []byte(`{"nodes":[],"edges":[]}`)
	}
	return raw
}

// The lifecycle checkpoints cannot be removed by changing layout or assignees.
// Return-for-rework is an explicit, audited runtime action, never an auto edge.
func ValidateTaskDefinition(g *FlowGraph) error {
	stages := []string{"start", "assignment", "execution", "review", "acceptance", "end"}
	if len(g.Nodes) != len(stages) || len(g.Edges) != len(stages)-1 {
		return fmt.Errorf("任务流程须保留发布、分配、执行、审核、验收和完成")
	}
	ids := map[string]string{}
	for id, node := range g.Nodes {
		stage := node.Type
		if node.Type == "approval" {
			stage, _ = node.Config["taskStage"].(string)
			if node.Config["businessType"] != TaskBusinessType || node.Config["assigneeStrategy"] != "business_role" || node.Config["approvalType"] != "single" {
				return fmt.Errorf("任务环节须由业务指定的负责人处理")
			}
		} else if node.Type != "start" && node.Type != "end" {
			return fmt.Errorf("任务必经环节不允许自动跳过")
		}
		if ids[stage] != "" {
			return fmt.Errorf("任务环节重复：%s", stage)
		}
		ids[stage] = id
	}
	for i, stage := range stages {
		if ids[stage] == "" {
			return fmt.Errorf("任务环节缺失：%s", stage)
		}
		if i == len(stages)-1 {
			break
		}
		found := 0
		for _, edge := range g.Edges {
			if edge.Source == ids[stage] && edge.Target == ids[stages[i+1]] && edge.SourceHandle == "" {
				found++
			}
		}
		if found != 1 {
			return fmt.Errorf("任务必须按分配、执行、审核、验收顺序完成")
		}
	}
	return nil
}
