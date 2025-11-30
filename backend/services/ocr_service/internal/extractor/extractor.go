package extractor

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"ocr_service/internal/models"

	"github.com/google/uuid"
	"github.com/otiai10/gosseract/v2"
)

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func ProcessFile(path string) ([]models.Chunk, error) {
	if !fileExists(path) {
		return nil, fmt.Errorf("file not found: %s", path)
	}

	ext := strings.ToLower(filepath.Ext(path))

	log.Printf("📄 Processing file: %s (type: %s)", filepath.Base(path), ext)

	if ext == ".pdf" {
		return processPDF(path)
	}

	// Image formats
	validImageExts := []string{".png", ".jpg", ".jpeg", ".tiff", ".tif", ".bmp"}
	for _, validExt := range validImageExts {
		if ext == validExt {
			return processImage(path)
		}
	}

	return nil, fmt.Errorf("unsupported file type: %s", ext)
}

func processPDF(path string) ([]models.Chunk, error) {
	log.Printf("📄 Processing PDF: %s", filepath.Base(path))

	tmpDir, err := os.MkdirTemp("", "ocr_pages")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	outPattern := filepath.Join(tmpDir, "page")

	cmd := exec.Command("pdftoppm", "-png", "-r", "300", path, outPattern)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("pdftoppm failed: %v, stderr: %s", err, stderr.String())
	}

	files, err := filepath.Glob(filepath.Join(tmpDir, "page*.png"))
	if err != nil {
		return nil, fmt.Errorf("failed to glob pages: %w", err)
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no pages extracted from PDF")
	}

	log.Printf("📊 Extracted %d pages from PDF", len(files))

	client := gosseract.NewClient()
	defer client.Close()

	// ✅ Türkçe + İngilizce birlikte
	if err := client.SetLanguage("tur+eng"); err != nil {
		log.Printf("⚠️ Turkish+English not available, trying Turkish only")
		if err := client.SetLanguage("tur"); err != nil {
			log.Printf("⚠️ Turkish not available, using English")
			client.SetLanguage("eng")
		}
	}

	// ✅ PSM 3: Fully automatic page segmentation
	client.SetPageSegMode(gosseract.PSM_AUTO)

	var chunks []models.Chunk
	totalText := 0

	for i, img := range files {
		if err := client.SetImage(img); err != nil {
			log.Printf("⚠️ Failed to set image for page %d: %v", i+1, err)
			continue
		}

		text, err := client.Text()
		if err != nil {
			log.Printf("⚠️ OCR failed on page %d: %v", i+1, err)
			continue
		}

		text = strings.TrimSpace(text)
		if text == "" {
			log.Printf("⚠️ Page %d: no text extracted", i+1)
			continue
		}

		totalText += len(text)
		page := i + 1

		subChunks := splitText(text, 1000)
		log.Printf("✅ Page %d: extracted %d characters, created %d chunks", page, len(text), len(subChunks))

		for idx, sc := range subChunks {
			chunks = append(chunks, models.Chunk{
				ChunkID:     uuid.New().String(),
				Text:        sc,
				Page:        page,
				StartOffset: idx * 1000,
				EndOffset:   idx*1000 + len(sc),
				Metadata: map[string]interface{}{
					"source_type": "pdf",
					"page_number": page,
					"chunk_index": idx,
					"total_pages": len(files),
				},
			})
		}
	}

	log.Printf("✅ PDF processing complete: %d total chunks, %d total characters", len(chunks), totalText)
	return chunks, nil
}

// ✅ TEK processImage fonksiyonu (güncellenmiş versiyon)
func processImage(path string) ([]models.Chunk, error) {
	log.Printf("🖼️ Processing image: %s", filepath.Base(path))

	client := gosseract.NewClient()
	defer client.Close()

	// ✅ Türkçe + İngilizce birlikte
	if err := client.SetLanguage("tur+eng"); err != nil {
		log.Printf("⚠️ Turkish+English not available, trying Turkish only")
		if err := client.SetLanguage("tur"); err != nil {
			log.Printf("⚠️ Turkish not available, using English")
			client.SetLanguage("eng")
		}
	}

	// ✅ PSM 3: Fully automatic
	client.SetPageSegMode(gosseract.PSM_AUTO)

	if err := client.SetImage(path); err != nil {
		return nil, fmt.Errorf("failed to set image: %w", err)
	}

	text, err := client.Text()
	if err != nil {
		return nil, fmt.Errorf("OCR failed: %w", err)
	}

	text = strings.TrimSpace(text)
	if text == "" {
		log.Printf("⚠️ No text extracted from image")
		return nil, nil
	}

	log.Printf("✅ Extracted %d characters from image", len(text))

	parts := splitText(text, 1000)
	var chunks []models.Chunk

	for idx, p := range parts {
		chunks = append(chunks, models.Chunk{
			ChunkID:     uuid.New().String(),
			Text:        p,
			Page:        1,
			StartOffset: idx * 1000,
			EndOffset:   idx*1000 + len(p),
			Metadata: map[string]interface{}{
				"source_type": "image",
				"chunk_index": idx,
			},
		})
	}

	log.Printf("✅ Image processing complete: %d chunks created", len(chunks))
	return chunks, nil
}

func splitText(s string, maxChars int) []string {
	if len(s) <= maxChars {
		return []string{s}
	}

	var chunks []string

	// Önce cümlelere böl
	sentences := strings.Split(s, ". ")

	var current strings.Builder

	for _, sentence := range sentences {
		sentence = strings.TrimSpace(sentence)
		if sentence == "" {
			continue
		}

		// Eğer bu cümleyi eklemek limiti aşacaksa
		if current.Len()+len(sentence)+2 > maxChars && current.Len() > 0 {
			chunks = append(chunks, strings.TrimSpace(current.String()))
			current.Reset()
		}

		if current.Len() > 0 {
			current.WriteString(". ")
		}
		current.WriteString(sentence)
	}

	// Son chunk'ı ekle
	if current.Len() > 0 {
		chunks = append(chunks, strings.TrimSpace(current.String()))
	}

	return chunks
}
