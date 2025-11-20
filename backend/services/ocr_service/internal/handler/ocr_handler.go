package handler

import (
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/otiai10/gosseract/v2"
)

type OCRResponse struct {
	Text string `json:"text"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func UploadHandler(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		log.Printf("❌ File upload error: %v", err)
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Dosya alınamadı: " + err.Error()})
		return
	}

	log.Printf("📄 Received file: %s (size: %d bytes)", file.Filename, file.Size)

	dst := "/tmp/" + file.Filename
	if err := c.SaveUploadedFile(file, dst); err != nil {
		log.Printf("❌ Save file error: %v", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Dosya kaydedilemedi: " + err.Error()})
		return
	}
	defer os.Remove(dst)

	ext := strings.ToLower(filepath.Ext(file.Filename))
	var text string

	client := gosseract.NewClient()
	defer client.Close()

	// Türkçe dilini ayarla
	if err := client.SetLanguage("tur"); err != nil {
		log.Printf("❌ Tesseract language error: %v", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Tesseract dili yüklenemedi: " + err.Error()})
		return
	}

	if ext == ".pdf" {
		log.Printf("📑 Processing PDF file...")
		// PDF → PNG sayfa sayısı kadar
		tmpPattern := "/tmp/page"
		cmd := exec.Command("pdftoppm", dst, tmpPattern, "-png")
		if err := cmd.Run(); err != nil {
			log.Printf("❌ PDF conversion error: %v", err)
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "PDF dönüştürme hatası: " + err.Error()})
			return
		}

		// PNG dosyalarını bul ve sırayla OCR uygula
		files, err := filepath.Glob("/tmp/page*.png")
		if err != nil {
			log.Printf("❌ File glob error: %v", err)
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "PNG dosyaları bulunamadı: " + err.Error()})
			return
		}

		log.Printf("📄 Found %d pages to process", len(files))

		for i, img := range files {
			log.Printf("   Processing page %d/%d...", i+1, len(files))
			client.SetImage(img)
			t, err := client.Text()
			if err != nil {
				log.Printf("❌ OCR error on page %d: %v", i+1, err)
				os.Remove(img)
				c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "OCR hatası: " + err.Error()})
				return
			}
			text += t + "\n"
			os.Remove(img)
		}
	} else {
		log.Printf("🖼️ Processing image file...")
		// Normal resim dosyası
		client.SetImage(dst)
		t, err := client.Text()
		if err != nil {
			log.Printf("❌ OCR error: %v", err)
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "OCR hatası: " + err.Error()})
			return
		}
		text = t
	}

	// Text kontrolü
	text = strings.TrimSpace(text)
	if text == "" {
		log.Printf("⚠️ OCR returned empty text for file: %s", file.Filename)
		c.JSON(http.StatusOK, OCRResponse{Text: ""})
		return
	}

	log.Printf("✅ OCR completed successfully (extracted %d characters)", len(text))
	c.JSON(http.StatusOK, OCRResponse{Text: text})
}
