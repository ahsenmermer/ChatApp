package router

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"mime/multipart"
	"net/http"

	"api_gateway/internal/config"
)

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
			http.Error(w, "failed to read file: "+err.Error(), http.StatusBadRequest)
			log.Printf("Failed to read uploaded file: %v", err)
			return
		}
		defer file.Close()

		const maxFileSize = 20 << 20 // 20 MB
		if header.Size > maxFileSize {
			http.Error(w, "file too large", http.StatusBadRequest)
			log.Printf("File too large: %s (%d bytes)", header.Filename, header.Size)
			return
		}

		buf := new(bytes.Buffer)
		if _, err := io.Copy(buf, file); err != nil {
			http.Error(w, "failed to read file content: "+err.Error(), http.StatusInternalServerError)
			log.Printf("Failed to copy file content: %v", err)
			return
		}
		log.Printf("Received file: %s, size: %d bytes", header.Filename, header.Size)

		// --- 1) OCR service ---
		ocrURL := cfg.OCRServiceURL + "/upload"
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, err := writer.CreateFormFile("file", header.Filename)
		if err != nil {
			http.Error(w, "failed to create multipart: "+err.Error(), http.StatusInternalServerError)
			log.Printf("Failed to create multipart writer: %v", err)
			return
		}
		if _, err := part.Write(buf.Bytes()); err != nil {
			http.Error(w, "failed to write file to multipart: "+err.Error(), http.StatusInternalServerError)
			log.Printf("Failed to write file to multipart: %v", err)
			return
		}
		writer.Close()

		req, err := http.NewRequest("POST", ocrURL, body)
		if err != nil {
			http.Error(w, "failed to create OCR request: "+err.Error(), http.StatusInternalServerError)
			log.Printf("Failed to create OCR request: %v", err)
			return
		}
		req.Header.Set("Content-Type", writer.FormDataContentType())

		ocrResp, err := http.DefaultClient.Do(req)
		if err != nil {
			http.Error(w, "OCR request failed: "+err.Error(), http.StatusInternalServerError)
			log.Printf("OCR request error: %v", err)
			return
		}
		defer ocrResp.Body.Close()

		ocrBody, err := io.ReadAll(ocrResp.Body)
		if err != nil {
			http.Error(w, "failed to read OCR response: "+err.Error(), http.StatusInternalServerError)
			log.Printf("Failed to read OCR response: %v", err)
			return
		}
		log.Printf("OCR response status: %d, body length: %d", ocrResp.StatusCode, len(ocrBody))

		if ocrResp.StatusCode != http.StatusOK {
			http.Error(w, "OCR service failed: "+string(ocrBody), http.StatusInternalServerError)
			log.Printf("OCR service returned non-OK: %d, body: %s", ocrResp.StatusCode, string(ocrBody))
			return
		}

		ocrText := string(ocrBody)
		if ocrText == "" {
			http.Error(w, "OCR returned empty text", http.StatusInternalServerError)
			log.Printf("OCR returned empty text for file: %s", header.Filename)
			return
		}

		// --- 2) Embedding service ---
		embedURL := cfg.EmbeddingServiceURL + "/embed"
		jsonBody, _ := json.Marshal(map[string]string{"text": ocrText})

		log.Printf("Sending text to embedding service, length: %d", len(ocrText))
		embedResp, err := http.Post(embedURL, "application/json", bytes.NewBuffer(jsonBody))
		if err != nil {
			http.Error(w, "Embedding request failed: "+err.Error(), http.StatusInternalServerError)
			log.Printf("Embedding request error: %v", err)
			return
		}
		defer embedResp.Body.Close()

		embedBody, err := io.ReadAll(embedResp.Body)
		if err != nil {
			http.Error(w, "failed to read embedding response: "+err.Error(), http.StatusInternalServerError)
			log.Printf("Failed to read embedding response: %v", err)
			return
		}
		log.Printf("Embedding response status: %d, body length: %d", embedResp.StatusCode, len(embedBody))

		if embedResp.StatusCode != http.StatusOK {
			http.Error(w, "Embedding service failed: "+string(embedBody), http.StatusInternalServerError)
			log.Printf("Embedding service returned non-OK: %d, body: %s", embedResp.StatusCode, string(embedBody))
			return
		}

		// --- 3) Parse embedding JSON ---
		var embedData struct {
			Embedding []float32 `json:"embedding"`
		}
		if err := json.Unmarshal(embedBody, &embedData); err != nil {
			http.Error(w, "failed to parse embedding JSON: "+err.Error(), http.StatusInternalServerError)
			log.Printf("Failed to parse embedding JSON: %v", err)
			return
		}

		// --- 4) Response ---
		resp := map[string]interface{}{
			"text":      ocrText,
			"embedding": embedData.Embedding,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		log.Printf("UploadHandler completed successfully for file: %s", header.Filename)
	}
}
