Chat Service
Kullanıcı ile OpenRouter AI arasında sohbet etme işlemlerini yürütür.
Ayrıca, konuşma geçmişini MemoryService ile saklar, SubscriptionService üzerinden kota kontrolü yapar
ve sohbet tamamlandığında Kafka aracılığıyla chat_completed event’ini yayınlar.

Mimari Yapı
chat_service/
├── cmd/
│ └── main.go # Servis giriş noktası (Kafka consumer + HTTP server)
├── internal/
│ ├── config/ # Ortam değişkenleri & yapılandırma
│ │ ├── .env
│ │ └── config.go
│ ├── handler/ # HTTP endpoint handler’ları
│ │ └── chat_handler.go
│ ├── models/ # Veri modelleri
│ │ └── chat_message.go
│ ├── repository/ # Kafka producer (event publish)
│ │ └── kafka_producer.go
│ ├── router/ # API route tanımları
│ │ └── router.go
│ ├── services/ # İş mantığı ve dış servislerle entegrasyon
│ │ ├── auth_service_client.go
│ │ ├── chat_service.go
│ │ ├── memory_service.go
│ │ └── subscription_client.go
│ └── utils/ # Yardımcı araçlar
│ └── logger.go
├── deployments/
│ └── Dockerfile
└── go.mod / go.sum

Başlangıç
Gereksinimler
Go 1.21+
Kafka 3.5+
OpenRouter API Key
(Ücretsiz anahtar almak için: https://openrouter.ai/)

Kurulum Adımları
1️⃣ Bağımlılıkları yükle
go mod download
2️⃣ .env dosyasını oluştur
CHAT_SERVICE_PORT=8082
OPENROUTER_KEY=sk-or-xxx
KAFKA_BROKERS=kafka:9092
KAFKA_TOPIC=chat_messages
AUTH_SERVICE_URL=http://auth_service:8080
SUBSCRIPTION_SERVICE_URL=http://subscription_service:8081
3️⃣ Servisi başlat
go run cmd/main.go
4️⃣ Docker üzerinden çalıştırmak için
docker build -t chat_service .
docker run -p 8082:8082 chat_service

API Endpoint’leri
1️⃣ Kullanıcı Chat Başlatma
curl -X POST http://localhost:8082/api/chat \
 -H "Content-Type: application/json" \
 -d '{
"user_id": "f572a9c5-e25f-4d2c-8491-ed2894ceb2a5",
"message": "Merhaba, nasılsın?"
}'
Response
{
"response": "Merhaba! Ben iyiyim, sen nasılsın?"
}

Akış Mantığı
Auth kontrolü:
Kullanıcı kimliği doğrulanır (AuthService).

Abonelik & kota kontrolü:
SubscriptionService üzerinden aktif plan ve kalan kota sorgulanır.

Bellek (MemoryService):
Kullanıcının geçmiş konuşmaları hafızada tutulur (LangChain benzeri).

AI cevabı üretimi:
OpenRouter API çağrılır (nvidia/nemotron-nano-9b-v2:free modeliyle).

Kota azaltma bildirimi:
SubscriptionService'e message_sent event’i POST edilir.

Kafka event publish:
chat_completed event’i Kafka’ya gönderilir → SubscriptionService tarafından dinlenir.

Veri Modelleri
ChatRequest
type ChatRequest struct {
UserID string `json:"user_id"`
Message string `json:"message"`
}

ChatMessage
type ChatMessage struct {
UserID string `json:"user_id"`
Message string `json:"message"`
Response string `json:"response"`
}

Kafka Event Sistemi
Event: chat_completed
Publisher: Publisher
Subscriber: Subscription Service

Event Format
{
"event_type": "chat_completed",
"user_id": "f572a9c5-e25f-4d2c-8491-ed2894ceb2a5",
"message": "Merhaba, nasılsın?",
"response": "İyiyim, teşekkür ederim!",
"timestamp": "2025-11-07T09:21:15Z",
"decrease_quota": true
}

Konfigürasyon Değişkenleri
| Değişken | Açıklama | Varsayılan |
| -------------------------- | -------------------------- | ---------------------------------- |
| `CHAT_SERVICE_PORT` | HTTP portu | `8082` |
| `OPENROUTER_KEY` | OpenRouter API anahtarı | - |
| `KAFKA_BROKERS` | Kafka broker adresleri | `kafka:9092` |
| `KAFKA_TOPIC` | Kafka topic adı | `chat_messages` |
| `AUTH_SERVICE_URL` | Auth Service URL’i | `http://auth_service:8080` |
| `SUBSCRIPTION_SERVICE_URL` | Subscription Service URL’i | `http://subscription_service:8081` |

Mikroservis Entegrasyonu
| Servis | Görev |
| ------------------------ | ----------------------------------------- |
| **auth_service** | Kullanıcı doğrulamasını yapar |
| **subscription_service** | Kota kontrolü ve düşürme |
| **chat_service** | Kullanıcıdan mesaj alır, AI yanıtı üretir |
| **chat_data_service** | Sohbet geçmişini saklar |
| **api_gateway** | `/api/chat` isteklerini yönlendirir |

Dockerfile
FROM golang:1.25.1-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o chat_service ./cmd/main.go

FROM alpine:3.18
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /app/chat_service .
EXPOSE 8082
CMD ["./chat_service"]
