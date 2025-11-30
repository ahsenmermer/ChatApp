FROM golang:1.25.1-alpine AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/api_gateway ./cmd

FROM alpine:3.18
COPY --from=builder /bin/api_gateway /bin/api_gateway
COPY --from=builder /src/internal/config/.env /internal/config/.env

EXPOSE 8085
ENTRYPOINT ["/bin/api_gateway"]