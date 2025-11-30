# -----------------------------
# 1️⃣ Base image
# -----------------------------
FROM golang:1.25.1-alpine AS builder

# Çalışma dizini
WORKDIR /app

# Go mod ve sum kopyala
COPY go.mod go.sum ./

# Modülleri indir
RUN go mod download

# Tüm kaynak kodları kopyala
COPY . .

# Binary oluştur
RUN go build -o chat_service ./cmd/main.go

# -----------------------------
# 2️⃣ Minimal image
# -----------------------------
FROM alpine:3.18

# SSL sertifikaları (OpenRouter API için)
RUN apk add --no-cache ca-certificates

WORKDIR /app

# Binary'yi kopyala
COPY --from=builder /app/chat_service .

# Port
EXPOSE 8080

# Container başlat
CMD ["./chat_service"]