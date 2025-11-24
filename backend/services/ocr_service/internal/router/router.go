package router

import (
	"ocr_service/internal/events"
	"ocr_service/internal/handler"

	"github.com/gin-gonic/gin"
)

func SetupRouter(uploadHandler *handler.UploadHandler, producer *events.Producer) *gin.Engine {
	r := gin.Default()

	// CORS middleware
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Upload endpoint
	r.POST("/api/upload", uploadHandler.HandleUpload)

	// File status endpoint
	r.GET("/api/file/status/:file_id", handler.HandleFileStatus)

	// Internal status update endpoint (for embedding service)
	r.POST("/internal/status", func(c *gin.Context) {
		var req struct {
			FileID      string `json:"file_id"`
			Status      string `json:"status"`
			TotalChunks int    `json:"total_chunks"`
			Message     string `json:"message"`
		}

		if err := c.BindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "invalid request"})
			return
		}

		handler.SetFileStatus(req.FileID, req.Status, req.TotalChunks, req.Message)
		c.JSON(200, gin.H{"success": true})
	})

	return r
}
