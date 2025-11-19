package router

import (
	"ocr_service/internal/handler"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()
	r.POST("/upload", handler.UploadHandler)
	return r
}
