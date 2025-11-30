package main

import (
	"log"
	"net/http"

	"github.com/joho/godotenv"

	"auth_service/internal/config"
	"auth_service/internal/database"
	"auth_service/internal/handler"
	"auth_service/internal/migrations"
	"auth_service/internal/repository"
	"auth_service/internal/router"
	"auth_service/internal/services"
)

func main() {
	// .env yükle
	_ = godotenv.Load()

	// Config
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	log.Printf("Config loaded: pg=%s:%d db=%s kafka=%v port=%s\n",
		cfg.PostgresHost, cfg.PostgresPort, cfg.PostgresDB, cfg.KafkaBrokers, cfg.ServicePort)

	// DB bağlantısı
	if err := database.Connect(cfg); err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer database.Close()

	// Migration (migrations.Run fonksiyonu artık path parametresi alıyor)
	if err := migrations.Run(cfg.MigrationsPath); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	// Repositories
	userRepo := repository.NewPostgresUserRepository(database.DB)

	// Services
	userService := services.NewUserService(userRepo, cfg.KafkaBrokers, cfg.KafkaTopicUserRegistered)

	// Auth handler
	authHandler := handler.NewAuthHandler(userService)

	// Router
	mux := http.NewServeMux()
	router.SetupAuthRoutes(mux, authHandler)

	log.Printf("✅ Auth Service running on :%s", cfg.ServicePort)
	if err := http.ListenAndServe(":"+cfg.ServicePort, mux); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
