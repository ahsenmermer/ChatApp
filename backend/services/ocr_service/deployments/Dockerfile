FROM golang:1.25-alpine AS builder
WORKDIR /app

# Build için gerekli paketler
RUN apk add --no-cache \
    gcc \
    g++ \
    musl-dev \
    tesseract-ocr \
    tesseract-ocr-dev \
    leptonica-dev \
    git \
    poppler-utils 

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 go build -o /app/ocr_service ./cmd/main.go

# Runtime stage
FROM alpine:latest
WORKDIR /app

# Runtime için sadece gerekli kütüphaneler
RUN apk add --no-cache \
    tesseract-ocr \
    leptonica \
    libstdc++ \
    libgomp \
    ca-certificates \
    poppler-utils \
    wget \
    unzip

# Türkçe traineddata indir
RUN mkdir -p /usr/share/tessdata \
    && wget -O /usr/share/tessdata/tur.traineddata https://github.com/tesseract-ocr/tessdata/raw/main/tur.traineddata

ENV TESSDATA_PREFIX=/usr/share/tessdata/


COPY --from=builder /app/ocr_service /app/ocr_service

CMD ["/app/ocr_service"]