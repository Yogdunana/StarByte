package engine

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// MemberApplicationDefinitionKey is the default membership approval template (#64).
const MemberApplicationDefinitionKey = "member_application"

// SkipMinisterVariable tells Start to pass through the minister node
// without creating assignees when a member application has no department.
const SkipMinisterVariable = "skip_minister"

// SkipOfficerVariable skips the officer node for member-type applications.
const SkipOfficerVariable = "skip_officer"

var roleCodePattern = regexp.MustCompile(`^[a-z][a-z0-9_]{1,32}$`)

func truthyVar(vars map[string]interface{}, key string) bool {
	if vars == nil || key == "" {
		return false
	}
	switch v := vars[key].(type) {
	case bool:
		return v
	case string:
		return v == "true" || v == "1"
	default:
		return false
	}
}

func skipWhenKey(node *FlowNode) string {
	if node == nil || (node.Type != "" && node.Type != "approval") {
		return ""
	}
	if node.Config != nil {
		if key, _ := node.Config["skipWhen"].(string); key != "" {
			return key
		}
	}
	switch node.ID {
	case "officer":
		return SkipOfficerVariable
	case "minister":
		return SkipMinisterVariable
	default:
		return ""
	}
}

func skipIfEmptyNode(node *FlowNode) bool {
	if node == nil || node.Config == nil {
		return false
	}
	switch v := node.Config["skipIfEmpty"].(type) {
	case bool:
		return v
	case string:
		return v == "true" || v == "1"
	default:
		return false
	}
}

// skipApplicationApproval reports whether an approval node should be passed
// through without creating tasks (condition variable or legacy skip flags).
func skipApplicationApproval(node *FlowNode, vars map[string]interface{}) bool {
	return truthyVar(vars, skipWhenKey(node))
}

// skipMinisterApproval is the phase-1 helper; kept for existing tests.
func skipMinisterApproval(node *FlowNode, vars map[string]interface{}) bool {
	return skipApplicationApproval(node, vars) && node != nil && node.ID == "minister"
}

func approvalRoleCode(node *FlowNode) string {
	if node == nil || node.Config == nil {
		return ""
	}
	code, _ := node.Config["roleCode"].(string)
	return strings.TrimSpace(code)
}

func approvalTypeOf(node *FlowNode) string {
	if node == nil || node.Config == nil {
		return ""
	}
	raw, _ := node.Config["approvalType"].(string)
	if raw == "" {
		return "any"
	}
	return raw
}

func nodeAllowsTransfer(node *FlowNode) bool {
	if node == nil || node.Config == nil {
		return true
	}
	switch v := node.Config["allowTransfer"].(type) {
	case bool:
		return v
	case string:
		return v != "false" && v != "0"
	default:
		return true
	}
}

// MemberApplicationBPMN is the React Flow graph seeded by 000061 / 000069.
// Default spine: 干事审批 → 部长审批 → 社长审批. Extra nodes are allowed at publish time.
func MemberApplicationBPMN() []byte {
	raw, err := json.Marshal(map[string]interface{}{
		"nodes": []map[string]interface{}{
			node("start", "start", "提交申请", 20, nil),
			node("officer", "approval", "干事审批", 160, map[string]interface{}{
				"assigneeStrategy": "role", "roleCode": "officer", "approvalType": "any",
				"departmentScope": true, "allowReject": true, "allowTransfer": true, "allowRollback": false,
				"skipWhen": SkipOfficerVariable, "skipIfEmpty": true,
			}),
			node("minister", "approval", "部长审批", 300, map[string]interface{}{
				"assigneeStrategy": "role", "roleCode": "minister", "approvalType": "any",
				"departmentScope": true, "allowReject": true, "allowTransfer": true, "allowRollback": false,
				"skipWhen": SkipMinisterVariable, "skipIfEmpty": true,
			}),
			node("president", "approval", "社长审批", 440, map[string]interface{}{
				"assigneeStrategy": "role", "roleCode": "president", "approvalType": "any",
				"allowReject": true, "allowTransfer": true, "allowRollback": false,
			}),
			node("end", "end", "结束", 580, nil),
		},
		"edges": []map[string]string{
			{"id": "start-officer", "source": "start", "target": "officer"},
			{"id": "officer-minister", "source": "officer", "target": "minister"},
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

func node(id, kind, label string, y int, config map[string]interface{}) map[string]interface{} {
	if config == nil {
		config = map[string]interface{}{}
	}
	return map[string]interface{}{
		"id": id, "type": kind, "position": map[string]int{"x": 280, "y": y},
		"data": map[string]interface{}{"label": label, "config": config},
	}
}

func validateMemberApplicationGraph(g *FlowGraph) error {
	if g == nil || len(g.Nodes) == 0 {
		return fmt.Errorf("入会申请流程不能为空")
	}
	starts, ends, approvals := 0, 0, 0
	for _, item := range g.Nodes {
		switch item.Type {
		case "start":
			starts++
		case "end":
			ends++
		case "approval":
			approvals++
			code := approvalRoleCode(item)
			if item.Config["assigneeStrategy"] != "role" || !roleCodePattern.MatchString(code) {
				return fmt.Errorf("入会审批节点须按角色指派（roleCode）")
			}
			switch approvalTypeOf(item) {
			case "any", "all", "single", "ratio":
			default:
				return fmt.Errorf("入会审批仅支持或签/会签/单人/比例")
			}
		case "condition", "exclusive_gateway", "parallel_gateway", "parallel", "merge", "notify":
		default:
			return fmt.Errorf("入会申请不支持节点类型：%s", item.Type)
		}
	}
	if starts != 1 || ends < 1 {
		return fmt.Errorf("入会申请须有且仅有一个开始节点，以及至少一个结束节点")
	}
	if approvals < 1 {
		return fmt.Errorf("入会申请至少需要一个审批节点")
	}
	start := g.FindStartNode()
	if start == nil {
		return fmt.Errorf("入会申请缺少开始节点")
	}
	if !memberApplicationReachesEnd(g, start.ID) {
		return fmt.Errorf("入会申请必须能从开始节点到达结束")
	}
	return nil
}

func memberApplicationReachesEnd(g *FlowGraph, startID string) bool {
	seen := map[string]bool{}
	queue := []string{startID}
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		if seen[id] {
			continue
		}
		seen[id] = true
		node := g.GetNode(id)
		if node != nil && node.Type == "end" {
			return true
		}
		for _, edge := range g.GetNextNodes(id, "") {
			queue = append(queue, edge.Target)
		}
	}
	return false
}
