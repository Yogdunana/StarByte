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
) {
	wf := r.Group("/workflow")
	{
		// ========== 流程定义 ==========
		defs := wf.Group("/definitions")
		{
			defs.GET("", defHandler.List)
			defs.POST("", defHandler.Create)
			defs.GET("/:id", defHandler.GetByID)
			defs.PUT("/:id", defHandler.Update)
			defs.DELETE("/:id", defHandler.Delete)
			defs.POST("/:id/publish", defHandler.Publish)
			defs.PUT("/:id/draft", defHandler.SaveDraft)
			defs.GET("/:id/versions", defHandler.ListVersions)
			defs.GET("/:id/versions/:versionId", defHandler.GetVersionByID)
		}

		// ========== 流程实例 ==========
		instances := wf.Group("/instances")
		{
			instances.POST("", instHandler.Start)
			instances.GET("", instHandler.List)
			instances.GET("/:id", instHandler.GetByID)
			instances.POST("/:id/terminate", instHandler.Terminate)
			instances.POST("/:id/suspend", instHandler.Suspend)
			instances.POST("/:id/resume", instHandler.Resume)
			instances.GET("/:id/history", instHandler.ListHistory)
		}

		// ========== 流程任务 ==========
		tasks := wf.Group("/tasks")
		{
			tasks.GET("/todo", taskHandler.ListTodoTasks)
			tasks.GET("/done", taskHandler.ListDoneTasks)
			tasks.GET("/:id", taskHandler.GetByID)
			tasks.POST("/:id/approve", taskHandler.Approve)
			tasks.POST("/:id/reject", taskHandler.Reject)
			tasks.POST("/:id/transfer", taskHandler.Transfer)
			tasks.POST("/:id/rollback", taskHandler.Rollback)
		}
	}
}
