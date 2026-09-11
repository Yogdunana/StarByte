package engine

import (
	"fmt"
	"strings"
)

const runtimePrefix = "__workflow_"

func validateInputVariables(variables map[string]interface{}) error {
	for key := range variables {
		if strings.HasPrefix(key, runtimePrefix) {
			return fmt.Errorf("变量名 %s 为流程引擎保留名称", key)
		}
	}
	return nil
}

// arriveParallelJoin persists one token per incoming predecessor. A join releases
// only when every branch has arrived; remaining tokens belong to a later cycle.
func arriveParallelJoin(graph *FlowGraph, node *FlowNode, source string, vars map[string]interface{}) (bool, error) {
	if node.Type != "parallel_gateway" {
		return true, nil
	}
	predecessors := map[string]bool{}
	for _, edge := range graph.Edges {
		if edge.Target == node.ID {
			predecessors[edge.Source] = true
		}
	}
	if len(predecessors) < 2 {
		return true, nil
	}
	if !predecessors[source] {
		return false, fmt.Errorf("并行汇合节点 %s 缺少有效入线来源", node.ID)
	}
	key := runtimePrefix + "join_" + node.ID
	tokens, ok := vars[key].(map[string]interface{})
	if !ok {
		tokens = map[string]interface{}{}
	}
	count, _ := tokens[source].(float64)
	tokens[source] = count + 1
	vars[key] = tokens
	for id := range predecessors {
		count, _ := tokens[id].(float64)
		if count < 1 {
			return false, nil
		}
	}
	for id := range predecessors {
		count, _ := tokens[id].(float64)
		tokens[id] = count - 1
	}
	return true, nil
}
