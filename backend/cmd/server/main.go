package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/user/handler"
	"github.com/Yogdunana/StarByte/backend/internal/user/repo"
	"github.com/Yogdunana/StarByte/backend/internal/user/service"
	"github.com/Yogdunana/StarByte/backend/pkg/config"
	"github.com/Yogdunana/StarByte/backend/pkg/database"
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
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger())
	r.Use(middleware.Recovery())
	r.Use(middleware.CORS())

	// 8. 健康检查
	r.GET("/health", func(c *gin.Context) {
		response.OK(c, gin.H{
			"status": "ok",
			"time":   time.Now().Unix(),
		})
	})

	// 9. 初始化业务模块
	// 用户模块
	userRepo := repo.NewUserRepo(database.DB())
	userService := service.NewUserService(userRepo, &cfg.JWT, redis.Client())
	userHandler := handler.NewUserHandler(userService)

	// 10. API 路由组
	api := r.Group("/api/v1")
	{
		// 公开路由（不需要鉴权）
		public := api.Group("")
		{
			// 健康检查
			public.GET("/ping", func(c *gin.Context) {
				response.OK(c, "pong")
			})

			// 认证相关
			handler.RegisterAuthRoutes(public, userHandler)
		}

		// 需要鉴权的路由
		auth := api.Group("")
		auth.Use(authmiddleware.JWTAuth(&cfg.JWT))
		{
			// 用户模块
			handler.RegisterUserRoutes(auth, userHandler)

			// TODO: 注册其他模块路由
			// memberModule.RegisterRoutes(auth)
			// workflowModule.RegisterRoutes(auth)
			// ...
		}
	}

	// 10. 404 处理
	r.NoRoute(func(c *gin.Context) {
		response.NotFound(c, "接口不存在")
	})

	// 11. 启动服务器
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      r,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
	}

	// 12. 优雅关闭
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

	logger.Info("server exited")
}
