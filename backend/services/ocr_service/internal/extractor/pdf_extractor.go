package extractor

import (
	"bytes"
	"fmt"
	"io/ioutil"
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
	if ext == ".pdf" {
		return processPDF(path)
	}
	return processImage(path)
}

func processPDF(path string) ([]models.Chunk, error) {
	tmpDir, err := ioutil.TempDir("", "ocr_pages")
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

	client := gosseract.NewClient()
	defer client.Close()

	if err := client.SetLanguage("tur"); err != nil {
		client.SetLanguage("eng")
	}

	var chunks []models.Chunk
	for i, img := range files {
		if err := client.SetImage(img); err != nil {
			return nil, fmt.Errorf("failed to set image: %w", err)
		}

		text, err := client.Text()
		if err != nil {
			return nil, fmt.Errorf("OCR failed on page %d: %w", i+1, err)
		}

		text = strings.TrimSpace(text)
		if text == "" {
			continue
		}

		page := i + 1
		subChunks := splitText(text, 1000)
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
				},
			})
		}
	}

	return chunks, nil
}

func processImage(path string) ([]models.Chunk, error) {
	client := gosseract.NewClient()
	defer client.Close()

	if err := client.SetLanguage("tur"); err != nil {
		client.SetLanguage("eng")
	}

	if err := client.SetImage(path); err != nil {
		return nil, fmt.Errorf("failed to set image: %w", err)
	}

	text, err := client.Text()
	if err != nil {
		return nil, fmt.Errorf("OCR failed: %w", err)
	}

	text = strings.TrimSpace(text)
	if text == "" {
		return nil, nil
	}

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
	return chunks, nil
}

func splitText(s string, maxChars int) []string {
	if len(s) <= maxChars {
		return []string{s}
	}

	var chunks []string
	sentences := strings.Split(s, ". ")

	var current strings.Builder
	for _, sentence := range sentences {
		if current.Len()+len(sentence)+2 > maxChars && current.Len() > 0 {
			chunks = append(chunks, strings.TrimSpace(current.String()))
			current.Reset()
		}
		if current.Len() > 0 {
			current.WriteString(". ")
		}
		current.WriteString(sentence)
	}

	if current.Len() > 0 {
		chunks = append(chunks, strings.TrimSpace(current.String()))
	}

	return chunks
}
