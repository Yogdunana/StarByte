package repo

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/Yogdunana/StarByte/backend/internal/workflow/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// setupTestDB creates a test database connection.
// Set WORKFLOW_TEST_DSN env var to enable DB tests; otherwise tests are skipped.
func setupTestDB(t *testing.T) *gorm.DB {
	dsn := "host=localhost user=postgres password=postgres dbname=starbyte_test port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Skipf("skipping DB test: %v", err)
	}

	// Auto-migrate workflow tables
	if err := db.AutoMigrate(
		&model.FlowDefinition{},
		&model.FlowDefinitionVersion{},
		&model.FlowInstance{},
		&model.FlowTask{},
		&model.FlowHistory{},
		&model.FlowVariable{},
	); err != nil {
		t.Skipf("skipping DB test (migrate failed): %v", err)
	}

	// Clean up tables before each test
	db.Exec("TRUNCATE flow_variables, flow_histories, flow_tasks, flow_instances, flow_definition_versions, flow_definitions CASCADE")

	return db
}

// ========== DefinitionRepo Tests ==========

func TestDefinitionRepo_CreateAndGet(t *testing.T) {
	db := setupTestDB(t)
	repo := NewDefinitionRepo(db)
	ctx := context.Background()

	def := &model.FlowDefinition{
		ID:       uuid.New(),
		Key:      "test-flow-001",
		Name:     "Test Flow",
		Category: "test",
		Status:   0,
	}

	err := repo.Create(ctx, nil, def)
	require.NoError(t, err)

	// Get by ID
	got, err := repo.GetByID(ctx, def.ID)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "test-flow-001", got.Key)
	assert.Equal(t, "Test Flow", got.Name)

	// Get by Key
	gotByKey, err := repo.GetByKey(ctx, "test-flow-001")
	require.NoError(t, err)
	require.NotNil(t, gotByKey)
	assert.Equal(t, def.ID, gotByKey.ID)
}

func TestDefinitionRepo_GetByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewDefinitionRepo(db)
	ctx := context.Background()

	got, err := repo.GetByID(ctx, uuid.New())
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestDefinitionRepo_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := NewDefinitionRepo(db)
	ctx := context.Background()

	def := &model.FlowDefinition{
		ID:       uuid.New(),
		Key:      "test-flow-update",
		Name:     "Before Update",
		Category: "test",
	}
	require.NoError(t, repo.Create(ctx, nil, def))

	def.Name = "After Update"
	require.NoError(t, repo.Update(ctx, nil, def))

	got, err := repo.GetByID(ctx, def.ID)
	require.NoError(t, err)
	assert.Equal(t, "After Update", got.Name)
}

func TestDefinitionRepo_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := NewDefinitionRepo(db)
	ctx := context.Background()

	def := &model.FlowDefinition{
		ID:   uuid.New(),
		Key:  "test-flow-delete",
		Name: "To Delete",
	}
	require.NoError(t, repo.Create(ctx, nil, def))

	require.NoError(t, repo.Delete(ctx, def.ID))

	got, err := repo.GetByID(ctx, def.ID)
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestDefinitionRepo_List(t *testing.T) {
	db := setupTestDB(t)
	repo := NewDefinitionRepo(db)
	ctx := context.Background()

	// Create multiple definitions
	for i := 0; i < 5; i++ {
		def := &model.FlowDefinition{
			ID:       uuid.New(),
			Key:      "list-flow-" + string(rune('0'+i)),
			Name:     "List Flow " + string(rune('0'+i)),
			Category: "test",
			Status:   1,
		}
		require.NoError(t, repo.Create(ctx, nil, def))
	}

	// List with keyword
	defs, total, err := repo.List(ctx, 1, 10, "List", "", nil)
	require.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, defs, 5)

	// List with category filter
	defs, total, err = repo.List(ctx, 1, 10, "", "test", nil)
	require.NoError(t, err)
	assert.Equal(t, int64(5), total)

	// List with status filter
	status := 1
	defs, total, err = repo.List(ctx, 1, 10, "", "test", &status)
	require.NoError(t, err)
	assert.Equal(t, int64(5), total)
}

func TestDefinitionRepo_CreateAndGetVersion(t *testing.T) {
	db := setupTestDB(t)
	repo := NewDefinitionRepo(db)
	ctx := context.Background()

	defID := uuid.New()
	def := &model.FlowDefinition{
		ID:   defID,
		Key:  "version-test",
		Name: "Version Test",
	}
	require.NoError(t, repo.Create(ctx, nil, def))

	// Create version 1
	ver1 := &model.FlowDefinitionVersion{
		ID:           uuid.New(),
		DefinitionID: defID,
		Version:      1,
		BpmnData:     []byte(`{"nodes":[],"edges":[]}`),
		Status:       1,
	}
	require.NoError(t, repo.CreateVersion(ctx, nil, ver1))

	// Create version 2
	ver2 := &model.FlowDefinitionVersion{
		ID:           uuid.New(),
		DefinitionID: defID,
		Version:      2,
		BpmnData:     []byte(`{"nodes":[],"edges":[]}`),
		Status:       0,
	}
	require.NoError(t, repo.CreateVersion(ctx, nil, ver2))

	// Get current version
	current, err := repo.GetCurrentVersion(ctx, defID)
	require.NoError(t, err)
	require.NotNil(t, current)
	assert.Equal(t, 1, current.Version)

	// Mark version 1 as historical, set version 2 as current
	require.NoError(t, repo.MarkVersionHistorical(ctx, nil, defID))

	// Verify version 1 is no longer current
	current2, err := repo.GetCurrentVersion(ctx, defID)
	require.NoError(t, err)
	assert.Nil(t, current2)
}

func TestDefinitionRepo_DuplicateVersion_UniqueConstraint(t *testing.T) {
	db := setupTestDB(t)
	repo := NewDefinitionRepo(db)
	ctx := context.Background()

	defID := uuid.New()
	def := &model.FlowDefinition{
		ID:   defID,
		Key:  "dup-version-test",
		Name: "Dup Version Test",
	}
	require.NoError(t, repo.Create(ctx, nil, def))

	ver1 := &model.FlowDefinitionVersion{
		ID:           uuid.New(),
		DefinitionID: defID,
		Version:      1,
		BpmnData:     []byte(`{}`),
	}
	require.NoError(t, repo.CreateVersion(ctx, nil, ver1))

	// Try to create another version with the same number
	ver1Dup := &model.FlowDefinitionVersion{
		ID:           uuid.New(),
		DefinitionID: defID,
		Version:      1,
		BpmnData:     []byte(`{}`),
	}
	err := repo.CreateVersion(ctx, nil, ver1Dup)
	assert.Error(t, err) // Should fail due to unique constraint
}

// ========== InstanceRepo Tests ==========

func TestInstanceRepo_CreateAndGet(t *testing.T) {
	db := setupTestDB(t)
	defRepo := NewDefinitionRepo(db)
	instRepo := NewInstanceRepo(db)
	ctx := context.Background()

	// Create a definition first (needed for FK)
	defID := uuid.New()
	require.NoError(t, defRepo.Create(ctx, nil, &model.FlowDefinition{
		ID:   defID,
		Key:  "inst-test",
		Name: "Instance Test",
	}))

	verID := uuid.New()
	require.NoError(t, defRepo.CreateVersion(ctx, nil, &model.FlowDefinitionVersion{
		ID:           verID,
		DefinitionID: defID,
		Version:      1,
		BpmnData:     []byte(`{}`),
		Status:       1,
	}))

	instID := uuid.New()
	inst := &model.FlowInstance{
		ID:                  instID,
		DefinitionID:        defID,
		DefinitionVersionID: verID,
		InitiatorID:         uuid.New(),
		Status:              0,
		CurrentNodeIDs:      []byte(`["node-1"]`),
	}
	require.NoError(t, instRepo.Create(ctx, nil, inst))

	got, err := instRepo.GetByID(ctx, instID)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, defID, got.DefinitionID)
}

func TestInstanceRepo_GetByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewInstanceRepo(db)
	ctx := context.Background()

	got, err := repo.GetByID(ctx, uuid.New())
	require.NoError(t, err)
	assert.Nil(t, got)
}

// ========== TaskRepo Tests ==========

func TestTaskRepo_CreateAndGet(t *testing.T) {
	db := setupTestDB(t)
	defRepo := NewDefinitionRepo(db)
	instRepo := NewInstanceRepo(db)
	taskRepo := NewTaskRepo(db)
	ctx := context.Background()

	// Setup prerequisite data
	defID := uuid.New()
	require.NoError(t, defRepo.Create(ctx, nil, &model.FlowDefinition{
		ID: defID, Key: "task-test", Name: "Task Test",
	}))
	verID := uuid.New()
	require.NoError(t, defRepo.CreateVersion(ctx, nil, &model.FlowDefinitionVersion{
		ID: verID, DefinitionID: defID, Version: 1, BpmnData: []byte(`{}`), Status: 1,
	}))
	instID := uuid.New()
	require.NoError(t, instRepo.Create(ctx, nil, &model.FlowInstance{
		ID: instID, DefinitionID: defID, DefinitionVersionID: verID, InitiatorID: uuid.New(),
	}))

	assignee := uuid.New()
	taskID := uuid.New()
	task := &model.FlowTask{
		ID:         taskID,
		InstanceID: instID,
		NodeID:     "approval-1",
		NodeName:   "Manager Approval",
		TaskType:   "approval",
		AssigneeID: &assignee,
		Status:     0,
	}
	require.NoError(t, taskRepo.CreateTask(ctx, nil, task))

	// Get by ID
	got, err := taskRepo.GetTaskByID(ctx, taskID)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "approval-1", got.NodeID)
	assert.Equal(t, "Manager Approval", got.NodeName)

	// List todo tasks
	todos, total, err := taskRepo.ListTodoTasks(ctx, assignee, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, todos, 1)
	assert.Equal(t, "approval-1", todos[0].NodeID)

	// Complete the task
	task.Status = 1
	require.NoError(t, taskRepo.UpdateTask(ctx, nil, task))

	// List done tasks
	dones, total, err := taskRepo.ListDoneTasks(ctx, assignee, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, dones, 1)
}

// ========== VariableRepo Tests ==========

func TestVariableRepo_SetAndGet(t *testing.T) {
	db := setupTestDB(t)
	defRepo := NewDefinitionRepo(db)
	instRepo := NewInstanceRepo(db)
	varRepo := NewVariableRepo(db)
	ctx := context.Background()

	// Setup prerequisite data
	defID := uuid.New()
	require.NoError(t, defRepo.Create(ctx, nil, &model.FlowDefinition{
		ID: defID, Key: "var-test", Name: "Var Test",
	}))
	verID := uuid.New()
	require.NoError(t, defRepo.CreateVersion(ctx, nil, &model.FlowDefinitionVersion{
		ID: verID, DefinitionID: defID, Version: 1, BpmnData: []byte(`{}`), Status: 1,
	}))
	instID := uuid.New()
	require.NoError(t, instRepo.Create(ctx, nil, &model.FlowInstance{
		ID: instID, DefinitionID: defID, DefinitionVersionID: verID, InitiatorID: uuid.New(),
	}))

	// Set a variable
	v := &model.FlowVariable{
		ID:         uuid.New(),
		InstanceID: instID,
		Key:        "applicant_name",
		Value:      []byte(`"Alice"`),
		Scope:      "global",
	}
	require.NoError(t, varRepo.Set(ctx, nil, v))

	// Get the variable
	got, err := varRepo.Get(ctx, instID, "applicant_name")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "applicant_name", got.Key)

	// Verify value
	var val string
	require.NoError(t, json.Unmarshal(got.Value, &val))
	assert.Equal(t, "Alice", val)

	// Update the variable (upsert)
	v2 := &model.FlowVariable{
		ID:         uuid.New(),
		InstanceID: instID,
		Key:        "applicant_name",
		Value:      []byte(`"Bob"`),
		Scope:      "global",
	}
	require.NoError(t, varRepo.Set(ctx, nil, v2))

	// Verify update
	got2, err := varRepo.Get(ctx, instID, "applicant_name")
	require.NoError(t, err)
	require.NotNil(t, got2)
	var val2 string
	require.NoError(t, json.Unmarshal(got2.Value, &val2))
	assert.Equal(t, "Bob", val2)
}

func TestVariableRepo_SetAndGetValueTypes(t *testing.T) {
	db := setupTestDB(t)
	defRepo := NewDefinitionRepo(db)
	instRepo := NewInstanceRepo(db)
	varRepo := NewVariableRepo(db)
	ctx := context.Background()

	defID := uuid.New()
	require.NoError(t, defRepo.Create(ctx, nil, &model.FlowDefinition{
		ID: defID, Key: "var-types-test", Name: "Var Types Test",
	}))
	verID := uuid.New()
	require.NoError(t, defRepo.CreateVersion(ctx, nil, &model.FlowDefinitionVersion{
		ID: verID, DefinitionID: defID, Version: 1, BpmnData: []byte(`{}`), Status: 1,
	}))
	instID := uuid.New()
	require.NoError(t, instRepo.Create(ctx, nil, &model.FlowInstance{
		ID: instID, DefinitionID: defID, DefinitionVersionID: verID, InitiatorID: uuid.New(),
	}))

	// Test various value types
	testVars := map[string]interface{}{
		"string_val": "hello",
		"int_val":    42,
		"bool_val":   true,
		"float_val":  3.14,
		"array_val":  []string{"a", "b", "c"},
	}

	require.NoError(t, varRepo.SetMap(ctx, nil, instID, testVars))

	got, err := varRepo.GetMap(ctx, instID)
	require.NoError(t, err)
	assert.Equal(t, "hello", got["string_val"])
	assert.Equal(t, float64(42), got["int_val"]) // JSON numbers become float64
	assert.Equal(t, true, got["bool_val"])
}

func TestVariableRepo_GetWithScope(t *testing.T) {
	db := setupTestDB(t)
	defRepo := NewDefinitionRepo(db)
	instRepo := NewInstanceRepo(db)
	varRepo := NewVariableRepo(db)
	ctx := context.Background()

	defID := uuid.New()
	require.NoError(t, defRepo.Create(ctx, nil, &model.FlowDefinition{
		ID: defID, Key: "scope-test", Name: "Scope Test",
	}))
	verID := uuid.New()
	require.NoError(t, defRepo.CreateVersion(ctx, nil, &model.FlowDefinitionVersion{
		ID: verID, DefinitionID: defID, Version: 1, BpmnData: []byte(`{}`), Status: 1,
	}))
	instID := uuid.New()
	require.NoError(t, instRepo.Create(ctx, nil, &model.FlowInstance{
		ID: instID, DefinitionID: defID, DefinitionVersionID: verID, InitiatorID: uuid.New(),
	}))

	// Create same key with different scopes
	require.NoError(t, varRepo.Set(ctx, nil, &model.FlowVariable{
		ID: uuid.New(), InstanceID: instID, Key: "counter", Value: []byte(`1`), Scope: "global",
	}))
	require.NoError(t, varRepo.Set(ctx, nil, &model.FlowVariable{
		ID: uuid.New(), InstanceID: instID, Key: "counter", Value: []byte(`2`), Scope: "local",
	}))

	// Get global scope (default)
	gotGlobal, err := varRepo.Get(ctx, instID, "counter")
	require.NoError(t, err)
	require.NotNil(t, gotGlobal)
	assert.Equal(t, "global", gotGlobal.Scope)

	// Get local scope
	gotLocal, err := varRepo.GetWithScope(ctx, instID, "counter", "local")
	require.NoError(t, err)
	require.NotNil(t, gotLocal)
	assert.Equal(t, "local", gotLocal.Scope)
}
