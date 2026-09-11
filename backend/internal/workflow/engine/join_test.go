package engine

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParallelJoinRetainsArrivalsAcrossPersistence(t *testing.T) {
	join := &FlowNode{ID: "join", Type: "parallel_gateway"}
	graph := &FlowGraph{Edges: []*FlowEdge{{Source: "a", Target: "join"}, {Source: "b", Target: "join"}}}
	variables := map[string]interface{}{}
	ready, err := arriveParallelJoin(graph, join, "a", variables)
	require.NoError(t, err)
	require.False(t, ready)
	encoded, err := json.Marshal(variables)
	require.NoError(t, err)
	restored := map[string]interface{}{}
	require.NoError(t, json.Unmarshal(encoded, &restored))
	ready, err = arriveParallelJoin(graph, join, "b", restored)
	require.NoError(t, err)
	require.True(t, ready)
	ready, err = arriveParallelJoin(graph, join, "b", restored)
	require.NoError(t, err)
	require.False(t, ready, "previous cycle arrivals must be consumed")
	_, err = arriveParallelJoin(graph, join, "unrelated", restored)
	require.Error(t, err)
}
func TestClientsCannotSupplyEngineStateVariables(t *testing.T) {
	require.Error(t, validateInputVariables(map[string]interface{}{runtimePrefix + "join_x": true}))
	require.NoError(t, validateInputVariables(map[string]interface{}{"amount": 100}))
}
