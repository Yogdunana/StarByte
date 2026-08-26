package workflow

import (
	"context"

	"github.com/Yogdunana/StarByte/backend/internal/workflow/engine"
	"github.com/Yogdunana/StarByte/backend/internal/workflow/engine/nodes"
	"github.com/Yogdunana/StarByte/backend/internal/workflow/handler"
	"github.com/Yogdunana/StarByte/backend/internal/workflow/model"
	"github.com/Yogdunana/StarByte/backend/internal/workflow/repo"
	"github.com/Yogdunana/StarByte/backend/internal/workflow/service"
	"github.com/Yogdunana/StarByte/backend/pkg/events"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Handlers 封装所有 workflow handler 和扩展能力，便于在 main.go 中使用。
type Handlers struct {
	Definition      *handler.DefinitionHandler
	Instance        *handler.InstanceHandler
	Task            *handler.TaskHandler
	serviceTaskNode *nodes.ServiceTaskNode
}

// RegisterServiceCallback 注册一个服务任务回调函数，供 service_task 节点调用。
// 业务模块可以通过此方法将自定义业务逻辑接入工作流引擎。
// 线程安全。
func (h *Handlers) RegisterServiceCallback(name string, cb func(ctx context.Context, inst *model.FlowInstance, node *engine.FlowNode, vars map[string]interface{}) (map[string]interface{}, error)) {
	h.serviceTaskNode.RegisterService(name, nodes.ServiceCallback(cb))
}

// Init 初始化工作流引擎的所有依赖（repo → engine → service → handler）。
//
// 依赖关系:
//   Repo 层 ← 数据库
//   Engine 层 ← Repo + EventBus + NodeRegistry + ExpressionEngine
//   Service 层 ← Repo + Engine
//   Handler 层 ← Service
//
// 返回所有 handler 的集合，调用方通过 RegisterRoutes 注册路由。
func Init(db *gorm.DB, eventBus *events.EventBus, logger *zap.Logger) *Handlers {
	// 1. Repo 层
	defRepo := repo.NewDefinitionRepo(db)
	instRepo := repo.NewInstanceRepo(db)
	taskRepo := repo.NewTaskRepo(db)
	varRepo := repo.NewVariableRepo(db)

	// 2. 表达式引擎
	exprEngine := engine.NewExpressionEngine()

	// 3. 节点注册表 + 内置节点处理器
	registry := nodes.NewNodeRegistry()
	registry.Register(&nodes.StartNode{})
	registry.Register(&nodes.EndNode{})
	registry.Register(&nodes.ApprovalNode{TaskRepo: taskRepo, EventBus: eventBus})
	registry.Register(&nodes.ExclusiveGatewayNode{ExprEngine: exprEngine})
	registry.Register(&nodes.ParallelGatewayNode{})

	serviceTaskNode := &nodes.ServiceTaskNode{Callbacks: make(map[string]nodes.ServiceCallback)}
	registry.Register(serviceTaskNode)

	registry.Register(&nodes.NotificationTaskNode{EventBus: eventBus})

	// 4. 流程引擎核心
	flowEngine := engine.NewFlowEngine(
		defRepo,
		instRepo,
		taskRepo,
		varRepo,
		db,
		registry,
		exprEngine,
		eventBus,
		logger,
	)

	// 5. Service 层
	defService := service.NewDefinitionService(defRepo, db)
	instService := service.NewInstanceService(instRepo, defRepo, taskRepo, flowEngine, db)
	taskService := service.NewTaskService(taskRepo, instRepo, flowEngine, db)

	// 6. Handler 层
	defHandler := handler.NewDefinitionHandler(defService)
	instHandler := handler.NewInstanceHandler(instService)
	taskHandler := handler.NewTaskHandler(taskService)

	return &Handlers{
		Definition:      defHandler,
		Instance:        instHandler,
		Task:            taskHandler,
		serviceTaskNode: serviceTaskNode,
	}
}
