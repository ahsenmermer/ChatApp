package handler

import (
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/otiai10/gosseract/v2"
)

func UploadHandler(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dosya alınamadı"})
		return
	}

	dst := "/tmp/" + file.Filename
	if err := c.SaveUploadedFile(file, dst); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	var text string

	client := gosseract.NewClient()
	defer client.Close()

	// Türkçe dilini ayarla
	if err := client.SetLanguage("tur"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Tesseract dili yüklenemedi: " + err.Error()})
		return
	}

	if ext == ".pdf" {
		// PDF → PNG sayfa sayısı kadar
		tmpPattern := "/tmp/page"
		cmd := exec.Command("pdftoppm", dst, tmpPattern, "-png")
		if err := cmd.Run(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// PNG dosyalarını bul ve sırayla OCR uygula
		files, err := filepath.Glob("/tmp/page*.png")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		for _, img := range files {
			client.SetImage(img)
			t, err := client.Text()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			text += t + "\n"
			os.Remove(img)
		}
	} else {
		// Normal resim dosyası
		client.SetImage(dst)
		t, err := client.Text()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		text = t
	}

	os.Remove(dst)
	c.JSON(http.StatusOK, gin.H{"text": text})
}
