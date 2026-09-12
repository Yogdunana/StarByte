package engine

import (
	"fmt"
	"strings"
)

// ValidateBusinessDefinition protects mandatory signatures while allowing
// editable labels/layout and sequential or parallel first-round countersigning.
// Generic process templates keep their full graph freedom.
func ValidateBusinessDefinition(key string, g *FlowGraph) error {
	if IsTaskTransferDefinition(key) {
		return ValidateTaskTransferDefinition(key, g)
	}
	if key == TaskDefinitionKey {
		return ValidateTaskDefinition(g)
	}
	if key == MemberApplicationDefinitionKey {
		return validateMemberApplicationGraph(g)
	}
	if key != "officer_interview" && key != "member_admission" {
		return nil
	}
	semantic := map[string]string{}
	seen := map[string]bool{}
	for id, node := range g.Nodes {
		name := node.Type
		switch node.Type {
		case "start", "end":
		case "parallel_gateway":
			incoming, outgoing := 0, 0
			for _, edge := range g.Edges {
				if edge.Source == id {
					outgoing++
				}
				if edge.Target == id {
					incoming++
				}
			}
			if incoming == 1 && outgoing == 2 {
				name = "fork"
			} else if incoming == 2 && outgoing == 1 {
				name = "join"
			} else {
				return fmt.Errorf("入会会签网关必须为二分支或二汇一")
			}
		case "approval":
			stage, _ := node.Config["admissionStage"].(string)
			role, _ := node.Config["admissionRole"].(string)
			if node.Config["businessType"] != "member_application" || node.Config["assigneeStrategy"] != "business_role" || node.Config["approvalType"] != "any" {
				return fmt.Errorf("入会审批必须由对应职务的真实签字人处理")
			}
			name = stage + ":" + role
		default:
			return fmt.Errorf("入会必经流程不允许添加可绕过签字的自动节点")
		}
		if seen[name] {
			return fmt.Errorf("入会环节重复：%s", name)
		}
		seen[name] = true
		semantic[id] = name
	}
	patterns := [][]string{{"start>materials:materials", "materials:materials>end"}}
	if key == "officer_interview" {
		common := []string{"start>materials:materials", "round2:center>president:president", "president:president>end"}
		patterns = [][]string{
			append(append([]string{}, common...), "materials:materials>fork", "fork>round1:minister", "fork>round1:center", "round1:minister>join", "round1:center>join", "join>round2:center"),
			append(append([]string{}, common...), "materials:materials>round1:minister", "round1:minister>round1:center", "round1:center>round2:center"),
			append(append([]string{}, common...), "materials:materials>round1:center", "round1:center>round1:minister", "round1:minister>round2:center"),
		}
	}
	actual := map[string]bool{}
	for _, edge := range g.Edges {
		actual[semantic[edge.Source]+">"+semantic[edge.Target]] = true
	}
	for _, pattern := range patterns {
		if len(actual) != len(pattern) || len(g.Edges) != len(pattern) {
			continue
		}
		valid := true
		expectedNodes := map[string]bool{}
		for _, edge := range pattern {
			if !actual[edge] {
				valid = false
			}
			for _, node := range strings.Split(edge, ">") {
				expectedNodes[node] = true
			}
		}
		if valid && len(expectedNodes) == len(g.Nodes) {
			return nil
		}
	}
	return fmt.Errorf("流程必须保留资料审核；干事还须一面部长与中心会签、二面中心签字、会长最终确认，不能删减或跳级")
}
