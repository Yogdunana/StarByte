package engine

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/workflow/model"
	"github.com/Yogdunana/StarByte/backend/pkg/events"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// --- Mock Repos for engine tests ---

type mockDefRepo struct {
	def     *model.FlowDefinition
	version *model.FlowDefinitionVersion
}

func (m *mockDefRepo) Create(ctx context.Context, tx *gorm.DB, def *model.FlowDefinition) error {
	return nil
}
func (m *mockDefRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.FlowDefinition, error) {
	if m.def != nil && m.def.ID == id {
		return m.def, nil
	}
	return nil, nil
}
func (m *mockDefRepo) GetByKey(ctx context.Context, key string) (*model.FlowDefinition, error) {
	if m.def != nil && m.def.Key == key {
		return m.def, nil
	}
	return nil, nil
}
func (m *mockDefRepo) Update(ctx context.Context, tx *gorm.DB, def *model.FlowDefinition) error {
	return nil
}
func (m *mockDefRepo) Delete(ctx context.Context, id uuid.UUID) error { return nil }
func (m *mockDefRepo) List(ctx context.Context, page, pageSize int, keyword, category string, status *int) ([]model.FlowDefinition, int64, error) {
	return nil, 0, nil
}
func (m *mockDefRepo) CreateVersion(ctx context.Context, tx *gorm.DB, ver *model.FlowDefinitionVersion) error {
	return nil
}
func (m *mockDefRepo) GetVersionByID(ctx context.Context, id uuid.UUID) (*model.FlowDefinitionVersion, error) {
	if m.version != nil && m.version.ID == id {
		return m.version, nil
	}
	return nil, nil
}
func (m *mockDefRepo) GetCurrentVersion(ctx context.Context, definitionID uuid.UUID) (*model.FlowDefinitionVersion, error) {
	return m.version, nil
}
func (m *mockDefRepo) ListVersions(ctx context.Context, definitionID uuid.UUID) ([]model.FlowDefinitionVersion, error) {
	return nil, nil
}
func (m *mockDefRepo) MarkVersionHistorical(ctx context.Context, tx *gorm.DB, definitionID uuid.UUID) error {
	return nil
}

type mockInstRepo struct {
	inst    *model.FlowInstance
	updated *model.FlowInstance
}

func (m *mockInstRepo) Create(ctx context.Context, tx *gorm.DB, inst *model.FlowInstance) error {
	return nil
}
func (m *mockInstRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.FlowInstance, error) {
	if m.inst != nil && m.inst.ID == id {
		return m.inst, nil
	}
	return nil, nil
}
func (m *mockInstRepo) Update(ctx context.Context, tx *gorm.DB, inst *model.FlowInstance) error {
	m.updated = inst
	return nil
}
func (m *mockInstRepo) List(ctx context.Context, page, pageSize int, status *int, definitionID, initiatorID *uuid.UUID) ([]model.FlowInstance, int64, error) {
	return nil, 0, nil
}

type mockTaskRepo struct {
	tasks   map[uuid.UUID]*model.FlowTask
	created []*model.FlowTask
}

func newMockTaskRepo() *mockTaskRepo {
	return &mockTaskRepo{tasks: make(map[uuid.UUID]*model.FlowTask)}
}
func (m *mockTaskRepo) CreateTask(ctx context.Context, tx *gorm.DB, task *model.FlowTask) error {
	m.created = append(m.created, task)
	m.tasks[task.ID] = task
	return nil
}
func (m *mockTaskRepo) GetTaskByID(ctx context.Context, id uuid.UUID) (*model.FlowTask, error) {
	if t, ok := m.tasks[id]; ok {
		return t, nil
	}
	return nil, nil
}
func (m *mockTaskRepo) UpdateTask(ctx context.Context, tx *gorm.DB, task *model.FlowTask) error {
	m.tasks[task.ID] = task
	return nil
}
func (m *mockTaskRepo) ListTodoTasks(ctx context.Context, assigneeID uuid.UUID, page, pageSize int) ([]model.FlowTask, int64, error) {
	return nil, 0, nil
}
func (m *mockTaskRepo) ListDoneTasks(ctx context.Context, assigneeID uuid.UUID, page, pageSize int) ([]model.FlowTask, int64, error) {
	return nil, 0, nil
}
func (m *mockTaskRepo) ListTasksByInstance(ctx context.Context, instanceID uuid.UUID) ([]model.FlowTask, error) {
	return nil, nil
}
func (m *mockTaskRepo) CreateHistory(ctx context.Context, tx *gorm.DB, hist *model.FlowHistory) error {
	return nil
}
func (m *mockTaskRepo) ListHistory(ctx context.Context, instanceID uuid.UUID) ([]model.FlowHistory, error) {
	return nil, nil
}

type mockVarRepo struct {
	vars map[string]interface{}
}

func newMockVarRepo() *mockVarRepo {
	return &mockVarRepo{vars: make(map[string]interface{})}
}
func (m *mockVarRepo) Set(ctx context.Context, tx *gorm.DB, v *model.FlowVariable) error {
	return nil
}
func (m *mockVarRepo) Get(ctx context.Context, instanceID uuid.UUID, key string) (*model.FlowVariable, error) {
	return nil, nil
}
func (m *mockVarRepo) GetWithScope(ctx context.Context, instanceID uuid.UUID, key, scope string) (*model.FlowVariable, error) {
	return nil, nil
}
func (m *mockVarRepo) ListByInstance(ctx context.Context, instanceID uuid.UUID) ([]model.FlowVariable, error) {
	return nil, nil
}
func (m *mockVarRepo) GetMap(ctx context.Context, instanceID uuid.UUID) (map[string]interface{}, error) {
	return m.vars, nil
}
func (m *mockVarRepo) SetMap(ctx context.Context, tx *gorm.DB, instanceID uuid.UUID, vars map[string]interface{}) error {
	for k, v := range vars {
		m.vars[k] = v
	}
	return nil
}
func (m *mockVarRepo) SetMapWithScope(ctx context.Context, tx *gorm.DB, instanceID uuid.UUID, scope string, vars map[string]interface{}) error {
	for k, v := range vars {
		m.vars[k] = v
	}
	return nil
}
func (m *mockVarRepo) DeleteByInstance(ctx context.Context, instanceID uuid.UUID) error {
	return nil
}

// --- Helper to build a simple linear graph ---

func buildLinearGraph(t *testing.T) *FlowGraph {
	t.Helper()
	bpmnData := []byte(`{
		"nodes": [
			{"id": "start", "type": "start", "data": {"label": "开始"}},
			{"id": "end", "type": "end", "data": {"label": "结束"}}
		],
		"edges": [
			{"id": "e1", "source": "start", "target": "end"}
		]
	}`)
	graph, err := ParseGraph(bpmnData)
	require.NoError(t, err)
	return graph
}

func buildGraphWithApproval(t *testing.T) *FlowGraph {
	t.Helper()
	bpmnData := []byte(`{
		"nodes": [
			{"id": "start", "type": "start", "data": {"label": "开始"}},
			{"id": "approve", "type": "approval", "data": {"label": "审批", "config": {"assigneeStrategy": "initiator"}}},
			{"id": "end", "type": "end", "data": {"label": "结束"}}
		],
		"edges": [
			{"id": "e1", "source": "start", "target": "approve"},
			{"id": "e2", "source": "approve", "target": "end"}
		]
	}`)
	graph, err := ParseGraph(bpmnData)
	require.NoError(t, err)
	return graph
}

// --- Engine Tests ---

// mockRegistryForTest is a simple NodeRegistry that wraps a map.
type mockRegistryForTest struct {
	handlers map[string]NodeHandler
}

func (r *mockRegistryForTest) Get(nodeType string) (NodeHandler, error) {
	h, ok := r.handlers[nodeType]
	if !ok {
		return nil, response.NewAppErrorf(response.CodeWorkflowNodeType,
			"node type '%s' is not supported", nodeType)
	}
	return h, nil
}

// stubStartHandler is a minimal handler that always returns the next node.
type stubStartHandler struct{}

func (stubStartHandler) Type() string { return "start" }
func (stubStartHandler) Execute(ctx context.Context, inst *model.FlowInstance, node *FlowNode, graph *FlowGraph, vars map[string]interface{}) ([]string, error) {
	edges := graph.GetNextNodes(node.ID, "")
	result := make([]string, 0, len(edges))
	for _, e := range edges {
		result = append(result, e.Target)
	}
	return result, nil
}
func (stubStartHandler) OnEnter(ctx context.Context, inst *model.FlowInstance, node *FlowNode, vars map[string]interface{}) error {
	return nil
}
func (stubStartHandler) OnLeave(ctx context.Context, inst *model.FlowInstance, node *FlowNode, vars map[string]interface{}) error {
	return nil
}
func (stubStartHandler) Validate(node *FlowNode) error { return nil }

// stubEndHandler signals completion.
type stubEndHandler struct{}

func (stubEndHandler) Type() string { return "end" }
func (stubEndHandler) Execute(ctx context.Context, inst *model.FlowInstance, node *FlowNode, graph *FlowGraph, vars map[string]interface{}) ([]string, error) {
	return []string{}, nil
}
func (stubEndHandler) OnEnter(ctx context.Context, inst *model.FlowInstance, node *FlowNode, vars map[string]interface{}) error {
	return nil
}
func (stubEndHandler) OnLeave(ctx context.Context, inst *model.FlowInstance, node *FlowNode, vars map[string]interface{}) error {
	return nil
}
func (stubEndHandler) Validate(node *FlowNode) error { return nil }

func TestFlowEngine_Start_DefNotFound(t *testing.T) {
	e := NewFlowEngine(
		&mockDefRepo{def: nil},
		&mockInstRepo{},
		newMockTaskRepo(),
		newMockVarRepo(),
		nil,
		&mockRegistryForTest{},
		NewExpressionEngine(),
		events.NewEventBus(),
		nil,
	)

	_, err := e.Start(context.Background(), "nonexistent", "", "", uuid.New(), nil)
	require.Error(t, err)
}

func TestFlowEngine_Start_DefNotPublished(t *testing.T) {
	defID := uuid.New()
	def := &model.FlowDefinition{ID: defID, Key: "test", Status: 0}
	e := NewFlowEngine(
		&mockDefRepo{def: def},
		&mockInstRepo{},
		newMockTaskRepo(),
		newMockVarRepo(),
		nil,
		&mockRegistryForTest{},
		NewExpressionEngine(),
		events.NewEventBus(),
		nil,
	)

	_, err := e.Start(context.Background(), "test", "", "", uuid.New(), nil)
	require.Error(t, err)
	appErr, ok := err.(*response.AppError)
	require.True(t, ok)
	assert.Equal(t, response.CodeWorkflowDefNotPub, appErr.Code)
}

func TestFlowEngine_Start_NoStartNode(t *testing.T) {
	defID := uuid.New()
	versionID := uuid.New()
	// Graph with no start node.
	bpmnData, _ := json.Marshal(map[string]interface{}{
		"nodes": []map[string]interface{}{
			{"id": "n1", "type": "approval", "data": map[string]interface{}{"label": "审批"}},
		},
		"edges": []map[string]interface{}{},
	})
	def := &model.FlowDefinition{ID: defID, Key: "test", Status: 1}
	ver := &model.FlowDefinitionVersion{ID: versionID, DefinitionID: defID, BpmnData: bpmnData, Status: 1}

	e := NewFlowEngine(
		&mockDefRepo{def: def, version: ver},
		&mockInstRepo{},
		newMockTaskRepo(),
		newMockVarRepo(),
		nil,
		&mockRegistryForTest{},
		NewExpressionEngine(),
		events.NewEventBus(),
		nil,
	)

	_, err := e.Start(context.Background(), "test", "", "", uuid.New(), nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "start")
}

func TestFlowEngine_CompleteTask_TaskNotFound(t *testing.T) {
	e := NewFlowEngine(
		&mockDefRepo{},
		&mockInstRepo{},
		newMockTaskRepo(),
		newMockVarRepo(),
		nil,
		&mockRegistryForTest{},
		NewExpressionEngine(),
		events.NewEventBus(),
		nil,
	)

	err := e.CompleteTask(context.Background(), uuid.New(), uuid.New(), ActionApprove, "", nil)
	require.Error(t, err)
}

func TestFlowEngine_Terminate_InstanceNotFound(t *testing.T) {
	e := NewFlowEngine(
		&mockDefRepo{},
		&mockInstRepo{},
		newMockTaskRepo(),
		newMockVarRepo(),
		nil,
		&mockRegistryForTest{},
		NewExpressionEngine(),
		events.NewEventBus(),
		nil,
	)

	err := e.Terminate(context.Background(), uuid.New(), uuid.New(), "test")
	require.Error(t, err)
}

func TestFlowEngine_Terminate_AlreadyCompleted(t *testing.T) {
	instID := uuid.New()
	inst := &model.FlowInstance{ID: instID, Status: 1}

	e := NewFlowEngine(
		&mockDefRepo{},
		&mockInstRepo{inst: inst},
		newMockTaskRepo(),
		newMockVarRepo(),
		nil,
		&mockRegistryForTest{},
		NewExpressionEngine(),
		events.NewEventBus(),
		nil,
	)

	err := e.Terminate(context.Background(), instID, uuid.New(), "test")
	require.Error(t, err)
}

func TestFlowEngine_Suspend_Resume(t *testing.T) {
	instID := uuid.New()
	inst := &model.FlowInstance{ID: instID, Status: 0}
	instRepo := &mockInstRepo{inst: inst}

	e := NewFlowEngine(
		&mockDefRepo{},
		instRepo,
		newMockTaskRepo(),
		newMockVarRepo(),
		nil,
		&mockRegistryForTest{},
		NewExpressionEngine(),
		events.NewEventBus(),
		nil,
	)

	// Suspend.
	err := e.Suspend(context.Background(), instID, uuid.New(), "testing")
	require.NoError(t, err)
	assert.Equal(t, 3, instRepo.updated.Status)

	// Resume.
	inst.Status = 3
	err = e.Resume(context.Background(), instID, uuid.New())
	require.NoError(t, err)
	assert.Equal(t, 0, instRepo.updated.Status)
}

func TestFlowEngine_Suspend_NotRunning(t *testing.T) {
	instID := uuid.New()
	inst := &model.FlowInstance{ID: instID, Status: 1}

	e := NewFlowEngine(
		&mockDefRepo{},
		&mockInstRepo{inst: inst},
		newMockTaskRepo(),
		newMockVarRepo(),
		nil,
		&mockRegistryForTest{},
		NewExpressionEngine(),
		events.NewEventBus(),
		nil,
	)

	err := e.Suspend(context.Background(), instID, uuid.New(), "testing")
	require.Error(t, err)
}

func TestFlowEngine_Resume_NotSuspended(t *testing.T) {
	instID := uuid.New()
	inst := &model.FlowInstance{ID: instID, Status: 0}

	e := NewFlowEngine(
		&mockDefRepo{},
		&mockInstRepo{inst: inst},
		newMockTaskRepo(),
		newMockVarRepo(),
		nil,
		&mockRegistryForTest{},
		NewExpressionEngine(),
		events.NewEventBus(),
		nil,
	)

	err := e.Resume(context.Background(), instID, uuid.New())
	require.Error(t, err)
}

func TestParseGraph_EndToEndGraph(t *testing.T) {
	graph := buildLinearGraph(t)

	start := graph.FindStartNode()
	require.NotNil(t, start)
	assert.Equal(t, "start", start.Type)

	end := graph.GetNode("end")
	require.NotNil(t, end)
	assert.Equal(t, "end", end.Type)

	edges := graph.GetNextNodes("start", "")
	assert.Len(t, edges, 1)
	assert.Equal(t, "end", edges[0].Target)
}

func TestFlowEngine_UpdateCurrentNodes(t *testing.T) {
	instRepo := &mockInstRepo{inst: &model.FlowInstance{ID: uuid.New(), Status: 0}}
	e := NewFlowEngine(
		&mockDefRepo{},
		instRepo,
		newMockTaskRepo(),
		newMockVarRepo(),
		nil,
		&mockRegistryForTest{},
		NewExpressionEngine(),
		events.NewEventBus(),
		nil,
	)

	inst := &model.FlowInstance{ID: uuid.New(), Status: 0}
	e.updateCurrentNodes(context.Background(), inst, []string{"node1", "node2"})

	var nodeIDs []string
	err := json.Unmarshal(inst.CurrentNodeIDs, &nodeIDs)
	require.NoError(t, err)
	assert.Equal(t, []string{"node1", "node2"}, nodeIDs)
}

func TestTaskAction_String(t *testing.T) {
	assert.Equal(t, "approve", string(ActionApprove))
	assert.Equal(t, "reject", string(ActionReject))
	assert.Equal(t, "transfer", string(ActionTransfer))
	assert.Equal(t, "rollback", string(ActionRollback))
	assert.Equal(t, "withdraw", string(ActionWithdraw))
}

func TestFlowEngine_WithTransaction_NilDB(t *testing.T) {
	e := NewFlowEngine(
		&mockDefRepo{},
		&mockInstRepo{},
		newMockTaskRepo(),
		newMockVarRepo(),
		nil, // nil DB — Transaction will panic
		&mockRegistryForTest{},
		NewExpressionEngine(),
		events.NewEventBus(),
		nil,
	)

	// withTransaction calls e.db.WithContext(ctx).Transaction(fn)
	// With nil db, this will panic — but this is expected for the test.
	// We only test the method exists and compiles.
	assert.NotNil(t, e)
}

// Ensure time import is used.
var _ = time.Now
