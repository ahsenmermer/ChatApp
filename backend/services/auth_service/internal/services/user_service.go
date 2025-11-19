package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/IBM/sarama"
	"github.com/google/uuid"

	"auth_service/internal/models"
	"auth_service/internal/repository"
	"auth_service/internal/utils"
)

type UserService struct {
	userRepo     repository.UserRepository
	KafkaBrokers []string
	KafkaTopic   string
}

func NewUserService(
	userRepo repository.UserRepository,
	kafkaBrokers []string,
	kafkaTopic string,
) *UserService {
	return &UserService{
		userRepo:     userRepo,
		KafkaBrokers: kafkaBrokers,
		KafkaTopic:   kafkaTopic,
	}
}

// RegisterUser handles user registration logic
func (s *UserService) RegisterUser(username, email, password string) (*models.User, error) {
	// 1. Check if user already exists
	existingUser, _ := s.userRepo.GetByEmail(email)
	if existingUser != nil {
		return nil, errors.New("user with this email already exists")
	}

	// 2. Hash the password
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// 3. Create user model
	user := &models.User{
		ID:           uuid.New(),
		Username:     username,
		Email:        email,
		PasswordHash: hashedPassword,
		CreatedAt:    time.Now().UTC(),
	}

	// 4. Insert into DB
	createdUser, err := s.userRepo.CreateUser(user)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// 5. Publish Kafka event
	if err := s.publishUserRegisteredEvent(createdUser); err != nil {
		log.Printf("⚠️ Failed to publish Kafka event: %v\n", err)
	}

	log.Printf("✅ User registered successfully: %s\n", createdUser.Email)
	return createdUser, nil
}

// LoginUser validates user credentials
func (s *UserService) LoginUser(email, password string) (*models.User, error) {
	user, err := s.userRepo.GetByEmail(email)
	if err != nil || user == nil {
		return nil, errors.New("invalid email or password")
	}

	if !utils.CheckPasswordHash(password, user.PasswordHash) {
		return nil, errors.New("invalid email or password")
	}

	return user, nil
}

// publishUserRegisteredEvent publishes Kafka event for other services
func (s *UserService) publishUserRegisteredEvent(user *models.User) error {
	if len(s.KafkaBrokers) == 0 {
		return fmt.Errorf("Kafka brokers not configured")
	}

	config := sarama.NewConfig()
	config.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer(s.KafkaBrokers, config)
	if err != nil {
		return fmt.Errorf("failed to create Kafka producer: %w", err)
	}
	defer producer.Close()

	event := map[string]interface{}{
		"type":     "user_registered",
		"user_id":  user.ID.String(),
		"email":    user.Email,
		"username": user.Username,
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	msg := &sarama.ProducerMessage{
		Topic: s.KafkaTopic,
		Value: sarama.ByteEncoder(payload),
	}

	partition, offset, err := producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to send Kafka message: %w", err)
	}

	log.Printf("📤 Event sent to Kafka topic %s [partition:%d, offset:%d]\n", s.KafkaTopic, partition, offset)
	return nil
}
