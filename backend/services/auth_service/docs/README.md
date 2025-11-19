Authentication Service
PostgreSQL tabanlı, Kafka ile event-driven çalışan kullanıcı yönetim servisi.
Bu servis kullanıcı kayıt (register) ve giriş (login) işlemlerini yürütür,
kayıt işlemlerinde diğer mikroservisleri bilgilendirmek için Kafka event üretir.

Mimari Yapı
auth_service/
├── cmd/
│ └── main.go # Uygulama başlangıç noktası
├── internal/
│ ├── config/ # Konfigürasyon yönetimi
│ │ ├── .env # Ortam değişkenleri
│ │ └── config.go
│ ├── database/ # PostgreSQL bağlantısı ve migration
│ │ ├── database.go
│ │── migrations/
│ │ ├── 001_create_tables.sql
│ │ └── run_migrations.go
│ ├── handler/ # HTTP endpoint handler'ları
│ │ └── auth_handler.go
│ ├── models/ # Veri modelleri
│ │ └── user.go
│ ├── repository/ # Veritabanı erişim katmanı
│ │ └── user_repository.go
│ ├── router/ # Route tanımlamaları
│ │ └── auth_router.go
│ ├── services/ # İş mantığı ve Kafka event üretimi
│ │ └── user_service.go
│ └── utils/ # Yardımcı fonksiyonlar (şifreleme vb.)
│ └── hash.go
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
CREATE DATABASE auth_db;
3️⃣ .env dosyasını ayarla
POSTGRES_HOST=auth_db
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=1234
POSTGRES_DB=auth_db

KAFKA_BROKERS=kafka:9092
KAFKA_TOPIC_USER_REGISTERED=user_registered

SERVICE_PORT=8080
LOG_LEVEL=debug
MIGRATIONS_PATH=internal/migrations
4️⃣ Servisi başlat
go run cmd/main.go

API Endpoint’leri
POST /api/auth/register (Yeni kullanıcı kaydı)
POST /api/auth/login(Kullanıcı girişi)

Kullanıcı Kaydı
curl -X POST http://localhost:8080/api/auth/register \
 -H "Content-Type: application/json" \
 -d '{
"username": "ahsen",
"email": "ahsen@example.com",
"password": "123456"
}'

Response
{
"id": "59d09c4a-9873-49bd-9508-2cadb8a52393",
"username": "ahsen",
"email": "ahsen@example.com",
"created_at": "2025-11-07T12:00:00Z"
}

Kafka Event
Kullanıcı kaydı başarılı olduğunda aşağıdaki event user_registered topic’ine gönderilir:
{
"type": "user_registered",
"user_id": "59d09c4a-9873-49bd-9508-2cadb8a52393",
"email": "ahsen@example.com",
"username": "ahsen"
}

Kullanıcı Girişi
curl -X POST http://localhost:8080/api/auth/login \
 -H "Content-Type: application/json" \
 -d '{
"email": "ahsen@example.com",
"password": "123456"
}'

Response
{
"id": "59d09c4a-9873-49bd-9508-2cadb8a52393",
"username": "ahsen",
"email": "ahsen@example.com",
"created_at": "2025-11-07T12:00:00Z"
}

Veritabanı Şeması
users Tablosu
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

    CREATE TABLE IF NOT EXISTS users (
        id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
        username VARCHAR(50) UNIQUE NOT NULL,
        email VARCHAR(100) UNIQUE NOT NULL,
        password_hash TEXT NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );

Kafka Event Sistemi
Topic: user_registered
Publisher: Auth Service
Subscriber: Subscription Service (örneğin: kullanıcıya Free plan atar)
Event Format
{
"type": "user_registered",
"user_id": "uuid",
"email": "string",
"username": "string"
}

Konfigürasyon Değişkenleri
| Değişken | Açıklama | Varsayılan |
| ----------------------------- | ------------------------------------------ | --------------------- |
| `POSTGRES_HOST` | PostgreSQL hostname | `localhost` |
| `POSTGRES_PORT` | PostgreSQL port | `5432` |
| `POSTGRES_USER` | PostgreSQL kullanıcı adı | `postgres` |
| `POSTGRES_PASSWORD` | PostgreSQL şifresi | `postgres` |
| `POSTGRES_DB` | Veritabanı adı | `auth_db` |
| `KAFKA_BROKERS` | Kafka broker adresleri (virgülle ayrılmış) | `localhost:9092` |
| `KAFKA_TOPIC_USER_REGISTERED` | Kullanıcı kayıt event topic adı | `user_registered` |
| `SERVICE_PORT` | Servis portu | `8080` |
| `LOG_LEVEL` | Log seviyesi | `info` |
| `MIGRATIONS_PATH` | Migration dosyalarının yolu | `internal/migrations` |

Mikroservis Mimarisi
Bu servis, diğer mikroservislerle birlikte çalışır:
auth_service → Kullanıcı kaydı ve giriş
subscription_service → Kafka üzerinden kullanıcıya Free plan atama
api_gateway → Tüm servisleri tek API noktasında birleştirir
chat_service → Kullanıcı sohbet işlemleri
chat_data_service → Sohbet geçmişi kaydı

Güvenlik Özellikleri
Parola hashleme (bcrypt)
SQL injection koruması (sqlx named params)
Kafka mesaj güvenliği
Basit HTTP routing (net/http)
Docker tabanlı izole orta

Docker Desteği
Dockerfile
FROM golang:1.25.1-alpine AS builder
WORKDIR /app
RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o auth_service ./cmd/main.go

FROM alpine:3.18
WORKDIR /app
RUN apk add --no-cache tzdata
ENV TZ=UTC
COPY --from=builder /app/auth_service .
COPY internal/migrations ./internal/migrations
COPY internal/config/.env ./internal/config/.env
EXPOSE 8080
ENTRYPOINT ["./auth_service"]
