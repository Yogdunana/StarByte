package handler

import (
	"github.com/Yogdunana/StarByte/backend/internal/meeting/service"
	rbacRepo "github.com/Yogdunana/StarByte/backend/internal/rbac/repo"
	rbacService "github.com/Yogdunana/StarByte/backend/internal/rbac/service"
	"github.com/Yogdunana/StarByte/backend/pkg/middleware"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type MeetingHandler struct {
	svc service.MeetingService
}

func NewMeetingHandler(svc service.MeetingService) *MeetingHandler {
	return &MeetingHandler{svc: svc}
}

func withPermission(group *gin.RouterGroup, permCode string, cacheService rbacService.PermissionCacheService) *gin.RouterGroup {
	g := group.Group("")
	g.Use(middleware.RequirePermission(permCode))
	g.Use(middleware.PermissionRequired(cacheService))
	return g
}

func withReadScope(
	group *gin.RouterGroup,
	cacheService rbacService.PermissionCacheService,
	db *gorm.DB,
	deptRepo rbacRepo.DepartmentRepo,
) *gin.RouterGroup {
	g := group.Group("")
	g.Use(middleware.RequirePermission("meeting:read"))
	g.Use(middleware.RequireDataScope("meeting"))
	g.Use(middleware.PermissionRequired(cacheService))
	g.Use(middleware.DataScopeMiddleware(db, deptRepo, cacheService))
	return g
}

// RegisterRoutes 注册 /api/v1/meetings、/votes、/system/vote-weight-config。
func RegisterRoutes(
	r *gin.RouterGroup,
	h *MeetingHandler,
	cacheService rbacService.PermissionCacheService,
	db *gorm.DB,
	deptRepo rbacRepo.DepartmentRepo,
) {
	g := r.Group("/meetings")
	g.POST("/:id/checkin", h.Checkin)

	read := withReadScope(g, cacheService, db, deptRepo)
	read.GET("", h.ListMeetings)
	read.GET("/:id", h.GetMeeting)
	read.GET("/:id/agendas", h.ListAgendas)
	read.GET("/:id/attendees", h.ListAttendees)
	read.GET("/:id/votes", h.ListVotes)

	create := withPermission(g, "meeting:create", cacheService)
	create.POST("", h.CreateMeeting)

	update := withPermission(g, "meeting:update", cacheService)
	update.PUT("/:id", h.UpdateMeeting)

	del := withPermission(g, "meeting:delete", cacheService)
	del.DELETE("/:id", h.DeleteMeeting)

	manage := withPermission(g, "meeting:manage", cacheService)
	manage.POST("/:id/start", h.StartMeeting)
	manage.POST("/:id/end", h.EndMeeting)
	manage.POST("/:id/cancel", h.CancelMeeting)
	manage.PUT("/:id/minutes", h.UpdateMinutes)
	manage.GET("/:id/qrcode", h.MeetingQRCode)
	manage.POST("/:id/agendas", h.AddAgenda)
	manage.PUT("/:id/agendas/sort", h.SortAgendas)
	manage.PUT("/:id/agendas/:aid", h.UpdateAgenda)
	manage.DELETE("/:id/agendas/:aid", h.DeleteAgenda)
	manage.POST("/:id/attendees", h.AddAttendees)
	manage.DELETE("/:id/attendees/:uid", h.RemoveAttendee)
	manage.POST("/:id/votes", h.CreateVote)

	votes := r.Group("/votes")
	votes.POST("/:id/cast", h.CastVote)
	votes.GET("/:id/my", h.MyVote)
	voteRead := withReadScope(votes, cacheService, db, deptRepo)
	voteRead.GET("/:id", h.GetVote)
	voteRead.GET("/:id/result", h.VoteResult)
	voteManage := withPermission(votes, "meeting:manage", cacheService)
	voteManage.POST("/:id/close", h.CloseVote)

	sysRead := withPermission(r, "meeting:read", cacheService)
	sysRead.GET("/system/vote-weight-config", h.GetWeightConfig)
	sysWrite := withPermission(r, "system:config", cacheService)
	sysWrite.PUT("/system/vote-weight-config", h.UpdateWeightConfig)
}
