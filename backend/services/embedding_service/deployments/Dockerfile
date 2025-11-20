# =========================
# 1) Node.js dependencies
# =========================
FROM node:18 AS node_builder
WORKDIR /app/internal/embedding

COPY internal/embedding/package*.json ./
RUN npm install

COPY internal/embedding/*.js ./

# Model'i önceden cache'leyelim (opsiyonel ama önerilen)
RUN node -e "import('@xenova/transformers').then(m => m.pipeline('feature-extraction', 'Xenova/all-MiniLM-L6-v2'))" || true

# =========================
# 2) Go build
# =========================
FROM golang:1.25 AS go_builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Statik binary oluştur
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /embedding_service ./cmd/main.go

# =========================
# 3) Final Image
# =========================
FROM node:18-slim
WORKDIR /app

# Node.js scriptleri ve node_modules
COPY --from=node_builder /app/internal/embedding /app/internal/embedding

# Go binary
COPY --from=go_builder /embedding_service /embedding_service
RUN chmod +x /embedding_service

# Environment variables
ENV NODE_PATH=/app/internal/embedding/node_modules
ENV NODE_ENV=production

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=40s \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8400/health || exit 1

EXPOSE 8400

ENTRYPOINT ["/embedding_service"]