package engine

import (
	"fmt"
	"strings"
)

// ValidateGraph rejects definitions that would strand tasks or silently skip routes.
func ValidateGraph(graph *FlowGraph, registry NodeRegistry) error {
	if graph == nil || len(graph.Nodes) == 0 {
		return fmt.Errorf("流程不能为空")
	}
	incoming, outgoing := map[string]int{}, map[string]int{}
	starts, ends := []string{}, []string{}
	for id, node := range graph.Nodes {
		if strings.TrimSpace(id) == "" {
			return fmt.Errorf("节点 ID 不能为空")
		}
		handler, err := registry.Get(node.Type)
		if err != nil {
			return err
		}
		if err := handler.Validate(node); err != nil {
			return err
		}
		if node.Type == "start" {
			starts = append(starts, id)
		}
		if node.Type == "end" {
			ends = append(ends, id)
		}
	}
	if len(starts) != 1 || len(ends) == 0 {
		return fmt.Errorf("流程必须有且只有一个开始节点，并至少有一个结束节点")
	}
	seen := map[string]bool{}
	for _, edge := range graph.Edges {
		if graph.GetNode(edge.Source) == nil || graph.GetNode(edge.Target) == nil {
			return fmt.Errorf("连线引用了不存在的节点")
		}
		key := edge.Source + "\x00" + edge.SourceHandle + "\x00" + edge.Target
		if seen[key] {
			return fmt.Errorf("不允许重复连线")
		}
		seen[key] = true
		outgoing[edge.Source]++
		incoming[edge.Target]++
	}
	for id, node := range graph.Nodes {
		if node.Type == "start" && incoming[id] != 0 {
			return fmt.Errorf("开始节点不能有入线")
		}
		if node.Type == "end" {
			if outgoing[id] != 0 {
				return fmt.Errorf("结束节点不能有出线")
			}
		} else if outgoing[id] == 0 {
			return fmt.Errorf("节点 %s 缺少出线", node.Label)
		}
		if node.Type != "parallel_gateway" && node.Type != "exclusive_gateway" && outgoing[id] > 1 {
			return fmt.Errorf("多分支必须使用网关节点")
		}
		if node.Type == "exclusive_gateway" {
			branches, err := ParseBranches(node.Config)
			if err != nil {
				return err
			}
			handles := map[string]bool{}
			for _, branch := range branches {
				if branch.ID == "" || handles[branch.ID] {
					return fmt.Errorf("条件分支 ID 必须唯一且非空")
				}
				handles[branch.ID] = true
				if len(graph.GetNextNodes(id, branch.ID)) != 1 {
					return fmt.Errorf("每个条件分支必须连接一个目标")
				}
			}
			for _, edge := range graph.GetNextNodes(id, "") {
				if !handles[edge.SourceHandle] {
					return fmt.Errorf("条件网关连线必须对应有效分支")
				}
			}
		}
	}
	reachable := map[string]bool{}
	var visit func(string)
	visit = func(id string) {
		if reachable[id] {
			return
		}
		reachable[id] = true
		for _, edge := range graph.GetNextNodes(id, "") {
			visit(edge.Target)
		}
	}
	visit(starts[0])
	if len(reachable) != len(graph.Nodes) {
		return fmt.Errorf("存在无法从开始节点到达的孤立节点")
	}
	reverse := map[string]bool{}
	var upstream func(string)
	upstream = func(id string) {
		if reverse[id] {
			return
		}
		reverse[id] = true
		for _, edge := range graph.Edges {
			if edge.Target == id {
				upstream(edge.Source)
			}
		}
	}
	for _, end := range ends {
		upstream(end)
	}
	if len(reverse) != len(graph.Nodes) {
		return fmt.Errorf("存在无法到达结束节点的路径")
	}
	// Cycles with a human wait may be intentional; automatic-only cycles cannot run.
	state := map[string]int{}
	var automaticCycle func(string) bool
	automaticCycle = func(id string) bool {
		handler, _ := registry.Get(graph.GetNode(id).Type)
		if wait, ok := handler.(WaitingNode); ok && wait.WaitForCompletion() {
			return false
		}
		if state[id] == 1 {
			return true
		}
		if state[id] == 2 {
			return false
		}
		state[id] = 1
		for _, edge := range graph.GetNextNodes(id, "") {
			if automaticCycle(edge.Target) {
				return true
			}
		}
		state[id] = 2
		return false
	}
	for id := range graph.Nodes {
		if automaticCycle(id) {
			return fmt.Errorf("自动节点不能形成无限循环")
		}
	}
	return nil
}
