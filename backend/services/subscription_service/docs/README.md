Subscription Service
Her kullanıcı, kayıt olduğunda Kafka üzerinden gelen user_registered eventiyle otomatik olarak Free Plan alır.
Servis ayrıca kullanıcı kotasını izler ve chat_completed eventiyle kota düşürür.

Mimari Yapı
subscription_service/
├── cmd/
│ └── main.go # Servis giriş noktası
├── internal/
│ ├── config/ # Ortam değişkenleri & yapılandırma
│ │ ├── .env
│ │ └── config.go
│ ├── database/ # PostgreSQL bağlantısı ve migration yönetimi
│ │ ├── database.go
│ |── migrations/
│ │ ├── 001_create_tables.sql
│ │ └── run_migrations.go
│ ├── handler/ # HTTP endpoint handler’ları
│ │ └── subscription_handler.go
│ ├── migrations/  
│ │ └── 001_create_tables.sql
│ │ └── run_migrations.go
│ ├── models/ # Veritabanı modelleri
│ │ ├── subscription_plan.go
│ │ ├── subscription_quota.go
│ │ └── user_subscription.go
│ ├── repository/ # DB erişim katmanı
│ │ ├── subscription_repository.go
│ ├── router/ # API route tanımları
│ │ └── subscription_router.go
│ ├── services/ # İş mantığı ve Kafka tüketicileri
│ │ ├── user_subscription_service.go
│ └── utils/ # Yardımcı fonksiyonlar
│ └── time.go
| └── uuid.go
├── deployments/
│ └── Dockerfile
└── go.mod / go.sum

Başlangıç
Gereksinimler
Go 1.21+
PostgreSQL 15+
Kafka 3.5+

Kurulum
1️⃣ Bağımlılıkları yükle
go mod download
2️⃣ Veritabanını oluştur
CREATE DATABASE subscription_db;
3️⃣ .env dosyasını ayarla
POSTGRES_HOST=subscription_db
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=1234
POSTGRES_DB=subscription_db

KAFKA_BROKERS=kafka:9092
KAFKA_TOPIC_USER_REGISTERED=user_registered

SERVICE_PORT=8081
LOG_LEVEL=debug
MIGRATIONS_PATH=internal/migrations
4️⃣ Servisi başlat
go run cmd/main.go

API Endpoint’leri
GET /api/subscription/quota/user_id (Kullanıcının kalan kotasını döner)
POST /api/subscription/assign (Belirli planı kullanıcıya manuel atar)
POST /api/subscription/event (Kullanıcının bir aksiyonunu kaydeder (örnek: chat kullandı))

1️⃣ Kullanıcının Kotası (Quota) Sorgulama
curl -X GET http://localhost:8081/api/subscription/quota/59d09c4a-9873-49bd-9508-2cadb8a52393
Response
{
"quota": 1000
}

2️⃣ Kullanıcıya Manuel Plan Atama
curl -X POST http://localhost:8081/api/subscription/assign \
 -H "Content-Type: application/json" \
 -d '{
"user_id": "59d09c4a-9873-49bd-9508-2cadb8a52393",
"plan_id": "1b5c9a6f-9270-4c6a-a8ff-41f3c5b2d2f9",
"start_date": "2025-11-07T00:00:00Z",
"end_date": "2025-12-07T00:00:00Z"
}'
Response
{
"status": "assigned",
"message": "Plan successfully assigned to user"
}

3️⃣ Kullanıcının Bir İşlem Yaptığını Kaydetme (Event)
curl -X POST http://localhost:8081/api/subscription/event \
 -H "Content-Type: application/json" \
 -d '{
"user_id": "59d09c4a-9873-49bd-9508-2cadb8a52393",
"event_type": "chat_used"
}'
Response
{
"status": "ok"
}

4️⃣ Kafka Üzerinden Otomatik Free Plan Atama
Auth Service yeni bir kullanıcı kaydettiğinde şu event’i Kafka’ya yollar 👇
Subscription Service bu mesajı dinler ve kullanıcıya otomatik olarak Free Plan oluşturur.
Kafka Event:
{
"type": "user_registered",
"user_id": "59d09c4a-9873-49bd-9508-2cadb8a52393",
"email": "ahsen@example.com",
"username": "ahsen"
}
Bu event geldiğinde:
Kullanıcıya otomatik olarak Free Plan atanır.
5 günlük süre ve 1000 mesaj kotası başlatılır.
Ayrıca Chat Service’den gelen event:
{
"type": "chat_completed",
"user_id": "59d09c4a-9873-49bd-9508-2cadb8a52393"
}
geldiğinde ilgili kullanıcının kotası 1 azaltılır.

Veritabanı Şeması
CREATE TABLE IF NOT EXISTS subscription_plans (
id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
name VARCHAR(50) NOT NULL,
quota INT NOT NULL,
duration_days INT NOT NULL
);

CREATE TABLE IF NOT EXISTS user_subscriptions (
id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
user_id UUID NOT NULL,
plan_id UUID NOT NULL REFERENCES subscription_plans(id),
start_date TIMESTAMP NOT NULL,
end_date TIMESTAMP NOT NULL,
remaining_quota INT DEFAULT 0
);

Veri Modelleri
subscription_plans
type SubscriptionPlan struct {
ID uuid.UUID `db:"id" json:"id"`
Name string `db:"name" json:"name"`
Quota int `db:"quota" json:"quota"`
DurationDays int `db:"duration_days" json:"duration_days"`
}
subscription_quota
type SubscriptionQuota struct {
ID uuid.UUID `db:"id" json:"id"`
SubscriptionID uuid.UUID `db:"subscription_id" json:"subscription_id"`
Quota int `db:"quota" json:"quota"`
}
user_subscriptions
type UserSubscription struct {
ID uuid.UUID `db:"id" json:"id"`
UserID uuid.UUID `db:"user_id" json:"user_id"`
SubscriptionID uuid.UUID `db:"subscription_id" json:"subscription_id"`
StartDate time.Time `db:"start_date" json:"start_date"`
EndDate time.Time `db:"end_date" json:"end_date"`
RemainingQuota int `db:"remaining_quota" json:"remaining_quota"`
}

Kafka Event Sistemi
Topic: user_registered
Publisher: Auth Service
Subscriber: Subscription Service
Event Format
{
"type": "user_registered",
"user_id": "uuid",
"email": "string",
"username": "string"
}

Konfigürasyon Değişkenleri
| Değişken | Açıklama | Varsayılan |
| ----------------------------- | ----------------------------- | ----------------- |
| `POSTGRES_HOST` | PostgreSQL hostname | `localhost` |
| `POSTGRES_PORT` | PostgreSQL port | `5432` |
| `POSTGRES_USER` | Kullanıcı adı | `postgres` |
| `POSTGRES_PASSWORD` | Şifre | `postgres` |
| `POSTGRES_DB` | Veritabanı adı | `subscription_db` |
| `KAFKA_BROKERS` | Kafka broker adresleri | `localhost:9092` |
| `KAFKA_TOPIC_USER_REGISTERED` | Kullanıcı kayıt event topic’i | `user_registered` |
| `SERVICE_PORT` | Servis portu | `8081` |
| `LOG_LEVEL` | Log seviyesi | `info` |

Mikroservis Entegrasyonu
| Servis | Görev |
| **auth_service** | Kullanıcı kaydı → Kafka’ya event gönderir |
| **subscription_service** | Kafka event’ini alır → Kullanıcıya Free plan atar |
| **api_gateway** | `/api/subscription/*` isteklerini yönlendirir |
| **chat_service** | Kullanıcı sohbet başlattığında event üretir |
| **chat_data_service** | Sohbet geçmişini saklar |

Docker Desteği
Dockerfile
FROM golang:1.25.1-alpine AS builder
WORKDIR /app
RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o subscription_service ./cmd/main.go

FROM alpine:3.18
WORKDIR /app
RUN apk add --no-cache tzdata
ENV TZ=UTC
COPY --from=builder /app/subscription_service .
COPY internal/migrations ./internal/migrations
COPY internal/config/.env ./internal/config/.env
EXPOSE 8081
ENTRYPOINT ["./subscription_service"]
