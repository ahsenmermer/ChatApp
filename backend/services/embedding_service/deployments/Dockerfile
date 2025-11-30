# =========================
# Stage 1: Node.js Dependencies
# =========================
FROM node:18-slim AS node_builder

WORKDIR /app/embedding

# Copy package files
COPY internal/embedding/package.json internal/embedding/package-lock.json ./

# Install dependencies
RUN npm ci --omit=dev

# Copy JavaScript files
COPY internal/embedding/*.js ./

# =========================
# Stage 2: Go Build
# =========================
FROM golang:1.25.1-alpine AS go_builder

RUN apk add --no-cache git build-base

WORKDIR /app

# Copy Go dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build Go binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s" \
    -o /embedding_service ./cmd/main.go

# =========================
# Stage 3: Final Runtime
# =========================
FROM node:18-slim

# Install curl for healthcheck
RUN apt-get update && apt-get install -y --no-install-recommends \
    curl \
    wget \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy Node.js app from node_builder
COPY --from=node_builder /app/embedding ./internal/embedding

# Copy Go binary from go_builder
COPY --from=go_builder /embedding_service /app/embedding_service
RUN chmod +x /app/embedding_service

# Environment variables
ENV NODE_PATH=/app/internal/embedding/node_modules
ENV PORT=3000
ENV EMBEDDING_SERVICE_PORT=8400

# Health check
HEALTHCHECK --interval=15s --timeout=5s --start-period=60s --retries=3 \
    CMD curl -f http://localhost:8400/health && wget -qO- http://localhost:3000/health || exit 1

EXPOSE 3000 8400

# Start both services
CMD sh -c "node /app/internal/embedding/server.js & /app/embedding_service"