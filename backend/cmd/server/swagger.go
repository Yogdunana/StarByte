package main

import (
	"net/http"
	"os"

	"github.com/Yogdunana/StarByte/backend/docs"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func swaggerEnabled() bool {
	return os.Getenv("APP_ENV") != "prod"
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
