FROM golang:1.25-alpine AS builder
WORKDIR /app

RUN apk add --no-cache \
    gcc \
    g++ \
    musl-dev \
    tesseract-ocr \
    tesseract-ocr-dev \
    leptonica-dev \
    poppler-utils \
    git

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 go build -o /app/ocr_service ./cmd/main.go

# Runtime stage
FROM alpine:latest
WORKDIR /app

RUN apk add --no-cache \
    tesseract-ocr \
    tesseract-ocr-data-eng \
    tesseract-ocr-data-tur \
    leptonica \
    poppler-utils \
    libstdc++ \
    libgomp \
    ca-certificates

COPY --from=builder /app/ocr_service .

CMD ["/app/ocr_service"]
