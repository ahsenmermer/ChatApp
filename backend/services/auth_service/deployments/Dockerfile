# ---------------- Build stage ----------------
FROM golang:1.25.1-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o auth_service ./cmd/main.go

# ---------------- Final stage ----------------
FROM alpine:3.18

WORKDIR /app

# Optional: timezone
RUN apk add --no-cache tzdata
ENV TZ=UTC

COPY --from=builder /app/auth_service .
COPY internal/migrations ./internal/migrations
COPY internal/config/.env ./internal/config/.env

ENV SERVICE_PORT=8080
ENV MIGRATIONS_PATH=/app/internal/migrations

EXPOSE 8080

ENTRYPOINT ["./auth_service"]
CMD []