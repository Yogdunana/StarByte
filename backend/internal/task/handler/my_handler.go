package handler

import (
	"github.com/Yogdunana/StarByte/backend/internal/task/dto"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

func (h *TaskHandler) myKind(c *gin.Context, kind string) {
	userID, err := getUserID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.MyTaskRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	list, total, err := h.svc.ListMy(c.Request.Context(), userID, kind, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	page, size := defaultPage(req.Page, req.PageSize)
	response.Page(c, list, total, page, size)
}

// MyTodo 我的待办
// @Summary 我的待办
// @Description 当前用户待办任务列表
// @Tags 任务
// @Produce json
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /tasks/my/todo [get]
// @Security BearerAuth
func (h *TaskHandler) MyTodo(c *gin.Context) { h.myKind(c, "todo") }

// MyDone 我的已办
// @Summary 我的已办
// @Description 当前用户已办任务列表
// @Tags 任务
// @Produce json
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /tasks/my/done [get]
// @Security BearerAuth
func (h *TaskHandler) MyDone(c *gin.Context) { h.myKind(c, "done") }

// MyCreated 我创建的
// @Summary 我创建的任务
// @Description 当前用户创建的任务列表
// @Tags 任务
// @Produce json
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /tasks/my/created [get]
// @Security BearerAuth
func (h *TaskHandler) MyCreated(c *gin.Context) { h.myKind(c, "created") }

// MyOverdue 我的超期
// @Summary 我的超期任务
// @Description 当前用户超期任务列表
// @Tags 任务
// @Produce json
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /tasks/my/overdue [get]
// @Security BearerAuth
func (h *TaskHandler) MyOverdue(c *gin.Context) { h.myKind(c, "overdue") }
