package engine

import (
	"encoding/json"
	"fmt"
)

const TaskTransferBusinessType = "task_transfer"

func IsTaskTransferDefinition(key string) bool {
	return key == "task_transfer_internal" || key == "task_transfer_department" || key == "task_transfer_center"
}

func TaskTransferDefinitionKey(kind string) string {
	switch kind {
	case "internal":
		return "task_transfer_internal"
	case "department":
		return "task_transfer_department"
	case "center":
		return "task_transfer_center"
	}
	return ""
}

func TaskTransferBPMN(kind string) []byte {
	roles := TaskTransferRoles(kind)
	if len(roles) == 0 {
		return []byte(`{"nodes":[],"edges":[]}`)
	}
	approval := func(id string, x, y int) map[string]interface{} {
		return map[string]interface{}{
			"id": id, "type": "approval", "position": map[string]int{"x": x, "y": y},
			"data": map[string]interface{}{"label": id, "config": map[string]interface{}{
				"assigneeStrategy": "business_role", "businessType": TaskTransferBusinessType,
				"transferStage": "handover", "transferRole": id, "approvalType": "any",
				"dueDays": 1, "allowReject": true, "allowTransfer": false, "allowRollback": false,
			}},
		}
	}
	gateway := func(id string, y int) map[string]interface{} {
		return map[string]interface{}{
			"id": id, "type": "parallel_gateway", "position": map[string]int{"x": 280, "y": y},
			"data": map[string]interface{}{"label": id, "config": map[string]interface{}{}},
		}
	}
	nodes := []map[string]interface{}{node("start", "start", "申请转办", 0, nil)}
	edges := []map[string]string{}
	if len(roles) == 1 {
		nodes = append(nodes, approval(roles[0], 280, 160), node("end", "end", "转办完成", 320, nil))
		edges = append(edges,
			map[string]string{"id": "start-" + roles[0], "source": "start", "target": roles[0]},
			map[string]string{"id": roles[0] + "-end", "source": roles[0], "target": "end"},
		)
	} else {
		nodes = append(nodes, gateway("fork", 140))
		for i, role := range roles {
			nodes = append(nodes, approval(role, 80+i*200, 280))
			edges = append(edges,
				map[string]string{"id": "fork-" + role, "source": "fork", "target": role},
				map[string]string{"id": role + "-join", "source": role, "target": "join"},
			)
		}
		nodes = append(nodes, gateway("join", 420), node("end", "end", "转办完成", 560, nil))
		edges = append(edges,
			map[string]string{"id": "start-fork", "source": "start", "target": "fork"},
			map[string]string{"id": "join-end", "source": "join", "target": "end"},
		)
	}
	raw, err := json.Marshal(map[string]interface{}{"nodes": nodes, "edges": edges, "viewport": map[string]interface{}{"x": 0, "y": 0, "zoom": 1}})
	if err != nil {
		return []byte(`{"nodes":[],"edges":[]}`)
	}
	return raw
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
