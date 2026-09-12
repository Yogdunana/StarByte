package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

	activityHandler "github.com/Yogdunana/StarByte/backend/internal/activity/handler"
	activityRepo "github.com/Yogdunana/StarByte/backend/internal/activity/repo"
	activityService "github.com/Yogdunana/StarByte/backend/internal/activity/service"
	announcementHandler "github.com/Yogdunana/StarByte/backend/internal/announcement/handler"
	announcementRepo "github.com/Yogdunana/StarByte/backend/internal/announcement/repo"
	announcementService "github.com/Yogdunana/StarByte/backend/internal/announcement/service"
	auditHandler "github.com/Yogdunana/StarByte/backend/internal/audit/handler"
	auditRepo "github.com/Yogdunana/StarByte/backend/internal/audit/repo"
	auditService "github.com/Yogdunana/StarByte/backend/internal/audit/service"
	authHandler "github.com/Yogdunana/StarByte/backend/internal/auth/handler"
	authRepo "github.com/Yogdunana/StarByte/backend/internal/auth/repo"
	authService "github.com/Yogdunana/StarByte/backend/internal/auth/service"
	backupHandler "github.com/Yogdunana/StarByte/backend/internal/backup/handler"
	backupRepo "github.com/Yogdunana/StarByte/backend/internal/backup/repo"
	backupService "github.com/Yogdunana/StarByte/backend/internal/backup/service"
	cacheadminHandler "github.com/Yogdunana/StarByte/backend/internal/cache/handler"
	cacheadminService "github.com/Yogdunana/StarByte/backend/internal/cache/service"
	cfgstoreHandler "github.com/Yogdunana/StarByte/backend/internal/configstore/handler"
	cfgstoreRepo "github.com/Yogdunana/StarByte/backend/internal/configstore/repo"
	cfgstoreService "github.com/Yogdunana/StarByte/backend/internal/configstore/service"
	dictHandler "github.com/Yogdunana/StarByte/backend/internal/dict/handler"
	dictRepo "github.com/Yogdunana/StarByte/backend/internal/dict/repo"
	dictService "github.com/Yogdunana/StarByte/backend/internal/dict/service"
	exportHandler "github.com/Yogdunana/StarByte/backend/internal/export/handler"
	exportRepo "github.com/Yogdunana/StarByte/backend/internal/export/repo"
	exportService "github.com/Yogdunana/StarByte/backend/internal/export/service"
	featureHandler "github.com/Yogdunana/StarByte/backend/internal/feature/handler"
	featureModel "github.com/Yogdunana/StarByte/backend/internal/feature/model"
	featureRepo "github.com/Yogdunana/StarByte/backend/internal/feature/repo"
	featureService "github.com/Yogdunana/StarByte/backend/internal/feature/service"
	fileHandler "github.com/Yogdunana/StarByte/backend/internal/file/handler"
	fileRepo "github.com/Yogdunana/StarByte/backend/internal/file/repo"
	fileService "github.com/Yogdunana/StarByte/backend/internal/file/service"
	formHandler "github.com/Yogdunana/StarByte/backend/internal/form/handler"
	formRepo "github.com/Yogdunana/StarByte/backend/internal/form/repo"
	formService "github.com/Yogdunana/StarByte/backend/internal/form/service"
	internshipHandler "github.com/Yogdunana/StarByte/backend/internal/internship/handler"
	internshipRepo "github.com/Yogdunana/StarByte/backend/internal/internship/repo"
	internshipService "github.com/Yogdunana/StarByte/backend/internal/internship/service"
	interviewHandler "github.com/Yogdunana/StarByte/backend/internal/interview/handler"
	interviewRepo "github.com/Yogdunana/StarByte/backend/internal/interview/repo"
	interviewService "github.com/Yogdunana/StarByte/backend/internal/interview/service"
	knowledgeHandler "github.com/Yogdunana/StarByte/backend/internal/knowledge/handler"
	knowledgeRepo "github.com/Yogdunana/StarByte/backend/internal/knowledge/repo"
	knowledgeService "github.com/Yogdunana/StarByte/backend/internal/knowledge/service"
	leaveHandler "github.com/Yogdunana/StarByte/backend/internal/leave/handler"
	leaveRepo "github.com/Yogdunana/StarByte/backend/internal/leave/repo"
	leaveService "github.com/Yogdunana/StarByte/backend/internal/leave/service"
	meetingHandler "github.com/Yogdunana/StarByte/backend/internal/meeting/handler"
	meetingRepo "github.com/Yogdunana/StarByte/backend/internal/meeting/repo"
	meetingService "github.com/Yogdunana/StarByte/backend/internal/meeting/service"
	memberHandler "github.com/Yogdunana/StarByte/backend/internal/member/handler"
	memberidentity "github.com/Yogdunana/StarByte/backend/internal/member/identity"
	memberRepo "github.com/Yogdunana/StarByte/backend/internal/member/repo"
	memberService "github.com/Yogdunana/StarByte/backend/internal/member/service"
	monitorHandler "github.com/Yogdunana/StarByte/backend/internal/monitor/handler"
	monitorService "github.com/Yogdunana/StarByte/backend/internal/monitor/service"
	notifHandler "github.com/Yogdunana/StarByte/backend/internal/notification/handler"
	notifModel "github.com/Yogdunana/StarByte/backend/internal/notification/model"
	notifRepo "github.com/Yogdunana/StarByte/backend/internal/notification/repo"
	notifService "github.com/Yogdunana/StarByte/backend/internal/notification/service"
	rbacHandler "github.com/Yogdunana/StarByte/backend/internal/rbac/handler"
	rbacRepo "github.com/Yogdunana/StarByte/backend/internal/rbac/repo"
	rbacService "github.com/Yogdunana/StarByte/backend/internal/rbac/service"
	reportHandler "github.com/Yogdunana/StarByte/backend/internal/report/handler"
	reportRepo "github.com/Yogdunana/StarByte/backend/internal/report/repo"
	reportService "github.com/Yogdunana/StarByte/backend/internal/report/service"
	scheduleHandler "github.com/Yogdunana/StarByte/backend/internal/schedule/handler"
	scheduleRepo "github.com/Yogdunana/StarByte/backend/internal/schedule/repo"
	scheduleService "github.com/Yogdunana/StarByte/backend/internal/schedule/service"
	schedHandler "github.com/Yogdunana/StarByte/backend/internal/scheduler/handler"
	schedRepo "github.com/Yogdunana/StarByte/backend/internal/scheduler/repo"
	schedService "github.com/Yogdunana/StarByte/backend/internal/scheduler/service"
	searchHandler "github.com/Yogdunana/StarByte/backend/internal/search/handler"
	searchService "github.com/Yogdunana/StarByte/backend/internal/search/service"
	statsHandler "github.com/Yogdunana/StarByte/backend/internal/stats/handler"
	statsRepo "github.com/Yogdunana/StarByte/backend/internal/stats/repo"
	statsService "github.com/Yogdunana/StarByte/backend/internal/stats/service"
	taskHandler "github.com/Yogdunana/StarByte/backend/internal/task/handler"
	taskRepo "github.com/Yogdunana/StarByte/backend/internal/task/repo"
	taskService "github.com/Yogdunana/StarByte/backend/internal/task/service"
	"github.com/Yogdunana/StarByte/backend/internal/user/handler"
	"github.com/Yogdunana/StarByte/backend/internal/user/repo"
	"github.com/Yogdunana/StarByte/backend/internal/user/service"
	"github.com/Yogdunana/StarByte/backend/internal/workflow"
	wfHandler "github.com/Yogdunana/StarByte/backend/internal/workflow/handler"
	"github.com/Yogdunana/StarByte/backend/pkg/config"
	"github.com/Yogdunana/StarByte/backend/pkg/configstore"
	"github.com/Yogdunana/StarByte/backend/pkg/database"
	"github.com/Yogdunana/StarByte/backend/pkg/events"
	"github.com/Yogdunana/StarByte/backend/pkg/logger"
	"github.com/Yogdunana/StarByte/backend/pkg/metrics"
	"github.com/Yogdunana/StarByte/backend/pkg/middleware"
	authmiddleware "github.com/Yogdunana/StarByte/backend/pkg/middleware/auth"
	"github.com/Yogdunana/StarByte/backend/pkg/middleware/circuitbreaker"
	"github.com/Yogdunana/StarByte/backend/pkg/middleware/ratelimit"
	"github.com/Yogdunana/StarByte/backend/pkg/redis"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/Yogdunana/StarByte/backend/pkg/storage"
)

// @title StarByte API
// @version 1.0
// @description 高校计算机协会一体化管理平台
// @host localhost:8080
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description 输入 Bearer {token}
func main() {
	// 1. 加载配置
	configPath := "configs/config.yaml"
	if envPath := os.Getenv("CONFIG_PATH"); envPath != "" {
		configPath = envPath
	}
	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Printf("load config failed: %v\n", err)
		os.Exit(1)
	}

	// 2. 初始化日志
	if err := logger.Init(&cfg.Logger); err != nil {
		fmt.Printf("init logger failed: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	if len(os.Args) > 1 && os.Args[1] == "backup" {
		if err := runBackupCLI(cfg, os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}

	logger.Info("server starting...")

	// 3. 初始化数据库
	if err := database.Init(&cfg.Database); err != nil {
		logger.Fatal("init database failed", zap.Error(err))
	}
	defer func() {
		if err := database.Close(); err != nil {
			logger.Error("close database failed", zap.Error(err))
		}
	}()

	// 3b. 自动迁移工作流引擎表
	if err := workflow.AutoMigrate(database.DB()); err != nil {
		logger.Fatal("auto migrate workflow tables failed", zap.Error(err))
	}

	// 3c. notifications 在 000001 之后模型字段有扩展，保留 AutoMigrate 做列对齐。
	// notification_templates 已有 000008 正式迁移，不再依赖 AutoMigrate 建表。
	if err := database.DB().AutoMigrate(&notifModel.Notification{}); err != nil {
		logger.Fatal("auto migrate notification tables failed", zap.Error(err))
	}

	// 4. 初始化 Redis
	if err := redis.Init(&cfg.Redis); err != nil {
		logger.Fatal("init redis failed", zap.Error(err))
	}
	defer func() {
		if err := redis.Close(); err != nil {
			logger.Error("close redis failed", zap.Error(err))
		}
	}()

	// 5. 设置 Gin 模式
	gin.SetMode(cfg.Server.Mode)

	// 6. 创建 Gin 引擎
	r := gin.New()
	// Default Gin trusts 0.0.0.0/0, so X-Forwarded-For is spoofable. Trust none
	// unless TRUSTED_PROXIES lists the load-balancer CIDRs (#75 security review).
	if err := r.SetTrustedProxies(ratelimit.TrustedProxiesFromEnv()); err != nil {
		logger.Fatal("invalid TRUSTED_PROXIES", zap.Error(err))
	}

	// 7. 注册全局中间件
	// 顺序: RequestID → Logger → Metrics → ErrorHandler → CORS
	// Metrics 必须在 ErrorHandler 之前，panic 被 recover 成 500 后仍能记账。
	// 注意: 全局限流不放在全局中间件中，以避免影响健康检查端点
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger())
	r.Use(middleware.Metrics())
	r.Use(middleware.ErrorHandler())
	r.Use(middleware.CORSWithConfig(cfg.CORS))

	// 8. 健康检查与 metrics（不受 API 限流影响）
	var pingMinio middleware.MinioPinger
	r.GET("/health", middleware.HealthCheck())
	r.GET("/health/ready", middleware.ReadinessCheck(database.DB(), redis.Client(), func(ctx context.Context) error {
		if pingMinio == nil {
			return fmt.Errorf("minio not configured")
		}
		return pingMinio(ctx)
	}))
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))
	registerSwagger(r)

	// 9. 初始化业务模块
	// 用户模块（共享 repo）
	userRepo := repo.NewUserRepo(database.DB())

	// RBAC 权限模块（repo 层先初始化，cacheService 供 auth 模块使用）
	roleRepo := rbacRepo.NewRoleRepo(database.DB())
	permRepo := rbacRepo.NewPermissionRepo(database.DB())
	deptRepo := rbacRepo.NewDepartmentRepo(database.DB())
	posRepo := rbacRepo.NewPositionRepo(database.DB())

	cacheService := rbacService.NewPermissionCacheService(database.DB(), redis.Client(), permRepo, roleRepo)

	// 特性开关 / 灰度一期（#98）：Redis 快照 + pub/sub 热更新，不拦截 CAS/登录
	featSvc := featureService.New(
		featureRepo.New(database.DB()),
		featureService.NewRedisSnapshot(redis.Client()),
		featureService.NewRedisBus(redis.Client()),
		cacheService,
		featureService.NewUserDepartments(userRepo),
		cacheService,
	)
	featureCtx, featureCancel := context.WithCancel(context.Background())
	defer featureCancel()
	if err := featSvc.StartHotReload(featureCtx); err != nil {
		logger.Error("start feature hot reload failed", zap.Error(err))
	}

	// 事件总线（登录/登出审计、工作流、通知共用）
	eventBus := events.NewEventBus()

	// 认证模块（依赖 cacheService 获取角色和权限）
	authR := authRepo.NewAuthRepo(redis.Client())
	memberProfRepo := memberRepo.NewProfileRepo(database.DB())
	authSvc := authService.NewAuthService(
		authR, userRepo, &cfg.JWT, cacheService, eventBus,
		memberidentity.NewLookup(memberProfRepo),
		&authService.CASDeps{
			Config:     &cfg.CAS,
			Store:      authRepo.NewCASStore(redis.Client()),
			Validator:  authService.NewHTTPTicketValidator(cfg.CAS.ServerURL, nil),
			AssignRole: authService.NewRoleAssigner(database.DB(), roleRepo, cfg.CAS.DefaultRole),
		},
	)
	authH := authHandler.NewAuthHandler(authSvc)

	// 用户管理模块
	userService := service.NewUserService(database.DB(), userRepo, &cfg.JWT)
	userHandler := handler.NewUserHandler(userService)

	// RBAC 权限模块（service 层）
	roleService := rbacService.NewRoleService(database.DB(), roleRepo, permRepo, cacheService)
	permService := rbacService.NewPermissionService(database.DB(), permRepo, cacheService)
	deptService := rbacService.NewDepartmentService(database.DB(), deptRepo)
	posService := rbacService.NewPositionService(database.DB(), posRepo)

	roleHandler := rbacHandler.NewRoleHandler(roleService)
	permHandler := rbacHandler.NewPermissionHandler(permService)
	deptHandler := rbacHandler.NewDepartmentHandler(deptService)
	posHandler := rbacHandler.NewPositionHandler(posService)

	// 工作流引擎模块
	wfHandlers := workflow.Init(database.DB(), eventBus, logger.GetLogger())

	// 对象存储（MinIO）：bucket 不存在则创建；失败只告警，避免拖垮其他模块
	var objectStore storage.ObjectStorage
	minioStore, err := storage.NewMinIO(cfg.MinIO)
	if err != nil {
		logger.Error("init MinIO client failed", zap.Error(err))
	} else {
		objectStore = minioStore
		pingMinio = minioStore.Ping
		if err := objectStore.EnsureBucket(context.Background()); err != nil {
			logger.Error("ensure MinIO bucket failed", zap.Error(err))
		}
	}

	// 通知模块
	notifR := notifRepo.NewNotificationRepo(database.DB())
	tplRepo := notifRepo.NewTemplateRepo(database.DB())

	hub := notifService.NewHub()
	emailLogs := notifRepo.NewEmailLogRepo(database.DB())
	cfgRows := cfgstoreRepo.NewConfigRepo(database.DB())
	cfgStore := configstore.New(redis.Client(), &cfgstoreRepo.BackendAdapter{Rows: cfgRows})
	emailCh := notifService.NewEmailChannelFromConfig(cfg.Email).WithStore(cfgStore)
	emailWorker := notifService.NewEmailWorker(
		emailCh, emailLogs,
		notifService.NewAttachmentLoader(database.DB(), objectStore),
		notifService.NewMinuteLimiter(50, nil),
	)
	emailWorker.Start(context.Background())
	emailCh.WithDispatcher(emailWorker)
	channelRegistry := notifService.NewChannelRegistry()
	channelRegistry.Register(notifService.NewInAppChannel(notifR))
	channelRegistry.Register(emailCh)
	channelRegistry.Register(notifService.NewWebSocketChannel(hub))

	tplEngine := notifService.NewTemplateEngine(tplRepo)
	notifSvc := notifService.NewNotificationService(notifR, tplRepo, tplEngine, channelRegistry)
	tplSvc := notifService.NewTemplateService(tplRepo, tplEngine)

	// 事件总线监听器：监听业务事件并自动发送通知
	eventListener := notifService.NewEventListener(tplRepo, tplEngine, channelRegistry)
	eventListener.RegisterAll(eventBus)

	// 通知处理器
	notificationHandler := notifHandler.NewNotificationHandler(notifSvc, hub)
	templateHandler := notifHandler.NewTemplateHandler(tplSvc)
	wsHandler := notifHandler.NewWSHandler(hub, &cfg.JWT, cfg.CORS.AllowedOrigins)
	emailSvc := notifService.NewEmailService(emailWorker, tplEngine, emailLogs)
	emailHandler := notifHandler.NewEmailHandler(emailSvc)

	// 文件管理模块
	fileR := fileRepo.NewFileRepo(database.DB())
	fileSvc := fileService.NewFileService(fileR, objectStore, cacheService, cfg.MinIO.Bucket)
	fileH := fileHandler.NewFileHandler(fileSvc)

	// 入会申请 + 人员档案
	memberAppRepo := memberRepo.NewApplicationRepo(database.DB())
	interviewStarter := memberService.NewInterviewStarter(wfHandlers.DefinitionRepo, wfHandlers.InstanceService)
	if err := wfHandlers.RegisterBusinessApprover("member_application", memberService.NewAdmissionApprover(database.DB())); err != nil {
		logger.Fatal("register admission workflow", zap.Error(err))
	}
	admissionSvc := memberService.NewAdmissionServiceWithWorkflow(database.DB(), wfHandlers.Engine, cacheService)
	schedService.RegisterHandler("admission_maintenance", "检查候补到期并按异议状态转正", admissionSvc.Maintenance)
	memberSvc := memberService.NewMemberService(memberAppRepo, memberProfRepo, interviewStarter, admissionSvc)
	memberH := memberHandler.NewMemberHandler(memberSvc, admissionSvc)

	// 面试管理
	ivSessionRepo := interviewRepo.NewSessionRepo(database.DB())
	ivRecordRepo := interviewRepo.NewInterviewRepo(database.DB())
	ivEvalRepo := interviewRepo.NewEvaluationRepo(database.DB())
	ivNotifier := interviewService.NewNotifier(notifSvc)
	ivSvc := interviewService.NewInterviewService(ivSessionRepo, ivRecordRepo, ivEvalRepo, ivNotifier, memberSvc)
	ivH := interviewHandler.NewInterviewHandler(ivSvc)

	// 会议管理 + 投票
	mtMeetingRepo := meetingRepo.NewMeetingRepo(database.DB())
	mtAgendaRepo := meetingRepo.NewAgendaRepo(database.DB())
	mtAttendeeRepo := meetingRepo.NewAttendeeRepo(database.DB())
	mtVoteRepo := meetingRepo.NewVoteRepo(database.DB())
	mtNotifier := meetingService.NewNotifier(notifSvc)
	mtSvc := meetingService.NewMeetingService(mtMeetingRepo, mtAgendaRepo, mtAttendeeRepo, mtVoteRepo, mtNotifier, database.DB())
	mtH := meetingHandler.NewMeetingHandler(mtSvc)
	schedService.RegisterHandler("meeting_vote_expiry", "按截止时间关闭会议投票", meetingService.NewVoteExpiryJob(database.DB()))

	// 活动管理与报名系统（/activities，#52）
	actActivityRepo := activityRepo.NewActivityRepo(database.DB())
	actRegRepo := activityRepo.NewRegistrationRepo(database.DB())
	actSurveyRepo := activityRepo.NewSurveyRepo(database.DB())
	actNotifier := activityService.NewNotifier(notifSvc)
	actSvc := activityService.NewActivityService(actActivityRepo, actRegRepo, actSurveyRepo, actNotifier, database.DB())
	actH := activityHandler.NewActivityHandler(actSvc)

	// 公告与通知中心（/announcements，#77 phase-1）
	annSvc := announcementService.New(
		announcementRepo.New(database.DB()),
		announcementService.NewNotifier(notifSvc),
	)
	annH := announcementHandler.New(annSvc)
	knH := knowledgeHandler.New(knowledgeService.New(knowledgeRepo.New(database.DB())))
	schedService.RegisterHandler("announcement_scheduled_publish", "扫描并发布到期定时公告", annSvc.DispatchDuePublishes)
	schedService.RegisterHandler("announcement_auto_unpublish", "扫描并归档到期公告", annSvc.DispatchExpired)

	// 请假管理（/leave，#56 phase-1，领域规则来自 #162）
	leaveH := leaveHandler.New(leaveService.New(leaveRepo.New(database.DB())))

	// 工作汇报（/reports，#57 phase-1 列表）
	reportH := reportHandler.New(reportService.New(reportRepo.NewReportRepo(database.DB())))

	// 数据备份与恢复（/system/backups，#88 phase-1）
	backupMinioCfg := cfg.MinIO
	if bucket := strings.TrimSpace(cfg.Backup.Bucket); bucket != "" {
		backupMinioCfg.Bucket = bucket
	}
	backupStore := objectStore
	if backupMinioCfg.Bucket != cfg.MinIO.Bucket {
		dedicated, err := storage.NewMinIO(backupMinioCfg)
		if err != nil {
			logger.Error("init backup MinIO client failed", zap.Error(err))
		} else {
			backupStore = dedicated
			if err := dedicated.EnsureBucket(context.Background()); err != nil {
				logger.Error("ensure backup MinIO bucket failed", zap.Error(err))
			}
		}
	}
	if cfg.Backup.Bucket == "" {
		cfg.Backup.Bucket = cfg.MinIO.Bucket
	}
	backupRows := backupRepo.New(database.DB())
	backupSvc := backupService.New(
		backupRows,
		backupStore,
		cfg.Database,
		cfg.Backup,
		backupService.NewNotifier(notifSvc, backupRows),
	)
	backupH := backupHandler.New(backupSvc)
	schedService.RegisterHandler("backup_scheduled_full", "按策略执行 PostgreSQL 全量备份", backupSvc.RunScheduled)
	schedService.RegisterHandler("backup_retention_cleanup", "按保留天数清理过期备份", backupSvc.CleanupExpired)

	// 运行时业务配置（#47，复用 configs 表，不改 pkg/config YAML）
	cfgSvc := cfgstoreService.NewConfigService(cfgRows, cfgStore).WithSMTP(cfg.Email, emailCh)
	cfgH := cfgstoreHandler.NewConfigHandler(cfgSvc)

	// IT 实习管理
	internRepo := internshipRepo.NewInternshipRepo(database.DB())
	internSvc := internshipService.NewInternshipService(internRepo)
	internH := internshipHandler.NewInternshipHandler(internSvc)

	// 任务流转
	tkTaskRepo := taskRepo.NewTaskRepo(database.DB())
	tkLogRepo := taskRepo.NewLogRepo(database.DB())
	tkCommentRepo := taskRepo.NewCommentRepo(database.DB())
	tkAttachRepo := taskRepo.NewAttachmentRepo(database.DB())
	tkNotifier := taskService.NewNotifier(notifSvc)
	tkSvc := taskService.NewTaskService(tkTaskRepo, tkLogRepo, tkCommentRepo, tkAttachRepo, tkNotifier, fileSvc, objectStore, database.DB())
	if err := wfHandlers.RegisterBusinessApprover("collaboration_task", taskService.NewTaskApprover(database.DB())); err != nil {
		logger.Fatal("register task workflow", zap.Error(err))
	}
	tkSvc.SetWorkflowEngine(wfHandlers.Engine)
	tkH := taskHandler.NewTaskHandler(tkSvc)
	taskReminder := taskService.NewReminderScheduler(tkSvc)
	taskReminder.Start()

	// 日程 / 日历（#78，/schedules；勿与 /system/scheduler 混淆）
	calSvc := scheduleService.New(
		scheduleRepo.New(database.DB()),
		scheduleService.NewNotifier(notifSvc),
		scheduleService.NewActivityFeed(database.DB()),
		scheduleService.NewInterviewFeed(database.DB()),
	)
	calH := scheduleHandler.New(calSvc)
	schedService.RegisterHandler("schedule_reminder", "扫描并推送到期日程提醒", calSvc.DispatchDueReminders)
	schedService.RegisterHandler("schedule_google_sync", "Google 日历同步挂钩（需用户已授权）", calSvc.DispatchGoogleSync)

	// 数据字典
	dictR := dictRepo.NewDictRepository(database.DB())
	dictSvc := dictService.NewDictService(dictR, dictService.NewRedisCache(redis.Client()))
	dictH := dictHandler.NewDictHandler(dictSvc, cacheService)

	// 打印 / 报表导出（#71，Redis 任务 + MinIO 临时文件）
	expRepo := exportRepo.NewRedisRepo(redis.Client())
	expSvc := exportService.NewExportService(expRepo, objectStore, exportService.NewNotifReady(notifSvc))
	expH := exportHandler.NewExportHandler(expSvc)

	// 审计日志模块
	auditR := auditRepo.NewAuditRepo(database.DB())
	auditSvc := auditService.NewAuditServiceWithStore(auditR, &cfg.MinIO, objectStore)
	auditH := auditHandler.NewAuditHandler(auditSvc)
	auditService.RegisterAuthEvents(eventBus, auditSvc)

	// 启动审计日志归档定时任务（每天 02:00 归档 90 天前日志）
	archiveScheduler := auditService.NewArchiveScheduler(auditSvc)
	archiveScheduler.Start()

	// 定时任务调度引擎（#73）
	schedR := schedRepo.New(database.DB())
	schedEng := schedService.NewEngine(schedR, redis.Client(), schedService.NewNotifAlerter(notifSvc))
	schedSvc := schedService.NewService(schedR, schedEng)

	monitorSvc := monitorService.New(database.DB(), redis.Client())

	// 10. API 路由组
	api := r.Group("/api/v1")
	// API 组限流：全局 1000 req/s（#14 固定窗口）+ 令牌桶 IP/接口 + 熔断（#75）
	// /health 不在此组，不受 API 限流与熔断影响
	api.Use(middleware.RateLimit(redis.Client(), middleware.GlobalRateLimit))
	trafficCfg := ratelimit.LoadFromEnv()
	applyAPITraffic(api, redis.Client(), trafficCfg)

	// 10a. 公开路由（不需要鉴权）；IP 令牌桶已挂在 api 组
	public := api.Group("")
	{
		public.GET("/ping", func(c *gin.Context) {
			response.OK(c, "pong")
		})

		// 认证路由（登录、刷新、第三方登录预留）
		// 登录端点额外应用 LoginRateLimit（5 req/min，防暴力破解）
		authHandler.RegisterRoutes(public, nil, authH, middleware.RateLimitWithFallback(redis.Client(), middleware.LoginRateLimit), cacheService)

		// 注册仍由 user handler 处理
		public.POST("/auth/register", userHandler.Register)
		scheduleHandler.RegisterPublicRoutes(public, calH)
		featureHandler.RegisterPublic(public, featureHandler.New(featSvc), &cfg.JWT, redis.Client())
		knowledgeHandler.RegisterPublicRoutes(
			featureHandler.GatePublicKnowledge(public, featSvc, &cfg.JWT, redis.Client()),
			knH, &cfg.JWT, redis.Client(), cacheService,
		)
	}

	// 10b. 需要鉴权的路由
	// 中间件链: AuditLog → JWTAuth → 熔断 → 用户令牌桶（#75）
	// 熔断在用户桶之前：打开时直接 21002，不消耗 rl:uid 配额
	// AuditLog 在 JWTAuth 之前以捕获失败认证尝试
	protected := api.Group("")
	protected.Use(middleware.AuditLog(database.DB()))
	protected.Use(authmiddleware.JWTAuth(&cfg.JWT, redis.Client()))
	applyProtectedTraffic(protected, redis.Client(), trafficCfg, circuitbreaker.New(circuitbreaker.DefaultSettings()))
	{
		// 认证路由（登出、当前用户、修改密码、在线会话 #50）
		authHandler.RegisterRoutes(nil, protected, authH, nil, cacheService)

		// 用户模块
		handler.RegisterUserRoutes(protected, userHandler)

		// RBAC 系统管理模块
		// 权限校验和数据权限中间件在 RegisterRoutes 内部按正确顺序注册
		rbacHandler.RegisterRoutes(protected, database.DB(), roleHandler, permHandler, deptHandler, posHandler, cacheService, deptRepo)

		// 工作流引擎模块
		wfHandler.RegisterRoutes(protected, wfHandlers.Definition, wfHandlers.Instance, wfHandlers.Task, wfHandler.RouteSecurity{DB: database.DB(), Cache: cacheService, Departments: deptRepo})

		// 通知模块路由
		notifHandler.RegisterRoutes(protected, protected, notificationHandler, templateHandler, wsHandler, emailHandler, cacheService)

		// 文件管理模块（/files，file:read / file:create；删除由服务层校验上传者或 file:delete）
		fileHandler.RegisterRoutes(protected, fileH, cacheService)

		// 入会申请 + 人员档案（/member）
		memberHandler.RegisterRoutes(protected, memberH, cacheService, database.DB(), deptRepo)

		// 面试管理（/interviews）
		interviewHandler.RegisterRoutes(protected, ivH, cacheService, database.DB(), deptRepo)

		// 会议管理 + 投票（/meetings, /votes, /system/vote-weight-config）
		meetingHandler.RegisterRoutes(protected, mtH, cacheService, database.DB(), deptRepo)

		// 活动管理与报名系统（/activities）
		activityHandler.RegisterRoutes(protected, actH, cacheService)

		// 公告中心（/announcements，#77）；成员侧信息流受 announcement.feed 灰度
		announcementHandler.RegisterRoutes(protected, annH, cacheService, featureHandler.RequireFlag(
			featSvc, featureModel.KeyAnnouncementFeed, featureHandler.StaffBypass(
				"announcement:create", "announcement:update", "announcement:delete",
				"announcement:publish", "announcement:manage",
			),
		))

		// 知识库 / CMS（/knowledge，#58）
		knowledgeHandler.RegisterRoutes(protected, knH, cacheService)

		// 请假管理（/leave，#56）
		leaveHandler.RegisterRoutes(protected, leaveH, cacheService, database.DB(), deptRepo)

		// 工作汇报（/reports，#57 phase-1 列表）
		reportHandler.RegisterRoutes(protected, reportH, cacheService, database.DB(), deptRepo)

		// 任务流转（/tasks，不与 /workflow/tasks 冲突）
		taskHandler.RegisterRoutes(protected, tkH, cacheService, database.DB(), deptRepo)

		// 日程管理（/schedules，#78）
		scheduleHandler.RegisterRoutes(protected, calH, cacheService, database.DB(), deptRepo)

		// IT 实习管理（/internships, /system/internship-config）
		internshipHandler.RegisterRoutes(protected, internH, cacheService, database.DB(), deptRepo)

		// 运行时配置（/system/configs）
		cfgstoreHandler.RegisterRoutes(protected, cfgH, cacheService)

		// 数据字典（/system/dicts，dict:read/create/update/delete；公开读启用项只需登录）
		dictHandler.RegisterRoutes(protected, dictH, cacheService)

		// 打印 / 报表导出（/export）
		exportHandler.RegisterRoutes(protected, expH, cacheService)

		// 缓存管理（/system/cache，#72）
		cacheAdminSvc := cacheadminService.NewCacheService(redis.Client())
		cacheAdminH := cacheadminHandler.NewCacheHandler(cacheAdminSvc)
		cacheadminHandler.RegisterRoutes(protected, cacheAdminH, cacheService)

		// 运维监控仪表盘（/monitor，#87）
		monitorH := monitorHandler.New(monitorSvc)
		monitorHandler.RegisterRoutes(protected, monitorH, cacheService)

		// 数据备份与恢复（/system/backups，#88 phase-1）
		backupHandler.RegisterRoutes(protected, backupH, cacheService)

		// 定时任务调度（/system/scheduler，#73）
		schedH := schedHandler.NewSchedulerHandler(schedSvc)
		schedHandler.RegisterRoutes(protected, schedH, cacheService)

		// 统一搜索（/system/search，#74）
		searchSvc := searchService.NewSearchService(database.DB())
		searchH := searchHandler.NewSearchHandler(searchSvc, database.DB(), deptRepo, cacheService)
		searchHandler.RegisterRoutes(protected, searchH, cacheService)

		// 数据统计（/stats，#11）
		statsSvc := statsService.NewStatsService(statsRepo.NewStatsRepo(database.DB()))
		statsH := statsHandler.NewStatsHandler(statsSvc)
		statsHandler.RegisterRoutes(protected, statsH, cacheService, database.DB(), deptRepo)

		// 动态表单（/forms，#28）
		formSvc := formService.New(formRepo.New(database.DB()))
		formH := formHandler.NewFormHandler(formSvc, cacheService)
		formHandler.RegisterRoutes(protected, formH, cacheService)

		// 特性开关管理 + 会员门户灰度。公开 CMS 挂在 /knowledge/public，不拦 /auth、CAS、入会申请、/knowledge 编辑。
		featureHandler.RegisterRoutes(protected, featureHandler.New(featSvc), cacheService)
		memberHandler.RegisterPortal(protected, memberH, featureHandler.RequireFlag(featSvc, featureModel.KeyMembershipPortal, nil))

		// 财务 / 处分 / 合同（#22 #23 #24）
		registerPhase1Ops(protected, database.DB(), cacheService, deptRepo, notifSvc, wfHandlers.DefinitionRepo, wfHandlers.InstanceService)

		// 调度引擎在业务 handler（如 contract_expiry）注册后再启动
		if err := backupSvc.SyncSchedule(context.Background()); err != nil {
			logger.Error("sync backup schedule failed", zap.Error(err))
		}
		schedEng.Start()

		// 审计日志模块路由（/system/audit-logs，audit:read / audit:export / audit:archive / audit:report）
		auditHandler.RegisterRoutes(protected, auditH, cacheService)
	}

	// WebSocket 路由（独立于 API 组，JWT 认证在 handler 内部完成）
	notifHandler.RegisterWSRoute(r, wsHandler)
	monitorHandler.RegisterWSRoute(r, monitorHandler.NewWSHandler(monitorSvc, &cfg.JWT, cacheService, cfg.CORS.AllowedOrigins).WithRedis(redis.Client()))

	// 11. 404 处理
	r.NoRoute(func(c *gin.Context) {
		response.NotFound(c, "接口不存在")
	})

	// 12. 启动服务器
	collectCtx, collectCancel := context.WithCancel(context.Background())
	defer collectCancel()
	metrics.StartCollectors(collectCtx, database.DB(), redis.Client())

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      r,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
	}

	// 13. 优雅关闭
	go func() {
		logger.Info("server started", zap.Int("port", cfg.Server.Port), zap.String("mode", cfg.Server.Mode))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("server listen failed", zap.Error(err))
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("server shutting down...")

	// 5秒超时关闭
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("server shutdown error", zap.Error(err))
	}

	// 优雅关闭审计日志 worker，刷新缓冲通道中的待写入条目
	middleware.CloseAuditWriter()

	// 停止审计日志归档定时任务
	archiveScheduler.Stop()
	taskReminder.Stop()
	schedEng.Stop()

	logger.Info("server exited")
}
