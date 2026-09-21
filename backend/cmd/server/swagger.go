package main

import (
	"net/http"
	"os"

	"github.com/Yogdunana/StarByte/backend/docs"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// swaggerEnabled 报告是否挂载 /swagger 与 /swagger/openapi.json。
//
// fail-closed：只有**显式**声明 APP_ENV=dev|test 才开启。
// 原实现是 `os.Getenv("APP_ENV") != "prod"`，在 APP_ENV 未设置（空串）时返回 true
// —— 而本项目的既定语义是「APP_ENV 空值按生产处理」（见 pkg/config/loader.go 的
// effectiveEnv 与 scripts/seed_users.go 的 seedUsersUseDevPasswords）。
// 两者相反会导致：只按文档补齐 DB_*/JWT_SECRET 等密钥、而未导出 APP_ENV 的部署
// （例如直接 `docker run` 本镜像）把 /swagger/index.html 与完整 OpenAPI 文档
// 暴露在公网 —— 接口清单、参数结构、内部路径一览无余，等于免费给攻击者做侦察。
//
// 与 metricsGuard 的判定保持一致：显式 dev/test 才放宽，空值按生产拒绝。
// 既有流程不受影响：docs/getting-started.md 推荐的启动命令本就带
// `APP_ENV=dev go run ./cmd/server`，`make seed` 也默认带 APP_ENV=dev
// （Makefile:34），生产 compose 则显式设 APP_ENV: prod
// （deploy/docker-compose.yml:105）。未带 APP_ENV 直接裸跑二进制会 404 ——
// 这是刻意的 fail-closed，不是故障。
func swaggerEnabled() bool {
	switch os.Getenv("APP_ENV") {
	case "dev", "test":
		return true
	default:
		return false
	}
}

func registerSwagger(r *gin.Engine) {
	if !swaggerEnabled() {
		return
	}
	ui := ginSwagger.WrapHandler(swaggerFiles.Handler)
	r.GET("/swagger/*any", func(c *gin.Context) {
		any := c.Param("any")
		if any == "/openapi.json" || any == "openapi.json" {
			serveOpenAPI3(c)
			return
		}
		ui(c)
	})
}

func serveOpenAPI3(c *gin.Context) {
	raw, err := swagger2ToOpenAPI3([]byte(docs.SwaggerInfo.ReadDoc()))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "生成 OpenAPI 3.0 失败"})
		return
	}
	c.Data(http.StatusOK, "application/json; charset=utf-8", raw)
}
