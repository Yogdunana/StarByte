package engine

import "fmt"

const TaskTransferBusinessType = "task_transfer"

func IsTaskTransferDefinition(key string) bool {
	return key == "task_transfer_internal" || key == "task_transfer_department" || key == "task_transfer_center"
}
func TaskTransferRoles(kind string) []string {
	switch kind {
	case "internal":
		return []string{"supervisor"}
	case "department":
		return []string{"source_minister", "target_minister"}
	case "center":
		return []string{"source_minister", "target_minister", "source_center", "target_center"}
	}
	return nil
}
func ValidateTaskTransferDefinition(key string, g *FlowGraph) error {
	kind := ""
	switch key {
	case "task_transfer_internal":
		kind = "internal"
	case "task_transfer_department":
		kind = "department"
	case "task_transfer_center":
		kind = "center"
	}
	roles := TaskTransferRoles(kind)
	if len(roles) == 0 {
		return fmt.Errorf("未知的任务转办流程")
	}
	semantic := map[string]string{}
	seen := map[string]bool{}
	for id, node := range g.Nodes {
		value := node.Type
		switch node.Type {
		case "start", "end":
		case "parallel_gateway":
			incoming, outgoing := 0, 0
			for _, e := range g.Edges {
				if e.Source == id {
					outgoing++
				}
				if e.Target == id {
					incoming++
				}
			}
			if incoming == 1 && outgoing == len(roles) {
				value = "fork"
			} else if incoming == len(roles) && outgoing == 1 {
				value = "join"
			} else {
				return fmt.Errorf("转办会签网关分支数不正确")
			}
		case "approval":
			value, _ = node.Config["transferRole"].(string)
			if node.Config["businessType"] != TaskTransferBusinessType || node.Config["transferStage"] != "handover" || node.Config["assigneeStrategy"] != "business_role" || node.Config["approvalType"] != "any" {
				return fmt.Errorf("转办必须由对应负责人的真实签字完成")
			}
		default:
			return fmt.Errorf("转办流程不能加入自动跳过签字的节点")
		}
		if value == "" || seen[value] {
			return fmt.Errorf("转办签字环节重复或缺失")
		}
		seen[value] = true
		semantic[id] = value
	}
	expected := map[string]bool{}
	if len(roles) == 1 {
		expected["start>supervisor"] = true
		expected["supervisor>end"] = true
	} else {
		expected["start>fork"] = true
		expected["join>end"] = true
		for _, role := range roles {
			expected["fork>"+role] = true
			expected[role+">join"] = true
		}
	}
	nodes := len(roles) + 2
	if len(roles) > 1 {
		nodes += 2
	}
	if len(g.Nodes) != nodes || len(g.Edges) != len(expected) {
		return fmt.Errorf("转办流程必须保留全部签字节点")
	}
	for _, e := range g.Edges {
		edge := semantic[e.Source] + ">" + semantic[e.Target]
		if !expected[edge] || e.SourceHandle != "" {
			return fmt.Errorf("转办流程不能删减或跳过签字")
		}
		delete(expected, edge)
	}
	if len(expected) != 0 {
		return fmt.Errorf("转办流程连接不完整")
	}
	return nil
}
