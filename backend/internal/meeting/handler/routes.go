package handler

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/Yogdunana/StarByte/backend/internal/meeting/service"
	rbacRepo "github.com/Yogdunana/StarByte/backend/internal/rbac/repo"
	rbacService "github.com/Yogdunana/StarByte/backend/internal/rbac/service"
	"github.com/Yogdunana/StarByte/backend/pkg/middleware"
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
	g.Use(middleware.RequireDataScope("meeting:read"))
	g.Use(middleware.PermissionRequired(cacheService))
	g.Use(middleware.DataScopeMiddleware(db, deptRepo, cacheService))
	g.Use(meetingViewer(false))
	g.Use(middleware.RequireDataScope("meeting:manage"), middleware.DataScopeMiddleware(db, deptRepo, cacheService), meetingCapabilities("meeting:manage"))
	g.Use(middleware.RequireDataScope("meeting:update"), middleware.DataScopeMiddleware(db, deptRepo, cacheService), meetingCapabilities("meeting:update"))
	g.Use(middleware.RequireDataScope("meeting:delete"), middleware.DataScopeMiddleware(db, deptRepo, cacheService), meetingCapabilities("meeting:delete"))
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
	g.Use(meetingViewer(false))
	g.POST("/:id/checkin", h.Checkin)

	read := withReadScope(g, cacheService, db, deptRepo)
	read.GET("", h.ListMeetings)
	read.GET("/:id", h.GetMeeting)
	read.GET("/:id/agendas", h.ListAgendas)
	read.GET("/:id/attendees", h.ListAttendees)
	read.GET("/:id/votes", h.ListVotes)

	create := withPermission(g, "meeting:create", cacheService)
	create.POST("", h.CreateMeeting)

	update := withMeetingScope(withPermission(g, "meeting:update", cacheService), "meeting:update", cacheService, db, deptRepo)
	update.PUT("/:id", h.UpdateMeeting)

	del := withMeetingScope(withPermission(g, "meeting:delete", cacheService), "meeting:delete", cacheService, db, deptRepo)
	del.DELETE("/:id", h.DeleteMeeting)

	manage := withMeetingScope(withPermission(g, "meeting:manage", cacheService), "meeting:manage", cacheService, db, deptRepo)
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
	votes.Use(meetingViewer(false))
	votes.POST("/:id/cast", h.CastVote)
	votes.GET("/:id/my", h.MyVote)
	voteRead := withReadScope(votes, cacheService, db, deptRepo)
	voteRead.GET("/:id", h.GetVote)
	voteRead.GET("/:id/result", h.VoteResult)
	voteManage := withMeetingScope(withPermission(votes, "meeting:manage", cacheService), "meeting:manage", cacheService, db, deptRepo)
	voteManage.POST("/:id/close", h.CloseVote)

	sysRead := withPermission(r, "meeting:read", cacheService)
	sysRead.GET("/system/vote-weight-config", h.GetWeightConfig)
	sysWrite := withPermission(r, "system:config", cacheService)
	sysWrite.PUT("/system/vote-weight-config", h.UpdateWeightConfig)
}
