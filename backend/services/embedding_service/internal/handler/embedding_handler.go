package handler

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os/exec"
	"strings"

	"github.com/gin-gonic/gin"
)

type EmbeddingRequest struct {
	Text string `json:"text"`
}

type EmbeddingResponse struct {
	Embedding []float32 `json:"embedding"`
	Dimension int       `json:"dimension"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func EmbedHandler(c *gin.Context) {
	var req EmbeddingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("❌ JSON binding error: %v", err)
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid JSON: " + err.Error()})
		return
	}

	// Text kontrolü
	text := strings.TrimSpace(req.Text)
	if text == "" {
		log.Printf("❌ Empty text received")
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Text is required"})
		return
	}

	log.Printf("📝 Creating embedding for text (length: %d chars)", len(text))

	// Node.js script'i çalıştır
	cmd := exec.Command("node", "/app/internal/embedding/model.js", text)

	// stdout ve stderr'ı ayrı ayrı yakala
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Komutu çalıştır
	err := cmd.Run()
	if err != nil {
		stderrStr := stderr.String()
		stdoutStr := stdout.String()

		log.Printf("❌ Node.js script failed:")
		log.Printf("   Error: %v", err)
		log.Printf("   Stderr: %s", stderrStr)
		log.Printf("   Stdout: %s", stdoutStr)

		errorMsg := stderrStr
		if errorMsg == "" {
			errorMsg = err.Error()
		}

		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Embedding generation failed: " + errorMsg,
		})
		return
	}

	// stdout'u parse et
	output := stdout.String()
	if output == "" {
		log.Printf("❌ Empty output from Node.js script")
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Empty output from embedding script"})
		return
	}

	log.Printf("✅ Node.js output: %s", output[:min(len(output), 100)])

	// JSON parse
	var result struct {
		Embedding []float32 `json:"embedding"`
		Dimension int       `json:"dimension"`
	}

	if err := json.Unmarshal([]byte(output), &result); err != nil {
		log.Printf("❌ Failed to parse JSON output: %v", err)
		log.Printf("   Output was: %s", output)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to parse embedding result: " + err.Error(),
		})
		return
	}

	log.Printf("✅ Embedding created successfully (dimension: %d)", result.Dimension)

	c.JSON(http.StatusOK, EmbeddingResponse{
		Embedding: result.Embedding,
		Dimension: result.Dimension,
	})
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
