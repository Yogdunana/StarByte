package main

import (
	"os"

	_ "github.com/Yogdunana/StarByte/backend/docs"
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
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}
