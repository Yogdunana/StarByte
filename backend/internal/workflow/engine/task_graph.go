package engine

import "fmt"

const TaskBusinessType = "collaboration_task"
const TaskDefinitionKey = "task_lifecycle"

func IsProtectedBusiness(kind string) bool {
	return kind == "member_application" || kind == TaskBusinessType || kind == TaskTransferBusinessType
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
