package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type XenovaClient struct {
	baseURL string
	client  *http.Client
}

func NewXenovaClient(url string) *XenovaClient {
	return &XenovaClient{
		baseURL: url,
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

type embedReq struct {
	Text string `json:"text"`
}

type embedResp struct {
	Embedding []float32 `json:"embedding"`
	Dimension int       `json:"dimension"`
}

func (x *XenovaClient) Embed(text string) ([]float32, error) {
	reqBody := embedReq{Text: text}
	b, _ := json.Marshal(reqBody)

	resp, err := x.client.Post(fmt.Sprintf("%s/embed", x.baseURL), "application/json", bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("xenova request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("xenova responded %d", resp.StatusCode)
	}

	var r embedResp
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, fmt.Errorf("decode failed: %w", err)
	}

	return r.Embedding, nil
}
