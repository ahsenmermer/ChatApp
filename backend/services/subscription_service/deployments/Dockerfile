# ---------------- Build stage ----------------
FROM golang:1.25.1-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o subscription_service ./cmd/main.go

# ---------------- Final stage ----------------
FROM alpine:3.18

WORKDIR /app

# Optional: timezone
RUN apk add --no-cache tzdata
ENV TZ=UTC

# Copy Go binary and migrations
COPY --from=builder /app/subscription_service .
COPY internal/migrations ./internal/migrations
COPY internal/config/.env ./internal/config/.env

ENV SERVICE_PORT=8081
ENV MIGRATIONS_PATH=/app/internal/migrations

EXPOSE 8081

ENTRYPOINT ["./subscription_service"]
CMD []
