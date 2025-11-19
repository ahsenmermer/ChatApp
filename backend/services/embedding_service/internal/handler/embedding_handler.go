package handler

import (
	"encoding/json"
	"net/http"
	"os/exec"

	"github.com/gin-gonic/gin"
)

type EmbeddingRequest struct {
	Text string `json:"text"`
}

type EmbeddingResponse struct {
	Embedding []float32 `json:"embedding"`
}

func EmbedHandler(c *gin.Context) {
	var req EmbeddingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Node.js script'i çalıştır
	cmd := exec.Command("node", "/app/internal/embedding/model.js", req.Text)

	output, err := cmd.Output()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// JSON parse
	var resp EmbeddingResponse
	if err := json.Unmarshal(output, &resp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse embedding: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}
