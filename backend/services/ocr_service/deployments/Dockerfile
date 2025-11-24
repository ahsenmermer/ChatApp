FROM golang:1.25.1-alpine AS builder
WORKDIR /app

# Alpine için Tesseract ve poppler kurulumu
RUN apk add --no-cache \
    poppler-utils \
    tesseract-ocr \
    tesseract-ocr-data-tur \
    git \
    leptonica-dev \
    tesseract-ocr-dev \
    gcc \
    g++ \
    musl-dev 

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 go build -o /ocr_service ./cmd/main.go

FROM alpine:3.18
WORKDIR /app

# Runtime dependencies
RUN apk add --no-cache \
    poppler-utils \
    tesseract-ocr \
    leptonica \
    libstdc++ \
    libgomp \
    tesseract-ocr-data-tur \
    ca-certificates \
    wget \
    unzip

# Türkçe traineddata indir
RUN mkdir -p /usr/share/tessdata \
    && wget -O /usr/share/tessdata/tur.traineddata https://github.com/tesseract-ocr/tessdata/raw/main/tur.traineddata

ENV TESSDATA_PREFIX=/usr/share/tessdata

COPY --from=builder /ocr_service /ocr_service

EXPOSE 8090
ENTRYPOINT ["/ocr_service"]