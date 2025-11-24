FROM golang:1.25.1-alpine AS builder
WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /embedding_service ./cmd/main.go

FROM alpine:3.18
WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY --from=builder /embedding_service /embedding_service

EXPOSE 8400
ENTRYPOINT ["/embedding_service"]