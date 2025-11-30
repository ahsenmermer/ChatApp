FROM golang:1.25.1-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o chat_data_service ./cmd/main.go

EXPOSE 8082

CMD ["./chat_data_service"]