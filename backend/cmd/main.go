package main

import (
	"fmt"
	"leave-backend/internal/leave/handler"
	"leave-backend/internal/leave/model"
	"leave-backend/internal/leave/repo"
	"leave-backend/internal/leave/service"
	"log"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	dsn := "root:Root@123456@tcp(127.0.0.1:3306)/leave_db?charset=utf8mb4&parseTime=True&loc=Local&allowNativePasswords=true"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	// 自动迁移（开发环境使用，生产环境建议用 migration 工具)
	if err := db.AutoMigrate(
		&model.LeaveType{},
		&model.LeaveBalance{},
		&model.LeaveApplication{},
	); err != nil {
		log.Fatalf("failed to migrate: %v", err)
	}

	// 依赖注入：repo -> service -> handler
	leaveRepo := repo.NewLeaveRepository(db)
	leaveService := service.NewLeaveService(leaveRepo)
	leaveHandler := handler.NewLeaveHandler(leaveService)

	// 启动 HTTP 服务
	r := gin.Default()
	leaveHandler.RegisterRoutes(r)

	fmt.Println("Leave service starting on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
