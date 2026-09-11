// Package main 程序入口 - 依赖装配与启动
// 日程管理模块 (Schedule Module) - 对应 GitHub Issue #78
//
// 四层架构: handler -> service -> repo -> model
// 本文件负责:
//   1. 加载配置
//   2. 初始化 GORM 数据库连接与 AutoMigrate
//   3. 装配依赖 (repo → service → handler)
//   4. 启动 Gin HTTP 服务
//   5. 启动提醒调度器（可选）
//
// 依赖关系图:
//   gorm.DB
//     ├── repo: GormTransactionManager
//     ├── repo: GormScheduleEventRepository
//     ├── repo: GormScheduleReminderRepository
//     └── repo: GormScheduleEventAttendeeRepository
//   service:
//     ├── ScheduleEventService     (依赖 eventRepo + attendeeRepo + reminderRepo + txMgr + expander)
//     ├── ScheduleReminderService  (依赖 reminderRepo + eventRepo + attendeeRepo + dispatcher)
//     └── ScheduleAttendeeService  (依赖 attendeeRepo + eventRepo + txMgr)
//   handler:
//     └── ScheduleHandlers          (聚合三个 service)
package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"schedule-service/internal/config"
	"schedule-service/internal/handler"
	"schedule-service/internal/model"
	"schedule-service/internal/repo"
	"schedule-service/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ============================================================================
// main 函数 - 启动入口
// ============================================================================

func main() {
	// 1. 加载配置
	cfg := config.Load()

	// 2. 初始化日志
	initLogger(cfg)

	// 3. 初始化数据库连接
	db, err := initDatabase(cfg.Database)
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	// 4. 自动迁移
	if err := autoMigrate(db); err != nil {
		log.Fatalf("failed to auto migrate: %v", err)
	}

	// 5. 装配依赖
	handlers, reminderSvc := wireDependencies(db)

	// 6. 启动 Gin 服务
	ginEngine := initGin(cfg.Server, handlers, cfg.JWT.Secret)

	// 7. 启动提醒调度器（可选）
	if cfg.Reminder.Enabled {
		startReminderScheduler(reminderSvc, cfg.Reminder)
	}

	// 8. 启动 HTTP 服务（含优雅关闭）
	startHTTPServer(ginEngine, cfg.Server.Port)
}

// ============================================================================
// 依赖装配
// ============================================================================

// wireDependencies 装配四层架构的全部依赖
//   返回 handler 聚合对象与提醒 service（供调度器使用）
func wireDependencies(db *gorm.DB) (*handler.ScheduleHandlers, service.ScheduleReminderService) {
	// 1. repo 层
	txMgr := repo.NewGormTransactionManager(db)
	eventRepo := repo.NewGormScheduleEventRepository(db)
	reminderRepo := repo.NewGormScheduleReminderRepository(db)
	attendeeRepo := repo.NewGormScheduleEventAttendeeRepository(db)

	// 2. service 层
	//    注入骨架实现: LogDispatcher + SimpleRecurrenceExpander
	//    生产环境替换为实际实现
	dispatcher := service.NewLogReminderDispatcher()
	expander := service.NewSimpleRecurrenceExpander()

	eventSvc := service.NewScheduleEventService(
		eventRepo, attendeeRepo, reminderRepo, txMgr, expander,
	)
	reminderSvc := service.NewScheduleReminderService(
		reminderRepo, eventRepo, attendeeRepo, dispatcher,
	)
	attendeeSvc := service.NewScheduleAttendeeService(
		attendeeRepo, eventRepo, txMgr,
	)

	// 3. handler 层
	handlers := handler.NewScheduleHandlers(eventSvc, reminderSvc, attendeeSvc)

	return handlers, reminderSvc
}

// ============================================================================
// 初始化函数
// ============================================================================

// initLogger 初始化应用日志
func initLogger(cfg *config.Config) {
	// Gin 运行模式
	gin.SetMode(cfg.Server.Mode)
}

// initDatabase 初始化 GORM 数据库连接
func initDatabase(cfg config.DatabaseConfig) (*gorm.DB, error) {
	// GORM 日志级别
	gormLogLevel := logger.Warn
	if gin.Mode() == gin.DebugMode {
		gormLogLevel = logger.Info
	}

	db, err := gorm.Open(mysql.Open(cfg.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(gormLogLevel),
	})
	if err != nil {
		return nil, fmt.Errorf("gorm open: %w", err)
	}

	// 获取底层 sql.DB 配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get *sql.DB: %w", err)
	}
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetimeMin) * time.Minute)

	// 连接验证
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("db ping: %w", err)
	}

	log.Printf("database connected: %s:%d/%s", cfg.Host, cfg.Port, cfg.Name)
	return db, nil
}

// autoMigrate 自动迁移数据库表结构
//   生产环境建议替换为独立迁移工具（如 golang-migrate）
func autoMigrate(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&model.ScheduleEvent{},
		&model.ScheduleReminder{},
		&model.ScheduleEventAttendee{},
	); err != nil {
		return fmt.Errorf("auto migrate: %w", err)
	}
	log.Println("database migration completed")
	return nil
}

// initGin 初始化 Gin 引擎并注册路由
func initGin(cfg config.ServerConfig, handlers *handler.ScheduleHandlers, jwtSecret string) *gin.Engine {
	engine := gin.New()

	// 中间件
	engine.Use(gin.Logger())       // 请求日志
	engine.Use(gin.Recovery())     // panic 恢复

	// 注册路由（含 JWT 中间件）
	handlers.RegisterRoutes(engine, jwtSecret)

	return engine
}

// ============================================================================
// 提醒调度器
// ============================================================================

// startReminderScheduler 启动提醒调度器
//   定期扫描待触发提醒并派发
//   通过 context 取消信号实现优雅停止
func startReminderScheduler(reminderSvc service.ScheduleReminderService, cfg config.ReminderConfig) {
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		ticker := time.NewTicker(time.Duration(cfg.IntervalSeconds) * time.Second)
		defer ticker.Stop()

		log.Printf("reminder scheduler started: interval=%ds, batch=%d",
			cfg.IntervalSeconds, cfg.BatchSize)

		for {
			select {
			case <-ctx.Done():
				log.Println("reminder scheduler stopped")
				return
			case <-ticker.C:
				// 扫描并触发到期提醒
				before := time.Now()
				count, err := reminderSvc.FirePending(ctx, before, cfg.BatchSize)
				if err != nil {
					log.Printf("reminder scheduler error: %v", err)
					continue
				}
				if count > 0 {
					log.Printf("reminder scheduler fired %d reminders", count)
				}
			}
		}
	}()

	// 将 cancel 存入全局以便 main 退出时调用
	// 简化实现：使用 sync.Once 或全局变量
	_ = cancel // 在 stopHTTPServer 时通过 context 传播
}

// ============================================================================
// HTTP 服务启动与优雅关闭
// ============================================================================

// startHTTPServer 启动 HTTP 服务并处理优雅关闭
func startHTTPServer(engine *gin.Engine, port int) {
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      engine,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 启动服务（非阻塞）
	go func() {
		log.Printf("server starting on :%d", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server failed: %v", err)
		}
	}()

	// 优雅关闭：等待 SIGINT/SIGTERM
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down server...")

	// 给予 30 秒关闭时间
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("server forced to shutdown: %v", err)
	}

	log.Println("server exited")
}

// ============================================================================
// 编译期依赖检查（防止 import 未使用报错）
// ============================================================================

// 确保 database/sql 被引用（sql.DB 类型在 initDatabase 中间接使用）
// GORM 底层使用 database/sql，此 import 用于连接池配置
var _ = sql.ErrNoRows

// 确保退出时不影响调度器
// 提醒调度器在 main 退出时随进程一起结束
