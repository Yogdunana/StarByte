package engine

import (
	"encoding/json"
	"fmt"
	"strings"
)

const LeaveBusinessType = "leave_application"
const LeaveDefinitionKey = "leave_approval"

// LeaveApprovalBPMN is the default published graph seeded by 000067.
// 申请 → 部长审批 → 社长审批 → 结束
func LeaveApprovalBPMN() []byte {
	raw, err := json.Marshal(map[string]interface{}{
		"nodes": []map[string]interface{}{
			node("start", "start", "提交请假", 20, nil),
			node("minister", "approval", "部长审批", 160, map[string]interface{}{
				"assigneeStrategy": "role", "roleCode": "minister", "approvalType": "any",
				"departmentScope": true, "leaveStage": "department",
				"allowReject": true, "allowTransfer": false, "allowRollback": false,
			}),
			node("president", "approval", "社长审批", 300, map[string]interface{}{
				"assigneeStrategy": "role", "roleCode": "president", "approvalType": "any",
				"leaveStage":  "org",
				"allowReject": true, "allowTransfer": false, "allowRollback": false,
			}),
			node("end", "end", "结束", 440, nil),
		},
		"edges": []map[string]string{
			{"id": "start-minister", "source": "start", "target": "minister"},
			{"id": "minister-president", "source": "minister", "target": "president"},
			{"id": "president-end", "source": "president", "target": "end"},
		},
		"viewport": map[string]interface{}{"x": 0, "y": 0, "zoom": 1},
	})
	if err != nil {
		return []byte(`{"nodes":[],"edges":[]}`)
	}
	return raw
}

func validateLeaveApprovalGraph(g *FlowGraph) error {
	semantic := map[string]string{}
	seen := map[string]bool{}
	for id, item := range g.Nodes {
		name := item.Type
		switch item.Type {
		case "start", "end":
		case "approval":
			code, _ := item.Config["roleCode"].(string)
			if item.Config["assigneeStrategy"] != "role" || (code != "minister" && code != "president") || item.Config["approvalType"] != "any" {
				return fmt.Errorf("请假审批节点须按角色指派部长或社长")
			}
			name = code
		default:
			return fmt.Errorf("默认请假链不支持额外节点类型")
		}
		if seen[name] {
			return fmt.Errorf("请假审批环节重复：%s", name)
		}
		seen[name] = true
		semantic[id] = name
	}
	actual := map[string]bool{}
	for _, edge := range g.Edges {
		actual[semantic[edge.Source]+">"+semantic[edge.Target]] = true
	}
	pattern := []string{"start>minister", "minister>president", "president>end"}
	if len(actual) != len(pattern) || len(g.Edges) != len(pattern) || len(g.Nodes) != 4 {
		return fmt.Errorf("默认请假链必须为：提交请假 → 部长审批 → 社长审批 → 结束")
	}
	for _, edge := range pattern {
		if !actual[edge] {
			return fmt.Errorf("默认请假链必须为：提交请假 → 部长审批 → 社长审批 → 结束")
		}
		for _, part := range strings.Split(edge, ">") {
			if !seen[part] {
				return fmt.Errorf("默认请假链缺少环节：%s", part)
			}
		}
	}
	return nil
}
