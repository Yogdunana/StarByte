package main

import (
	contractHandler "github.com/Yogdunana/StarByte/backend/internal/contract/handler"
	contractRepo "github.com/Yogdunana/StarByte/backend/internal/contract/repo"
	contractService "github.com/Yogdunana/StarByte/backend/internal/contract/service"
	disciplineHandler "github.com/Yogdunana/StarByte/backend/internal/discipline/handler"
	disciplineRepo "github.com/Yogdunana/StarByte/backend/internal/discipline/repo"
	disciplineService "github.com/Yogdunana/StarByte/backend/internal/discipline/service"
	financeHandler "github.com/Yogdunana/StarByte/backend/internal/finance/handler"
	financeRepo "github.com/Yogdunana/StarByte/backend/internal/finance/repo"
	financeService "github.com/Yogdunana/StarByte/backend/internal/finance/service"
	notifService "github.com/Yogdunana/StarByte/backend/internal/notification/service"
	rbacRepo "github.com/Yogdunana/StarByte/backend/internal/rbac/repo"
	rbacService "github.com/Yogdunana/StarByte/backend/internal/rbac/service"
	schedService "github.com/Yogdunana/StarByte/backend/internal/scheduler/service"
	wfRepo "github.com/Yogdunana/StarByte/backend/internal/workflow/repo"
	wfService "github.com/Yogdunana/StarByte/backend/internal/workflow/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func registerPhase1Ops(
	r *gin.RouterGroup,
	db *gorm.DB,
	cache rbacService.PermissionCacheService,
	deptRepo rbacRepo.DepartmentRepo,
	notif notifService.NotificationService,
	defs wfRepo.DefinitionRepo,
	inst wfService.InstanceService,
) {
	finH := financeHandler.New(financeService.New(financeRepo.New(db)))
	financeHandler.RegisterRoutes(r, finH, cache)

	discSvc := disciplineService.New(
		disciplineRepo.New(db),
		disciplineService.NewNotifier(notif),
		disciplineService.NewFlowStarter(defs, inst),
	)
	disciplineHandler.RegisterRoutes(r, disciplineHandler.New(discSvc), cache, db, deptRepo)

	conSvc := contractService.New(contractRepo.New(db), contractService.NewNotifier(notif))
	contractHandler.RegisterRoutes(r, contractHandler.New(conSvc), cache)
	schedService.RegisterHandler("contract_expiry", "扫描并标记到期合同、提醒临期合同", conSvc.ExpiryJob)
}
