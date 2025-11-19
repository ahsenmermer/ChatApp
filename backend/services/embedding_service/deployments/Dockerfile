# =========================
# 1) Node.js dependencies
# =========================
FROM node:18 AS node_builder
WORKDIR /app/internal/embedding
COPY internal/embedding/package*.json ./
RUN npm install
COPY internal/embedding/*.js ./

# =========================
# 2) Go build (use Go 1.25)
# =========================
FROM golang:1.25 AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Statik binary oluştur
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /embedding_service ./cmd/main.go

# =========================
# 3) Final Image
# =========================
FROM node:18-slim AS final
WORKDIR /app

# Node.js scriptleri
COPY --from=node_builder /app/internal/embedding /app/internal/embedding

# NODE_PATH ayarla (Node.js doğru klasörü görsün)
ENV NODE_PATH=/app/internal/embedding/node_modules

# Go binary ve çalıştırma izni
COPY --from=builder /embedding_service /embedding_service
RUN chmod +x /embedding_service

# Health check için opsiyonel: embedding endpoint’i test edilebilir
HEALTHCHECK --interval=30s --timeout=5s CMD curl -f http://localhost:8400/health || exit 1

EXPOSE 8400
ENTRYPOINT ["/embedding_service"]
