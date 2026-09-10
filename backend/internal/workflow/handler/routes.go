package handler

import (
	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册工作流引擎路由。
//
// 路由前缀: /workflow
//
// 流程定义管理:
//
//	GET    /definitions              流程定义列表
//	POST   /definitions              创建流程定义
//	GET    /definitions/:id          流程定义详情
//	PUT    /definitions/:id          更新流程定义
//	DELETE /definitions/:id          删除流程定义
//	POST   /definitions/:id/publish  发布流程定义
//	PUT    /definitions/:id/draft    保存流程草稿图
//	GET    /definitions/:id/versions 版本列表
//	GET    /definitions/:id/versions/:versionId 版本详情
//
// 流程实例管理:
//
//	POST   /instances                启动流程实例
//	GET    /instances                流程实例列表
//	GET    /instances/:id            流程实例详情
//	POST   /instances/:id/terminate  终止流程
//	POST   /instances/:id/suspend    挂起流程
//	POST   /instances/:id/resume     恢复流程
//	GET    /instances/:id/history    流程历史记录
//
// 流程任务管理:
//
//	GET    /tasks/todo               我的待办任务
//	GET    /tasks/done               我的已办任务
//	GET    /tasks/:id                任务详情
//	POST   /tasks/:id/approve        审批通过
//	POST   /tasks/:id/reject         审批驳回
//	POST   /tasks/:id/transfer       转办任务
//	POST   /tasks/:id/rollback       退回任务
func RegisterRoutes(
	r *gin.RouterGroup,
	defHandler *DefinitionHandler,
	instHandler *InstanceHandler,
	taskHandler *TaskHandler,
	security ...RouteSecurity,
) {
	var config RouteSecurity
	if len(security) > 0 {
		config = security[0]
	}
	wf := r.Group("/workflow")
	{
		// ========== 流程定义 ==========
		defs := wf.Group("/definitions")
		{
			workflowPermission(defs, "workflow:read", config).GET("", defHandler.List)
			workflowPermission(defs, "workflow:create", config).POST("", defHandler.Create)
			workflowPermission(defs, "workflow:read", config).GET("/:id", defHandler.GetByID)
			workflowPermission(workflowViewer(defs, "workflow:update", config), "workflow:update", config).PUT("/:id", defHandler.Update)
			workflowPermission(workflowViewer(defs, "workflow:delete", config), "workflow:delete", config).DELETE("/:id", defHandler.Delete)
			workflowPermission(workflowViewer(defs, "workflow:update", config), "workflow:update", config).POST("/:id/publish", defHandler.Publish)
			workflowPermission(workflowViewer(defs, "workflow:update", config), "workflow:update", config).PUT("/:id/draft", defHandler.SaveDraft)
			workflowPermission(defs, "workflow:read", config).GET("/:id/versions", defHandler.ListVersions)
			workflowPermission(defs, "workflow:read", config).GET("/:id/versions/:versionId", defHandler.GetVersionByID)
		}

		// ========== 流程实例 ==========
		instances := wf.Group("/instances")
		{
			workflowPermission(instances, "workflow:create", config).POST("", instHandler.Start)
			workflowViewer(instances, "workflow:read", config).GET("", instHandler.List)
			workflowViewer(instances, "workflow:read", config).GET("/:id", instHandler.GetByID)
			workflowViewer(instances, "workflow:update", config).POST("/:id/terminate", instHandler.Terminate)
			workflowViewer(instances, "workflow:update", config).POST("/:id/suspend", instHandler.Suspend)
			workflowViewer(instances, "workflow:update", config).POST("/:id/resume", instHandler.Resume)
			workflowViewer(instances, "workflow:read", config).GET("/:id/history", instHandler.ListHistory)
		}

		// ========== 流程任务 ==========
		tasks := wf.Group("/tasks")
		{
			tasks.GET("/todo", taskHandler.ListTodoTasks)
			tasks.GET("/done", taskHandler.ListDoneTasks)
			workflowViewer(tasks, "workflow:read", config).GET("/:id", taskHandler.GetByID)
			workflowViewer(tasks, "workflow:read", config).GET("/:id/assignees", taskHandler.TransferCandidates)
			tasks.POST("/:id/approve", taskHandler.Approve)
			tasks.POST("/:id/reject", taskHandler.Reject)
			tasks.POST("/:id/transfer", taskHandler.Transfer)
			tasks.POST("/:id/rollback", taskHandler.Rollback)
		}
	}
}
