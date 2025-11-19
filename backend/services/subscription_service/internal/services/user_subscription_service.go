package services

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/IBM/sarama"
	"github.com/google/uuid"

	"subscription_service/internal/database"
	"subscription_service/internal/models"
	"subscription_service/internal/repository"
)

type UserSubscriptionService struct {
	subRepo      repository.SubscriptionRepository
	KafkaBrokers []string
	KafkaTopic   string
}

// NewUserSubscriptionService constructs a new service
func NewUserSubscriptionService(repo repository.SubscriptionRepository, brokers []string, topic string) *UserSubscriptionService {
	return &UserSubscriptionService{
		subRepo:      repo,
		KafkaBrokers: brokers,
		KafkaTopic:   topic,
	}
}

// AssignFreePlanToUser assigns Free plan to a user using DB quota
func (s *UserSubscriptionService) AssignFreePlanToUser(userID uuid.UUID) error {
	plan, err := s.subRepo.GetPlanByName("Free")
	if err != nil {
		return fmt.Errorf("failed to get Free plan: %w", err)
	}

	quotaObj, err := s.subRepo.GetQuotaByPlanID(plan.ID)
	if err != nil {
		return fmt.Errorf("failed to get Free plan quota: %w", err)
	}
	quota := quotaObj.Quota

	start := time.Now().UTC()
	end := start.AddDate(0, 0, 5) // Free plan 5 gün

	_, err = s.subRepo.AssignPlanToUserWithQuota(userID, plan.ID, start, end, quota)
	if err != nil {
		return fmt.Errorf("failed to assign Free plan: %w", err)
	}

	log.Printf("✅ Free plan assigned to user %s with %d quota", userID.String(), quota)
	return nil
}

// AssignPlanToUserByName assigns a specific plan to a user using DB quota
func (s *UserSubscriptionService) AssignPlanToUserByName(userID uuid.UUID, planName string) error {
	plan, err := s.subRepo.GetPlanByName(planName)
	if err != nil {
		return fmt.Errorf("plan not found: %w", err)
	}

	quotaObj, err := s.subRepo.GetQuotaByPlanID(plan.ID)
	if err != nil {
		return fmt.Errorf("failed to get plan quota: %w", err)
	}
	quota := quotaObj.Quota

	start := time.Now().UTC()
	end := start.AddDate(0, 0, 30) // Örn: 30 gün

	_, err = s.subRepo.AssignPlanToUserWithQuota(userID, plan.ID, start, end, quota)
	if err != nil {
		return fmt.Errorf("failed to assign plan: %w", err)
	}

	log.Printf("✅ Plan %s assigned to user %s with %d quota", planName, userID.String(), quota)
	return nil
}

// GetUserQuota returns remaining quota of a user
func (s *UserSubscriptionService) GetUserQuota(userID uuid.UUID) (int, error) {
	sub, err := s.subRepo.GetUserSubscription(userID)
	if err != nil {
		return 0, err
	}
	return sub.RemainingQuota, nil
}

// LogEventAndDecrementQuota logs user events and decreases quota
func (s *UserSubscriptionService) LogEventAndDecrementQuota(userIDStr string) error {
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return fmt.Errorf("invalid user_id: %w", err)
	}

	sub, err := s.subRepo.GetUserSubscription(userID)
	if err != nil {
		return fmt.Errorf("get user subscription: %w", err)
	}

	if sub.RemainingQuota <= 0 {
		return fmt.Errorf("quota exceeded")
	}

	sub.RemainingQuota -= 1
	query := `UPDATE user_subscription SET remaining_quota=$1 WHERE id=$2`
	if _, err := database.DB.Exec(query, sub.RemainingQuota, sub.ID); err != nil {
		return fmt.Errorf("decrement quota: %w", err)
	}

	return nil
}

// ListAllPlans lists all subscription plans
func (s *UserSubscriptionService) ListAllPlans() ([]models.SubscriptionPlan, error) {
	query := `SELECT id, name, description FROM subscription_plans`
	rows, err := database.DB.Queryx(query)
	if err != nil {
		return nil, fmt.Errorf("list all plans: %w", err)
	}
	defer rows.Close()

	var plans []models.SubscriptionPlan
	for rows.Next() {
		var p models.SubscriptionPlan
		if err := rows.StructScan(&p); err != nil {
			return nil, fmt.Errorf("scan plan: %w", err)
		}
		plans = append(plans, p)
	}

	return plans, nil
}

// StartKafkaConsumer starts a Kafka consumer to handle registration and chat events
func (s *UserSubscriptionService) StartKafkaConsumer() error {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	config.Version = sarama.V2_8_0_0

	master, err := sarama.NewConsumer(s.KafkaBrokers, config)
	if err != nil {
		return fmt.Errorf("failed to start Kafka consumer: %w", err)
	}

	consumer, err := master.ConsumePartition(s.KafkaTopic, 0, sarama.OffsetNewest)
	if err != nil {
		return fmt.Errorf("failed to consume partition: %w", err)
	}

	log.Printf("🎧 Kafka consumer started for topic: %s", s.KafkaTopic)

	go func() {
		for msg := range consumer.Messages() {
			log.Printf("📥 Kafka message received: %s", string(msg.Value))

			var event struct {
				Type    string `json:"type"`
				UserID  string `json:"user_id"`
				Email   string `json:"email,omitempty"`
				Message string `json:"message,omitempty"`
			}
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				log.Printf("⚠️ Failed to unmarshal event: %v", err)
				continue
			}

			switch event.Type {
			case "user_registered":
				// Auth service’den gelen event → Free plan ata
				uid, err := uuid.Parse(event.UserID)
				if err != nil {
					log.Printf("⚠️ Invalid UUID in event: %v", err)
					continue
				}
				if err := s.AssignFreePlanToUser(uid); err != nil {
					log.Printf("⚠️ Failed to assign Free plan: %v", err)
				}

			case "chat_completed":
				// Chat service’den gelen event → Kota azalt
				log.Printf("💬 Chat completed for user %s, decreasing quota...", event.UserID)
				if err := s.LogEventAndDecrementQuota(event.UserID); err != nil {
					log.Printf("⚠️ Failed to decrease quota: %v", err)
				}

			default:
				log.Printf("ℹ️ Ignored unknown event type: %s", event.Type)
			}
		}
	}()

	return nil
}
