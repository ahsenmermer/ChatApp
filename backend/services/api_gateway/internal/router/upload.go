package router

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"strings"

	"api_gateway/internal/config"
)

type OCRResponse struct {
	Text string `json:"text"`
}

type EmbeddingResponse struct {
	Embedding []float32 `json:"embedding"`
	Dimension int       `json:"dimension"`
}

type UploadSuccessResponse struct {
	Message   string    `json:"message"`
	Text      string    `json:"text"`
	Dimension int       `json:"dimension"`
	Filename  string    `json:"filename"`
	Embedding []float32 `json:"embedding"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

// UploadHandler handles file upload -> OCR -> Embedding
func UploadHandler(cfg config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// --- 0) Dosyayı oku ---
		file, header, err := r.FormFile("file")
		if err != nil {
			log.Printf("❌ Failed to read uploaded file: %v", err)
			respondJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Dosya okunamadı: " + err.Error()})
			return
		}
		defer file.Close()

		const maxFileSize = 20 << 20 // 20 MB
		if header.Size > maxFileSize {
			log.Printf("❌ File too large: %s (%d bytes)", header.Filename, header.Size)
			respondJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Dosya çok büyük (max 20MB)"})
			return
		}

		buf := new(bytes.Buffer)
		if _, err := io.Copy(buf, file); err != nil {
			log.Printf("❌ Failed to copy file content: %v", err)
			respondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "Dosya içeriği okunamadı"})
			return
		}
		log.Printf("📥 Received file: %s, size: %d bytes", header.Filename, header.Size)

		// --- 1) OCR service ---
		ocrURL := cfg.OCRServiceURL + "/upload"
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, err := writer.CreateFormFile("file", header.Filename)
		if err != nil {
			log.Printf("❌ Failed to create multipart writer: %v", err)
			respondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "Multipart oluşturulamadı"})
			return
		}
		if _, err := part.Write(buf.Bytes()); err != nil {
			log.Printf("❌ Failed to write file to multipart: %v", err)
			respondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "Dosya yazılamadı"})
			return
		}
		writer.Close()

		req, err := http.NewRequest("POST", ocrURL, body)
		if err != nil {
			log.Printf("❌ Failed to create OCR request: %v", err)
			respondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "OCR isteği oluşturulamadı"})
			return
		}
		req.Header.Set("Content-Type", writer.FormDataContentType())

		log.Printf("🔄 Sending file to OCR service: %s", ocrURL)
		ocrResp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Printf("❌ OCR request error: %v", err)
			respondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "OCR servisine ulaşılamadı"})
			return
		}
		defer ocrResp.Body.Close()

		ocrBody, err := io.ReadAll(ocrResp.Body)
		if err != nil {
			log.Printf("❌ Failed to read OCR response: %v", err)
			respondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "OCR yanıtı okunamadı"})
			return
		}
		log.Printf("📄 OCR response status: %d, body length: %d", ocrResp.StatusCode, len(ocrBody))

		if ocrResp.StatusCode != http.StatusOK {
			log.Printf("❌ OCR service returned non-OK: %d, body: %s", ocrResp.StatusCode, string(ocrBody))
			respondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "OCR işlemi başarısız"})
			return
		}

		// Parse OCR response
		var ocrResult OCRResponse
		if err := json.Unmarshal(ocrBody, &ocrResult); err != nil {
			log.Printf("❌ Failed to parse OCR JSON: %v, body: %s", err, string(ocrBody))
			respondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "OCR yanıtı parse edilemedi"})
			return
		}

		ocrText := strings.TrimSpace(ocrResult.Text)
		if ocrText == "" {
			log.Printf("⚠️ OCR returned empty text for file: %s", header.Filename)
			respondJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Dosyadan metin çıkarılamadı"})
			return
		}

		log.Printf("✅ OCR completed, extracted %d characters", len(ocrText))

		// --- 2) Embedding service ---
		embedURL := cfg.EmbeddingServiceURL + "/embed"
		embedReq := map[string]string{"text": ocrText}
		embedJSON, err := json.Marshal(embedReq)
		if err != nil {
			log.Printf("❌ Failed to marshal embedding request: %v", err)
			respondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "Embedding isteği oluşturulamadı"})
			return
		}

		log.Printf("🔄 Sending text to embedding service: %s (length: %d)", embedURL, len(ocrText))
		embedResp, err := http.Post(embedURL, "application/json", bytes.NewBuffer(embedJSON))
		if err != nil {
			log.Printf("❌ Embedding request error: %v", err)
			respondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "Embedding servisine ulaşılamadı"})
			return
		}
		defer embedResp.Body.Close()

		embedBody, err := io.ReadAll(embedResp.Body)
		if err != nil {
			log.Printf("❌ Failed to read embedding response: %v", err)
			respondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "Embedding yanıtı okunamadı"})
			return
		}
		log.Printf("🧠 Embedding response status: %d, body length: %d", embedResp.StatusCode, len(embedBody))

		if embedResp.StatusCode != http.StatusOK {
			log.Printf("❌ Embedding service returned non-OK: %d, body: %s", embedResp.StatusCode, string(embedBody))
			respondJSON(w, http.StatusInternalServerError, ErrorResponse{
				Error: fmt.Sprintf("Embedding oluşturulamadı: %s", string(embedBody)),
			})
			return
		}

		// --- 3) Parse embedding JSON ---
		var embedResult EmbeddingResponse
		if err := json.Unmarshal(embedBody, &embedResult); err != nil {
			log.Printf("❌ Failed to parse embedding JSON: %v, body: %s", err, string(embedBody))
			respondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "Embedding yanıtı parse edilemedi"})
			return
		}

		log.Printf("✅ Embedding created successfully, dimension: %d", embedResult.Dimension)

		// --- 4) TODO: Qdrant'a kaydet ---
		// Buraya Qdrant kaydetme kodu eklenecek

		// --- 5) Response ---
		respondJSON(w, http.StatusOK, UploadSuccessResponse{
			Message:   "Dosya başarıyla işlendi",
			Text:      ocrText,
			Dimension: embedResult.Dimension,
			Filename:  header.Filename,
			Embedding: embedResult.Embedding,
		})
		log.Printf("✅ Upload completed successfully for file: %s", header.Filename)
	}
}

// Helper function to respond with JSON
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
