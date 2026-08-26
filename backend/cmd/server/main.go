package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	rbacHandler "github.com/Yogdunana/StarByte/backend/internal/rbac/handler"
	rbacRepo "github.com/Yogdunana/StarByte/backend/internal/rbac/repo"
	rbacService "github.com/Yogdunana/StarByte/backend/internal/rbac/service"
	"github.com/Yogdunana/StarByte/backend/internal/user/handler"
	"github.com/Yogdunana/StarByte/backend/internal/user/repo"
	"github.com/Yogdunana/StarByte/backend/internal/user/service"
	"github.com/Yogdunana/StarByte/backend/internal/workflow"
	wfHandler "github.com/Yogdunana/StarByte/backend/internal/workflow/handler"
	"github.com/Yogdunana/StarByte/backend/pkg/config"
	"github.com/Yogdunana/StarByte/backend/pkg/database"
	"github.com/Yogdunana/StarByte/backend/pkg/events"
	"github.com/Yogdunana/StarByte/backend/pkg/logger"
	"github.com/Yogdunana/StarByte/backend/pkg/middleware"
	authmiddleware "github.com/Yogdunana/StarByte/backend/pkg/middleware/auth"
	"github.com/Yogdunana/StarByte/backend/pkg/redis"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

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

	logger.Info("server starting...")

	// 3. 初始化数据库
	if err := database.Init(&cfg.Database); err != nil {
		logger.Fatal("init database failed", zap.Error(err))
	}
	defer database.Close()

	// 3a. 自动迁移审计日志表
	if err := database.DB().AutoMigrate(&middleware.AuditLogEntry{}); err != nil {
		logger.Fatal("auto migrate audit_logs failed", zap.Error(err))
	}

	// 4. 初始化 Redis
	if err := redis.Init(&cfg.Redis); err != nil {
		logger.Fatal("init redis failed", zap.Error(err))
	}
	defer redis.Close()

	// 5. 设置 Gin 模式
	gin.SetMode(cfg.Server.Mode)

	// 6. 创建 Gin 引擎
	r := gin.New()

	// 7. 注册全局中间件
	// 顺序: RequestID → Logger → ErrorHandler → CORS
	// 注意: 全局限流不放在全局中间件中，以避免影响健康检查端点
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger())
	r.Use(middleware.ErrorHandler())
	r.Use(middleware.CORSWithConfig(cfg.CORS))

	// 8. 健康检查端点（不受限流影响，供 K8s/负载均衡探活使用）
	r.GET("/health", middleware.HealthCheck())
	r.GET("/health/ready", middleware.ReadinessCheck(database.DB(), redis.Client()))

	// 9. 初始化业务模块
	// 用户模块
	userRepo := repo.NewUserRepo(database.DB())
	userService := service.NewUserService(database.DB(), userRepo, &cfg.JWT)
	userHandler := handler.NewUserHandler(userService)

	// RBAC 权限模块
	roleRepo := rbacRepo.NewRoleRepo(database.DB())
	permRepo := rbacRepo.NewPermissionRepo(database.DB())
	deptRepo := rbacRepo.NewDepartmentRepo(database.DB())
	posRepo := rbacRepo.NewPositionRepo(database.DB())

	cacheService := rbacService.NewPermissionCacheService(database.DB(), redis.Client(), permRepo, roleRepo)

	roleService := rbacService.NewRoleService(database.DB(), roleRepo, permRepo, cacheService)
	permService := rbacService.NewPermissionService(database.DB(), permRepo, cacheService)
	deptService := rbacService.NewDepartmentService(database.DB(), deptRepo)
	posService := rbacService.NewPositionService(database.DB(), posRepo)

	roleHandler := rbacHandler.NewRoleHandler(roleService)
	permHandler := rbacHandler.NewPermissionHandler(permService)
	deptHandler := rbacHandler.NewDepartmentHandler(deptService)
	posHandler := rbacHandler.NewPositionHandler(posService)

	// 工作流引擎模块
	eventBus := events.NewEventBus()
	wfHandlers := workflow.Init(database.DB(), eventBus, logger.GetLogger())

	// 10. API 路由组
	api := r.Group("/api/v1")
	// API 组限流：全局 1000 req/s
	api.Use(middleware.RateLimit(redis.Client(), middleware.GlobalRateLimit))

	// 10a. 公开路由（不需要鉴权）
	// 中间件: PerIPRateLimit (100 req/min per IP)
	public := api.Group("")
	public.Use(middleware.RateLimit(redis.Client(), middleware.PerIPRateLimit))
	{
		public.GET("/ping", func(c *gin.Context) {
			response.OK(c, "pong")
		})

		// 认证相关路由
		authGroup := public.Group("/auth")
		{
			// 注册：PerIPRateLimit 已在 group 级别生效
			authGroup.POST("/register", userHandler.Register)
			authGroup.POST("/refresh", userHandler.RefreshToken)

			// 登录端点：额外限流（5 req/min，防暴力破解）
			authGroup.POST("/login",
				middleware.RateLimitWithFallback(redis.Client(), middleware.LoginRateLimit),
				userHandler.Login,
			)
		}
	}

	// 10b. 需要鉴权的路由
	// 中间件链: AuditLog → JWTAuth → PerIPRateLimit
	// AuditLog 在 JWTAuth 之前以捕获失败认证尝试
	protected := api.Group("")
	protected.Use(middleware.AuditLog(database.DB()))
	protected.Use(authmiddleware.JWTAuth(&cfg.JWT))
	protected.Use(middleware.RateLimit(redis.Client(), middleware.PerIPRateLimit))
	{
		// 登出（需要鉴权）
		protected.POST("/auth/logout", userHandler.Logout)

		// 用户模块
		handler.RegisterUserRoutes(protected, userHandler)

		// RBAC 系统管理模块
		// 权限校验和数据权限中间件在 RegisterRoutes 内部按正确顺序注册
		rbacHandler.RegisterRoutes(protected, database.DB(), roleHandler, permHandler, deptHandler, posHandler, cacheService, deptRepo)

		// 工作流引擎模块
		wfHandler.RegisterRoutes(protected, wfHandlers.Definition, wfHandlers.Instance, wfHandlers.Task)
	}

	// 11. 404 处理
	r.NoRoute(func(c *gin.Context) {
		response.NotFound(c, "接口不存在")
	})

	// 12. 启动服务器
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

	logger.Info("server exited")
}
