Chat Data Service
Kullanıcıların sohbet geçmişlerini ClickHouse veritabanında saklayan mikroservistir.
Bu servis, Kafka üzerinden gelen chat_completed eventlerini dinler, kullanıcı mesajlarını ve yapay zeka yanıtlarını kaydeder.
Ayrıca /api/chat/history/{user_id} endpoint’i üzerinden geçmiş sorgulama sağlar.

Mimari Yapı
chat_data_service/
├── cmd/
│ └── main.go # Servis giriş noktası
├── internal/
│ ├── config/ # Ortam değişkenleri & yapılandırma
│ │ ├── .env
│ │ └── config.go
│ ├── database/ # ClickHouse bağlantısı (retry mekanizmalı)
│ │ └── database.go
│ ├── migrations/ # Tablo oluşturma scriptleri
│ │ ├── 001_create_tables.sql
│ │ └── run_migrations.go
│ ├── handler/ # HTTP endpoint handler'ları
│ │ └── chat_handler.go
│ ├── models/ # Veritabanı modelleri
│ │ └── chat_message.go
│ ├── repository/ # ClickHouse sorguları
│ │ └── clickhouse_chat_repository.go
│ ├── router/ # HTTP router tanımları
│ │ └── router.go
│ ├── services/ # İş mantığı (mesaj kaydetme & sorgulama)
│ │ └── chat_data_service.go
│ └── utils/ # Yardımcı Kafka consumer (alternatif kullanım)
│ └── kafka_consumer.go
├── deployments/
│ └── Dockerfile # Docker imaj tanımı
└── go.mod / go.sum

Başlangıç
Gereksinimler
Go 1.21+
ClickHouse 23+
Kafka 3.5+

Kurulum Adımları
1️⃣ Bağımlılıkları yükle
go mod download
2️⃣ .env dosyasını ayarla
PORT=8083
CLICKHOUSE_HOST=clickhouse:9000
CLICKHOUSE_DB=chat_data
CLICKHOUSE_USER=default
CLICKHOUSE_PASSWORD=
KAFKA_BROKERS=kafka:9092
KAFKA_TOPIC=chat_messages
3️⃣ ClickHouse tabloyu oluştur
CREATE TABLE IF NOT EXISTS chat_messages (
user_id String,
user_message String,
ai_response String,
timestamp DateTime
)
ENGINE = MergeTree()
ORDER BY (user_id, timestamp);
4️⃣ Servisi başlat
go run cmd/main.go

Servis İşleyişi
1-Chat Service bir kullanıcı mesaj gönderdiğinde Kafka’ya event yollar:
{
"event_type": "chat_completed",
"user_id": "b87b5011-65a6-4fb1-9aaf-08bdaac91358",
"message": "Merhaba!",
"response": "Merhaba, sana nasıl yardımcı olabilirim?",
"timestamp": "2025-11-07T09:00:00Z"
}

2-Chat Data Service bu mesajı Kafka’dan dinler, ClickHouse’a kaydeder.
3-Kullanıcı geçmişini görüntülemek için:
GET /api/chat/history/{user_id}

API Endpoint’leri
1️⃣ Kullanıcı Mesaj Geçmişini Getir
curl -X GET "http://localhost:8083/api/chat/history/b87b5011-65a6-4fb1-9aaf-08bdaac91358?limit=10"
Response
[
{
"user_id": "b87b5011-65a6-4fb1-9aaf-08bdaac91358",
"user_message": "Merhaba!",
"ai_response": "Merhaba, sana nasıl yardımcı olabilirim?",
"timestamp": "2025-11-07T09:00:00Z"
}
]

Veri Modeli
chat_messages
| Alan | Tip | Açıklama |
| ------------ | -------- | ------------------------------ |
| user_id | String | Kullanıcının benzersiz kimliği |
| user_message | String | Kullanıcının gönderdiği mesaj |
| ai_response | String | Yapay zekanın cevabı |
| timestamp | DateTime | Mesajın gönderilme zamanı |

Go Model:
type ChatMessage struct {
UserID string `ch:"user_id" json:"user_id"`
UserMessage string `ch:"user_message" json:"user_message"`
AIResponse string `ch:"ai_response" json:"ai_response"`
Timestamp time.Time `ch:"timestamp" json:"timestamp"`
}

Kafka Event Sistemi
Event Türü: chat_completed
Publisher: Chat Service
Subscriber: Chat Data Service

Event Format:
{
"event_type": "chat_completed",
"user_id": "uuid",
"message": "string",
"response": "string",
"timestamp": "RFC3339"
}

Konfigürasyon Değişkenleri
| Değişken | Açıklama | Varsayılan |
| --------------------- | -------------------------- | ----------------- |
| `PORT` | Servis portu | `8083` |
| `CLICKHOUSE_HOST` | ClickHouse bağlantı adresi | `clickhouse:9000` |
| `CLICKHOUSE_DB` | ClickHouse veritabanı adı | `chat_data` |
| `CLICKHOUSE_USER` | Kullanıcı adı | `default` |
| `CLICKHOUSE_PASSWORD` | Şifre | (boş) |
| `KAFKA_BROKERS` | Kafka broker adresleri | `kafka:9092` |
| `KAFKA_TOPIC` | Kafka topic adı | `chat_messages` |

Mikroservis Entegrasyonu
| Servis | Görev |
| ------------------------ | ---------------------------------------------- |
| **auth_service** | Kullanıcı kayıtlarını Kafka’ya gönderir |
| **subscription_service** | Kullanıcı planlarını yönetir |
| **chat_service** | Kullanıcı mesajlarını işler ve Kafka’ya yollar |
| **chat_data_service** | Mesaj geçmişini ClickHouse’da saklar |
| **api_gateway** | `/api/chat/history/*` isteklerini yönlendirir |

Docker Desteği
Dockerfile
FROM golang:1.25.1-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o chat_data_service ./cmd/main.go

EXPOSE 8083
CMD ["./chat_data_service"]
