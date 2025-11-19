package repository

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"subscription_service/internal/models"
)

type SubscriptionRepository interface {
	GetPlanByName(name string) (*models.SubscriptionPlan, error)
	GetQuotaByPlanID(planID uuid.UUID) (*models.SubscriptionQuota, error)
	AssignPlanToUser(userID uuid.UUID, planID uuid.UUID, start, end time.Time) (*models.UserSubscription, error)
	AssignPlanToUserWithQuota(userID uuid.UUID, planID uuid.UUID, start, end time.Time, quota int) (*models.UserSubscription, error)
	GetUserSubscription(userID uuid.UUID) (*models.UserSubscription, error)
}

type PostgresSubscriptionRepository struct {
	db *sqlx.DB
}

func NewPostgresSubscriptionRepository(db *sqlx.DB) *PostgresSubscriptionRepository {
	return &PostgresSubscriptionRepository{db: db}
}

func (r *PostgresSubscriptionRepository) GetPlanByName(name string) (*models.SubscriptionPlan, error) {
	var plan models.SubscriptionPlan
	query := `SELECT id, name, description FROM subscription_plans WHERE name=$1 LIMIT 1`
	if err := r.db.Get(&plan, query, name); err != nil {
		return nil, fmt.Errorf("get plan by name: %w", err)
	}
	return &plan, nil
}

func (r *PostgresSubscriptionRepository) GetQuotaByPlanID(planID uuid.UUID) (*models.SubscriptionQuota, error) {
	var quota models.SubscriptionQuota
	query := `SELECT id, subscription_id, quota FROM subscription_quotas WHERE subscription_id = $1 LIMIT 1`
	if err := r.db.Get(&quota, query, planID); err != nil {
		return nil, fmt.Errorf("get quota by plan id: %w", err)
	}
	return &quota, nil
}

func (r *PostgresSubscriptionRepository) AssignPlanToUser(userID, planID uuid.UUID, start, end time.Time) (*models.UserSubscription, error) {
	return r.AssignPlanToUserWithQuota(userID, planID, start, end, 0)
}

func (r *PostgresSubscriptionRepository) AssignPlanToUserWithQuota(userID, planID uuid.UUID, start, end time.Time, quota int) (*models.UserSubscription, error) {
	sub := &models.UserSubscription{
		ID:             uuid.New(),
		UserID:         userID,
		SubscriptionID: planID,
		StartDate:      start,
		EndDate:        end,
		RemainingQuota: quota,
	}

	query := `
		INSERT INTO user_subscription (id, user_id, subscription_id, start_date, end_date, remaining_quota)
		VALUES (:id, :user_id, :subscription_id, :start_date, :end_date, :remaining_quota)
		RETURNING id, start_date, end_date, remaining_quota
	`

	rows, err := r.db.NamedQuery(query, sub)
	if err != nil {
		return nil, fmt.Errorf("assign plan to user: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		var id uuid.UUID
		var s, e time.Time
		var rq int
		if err := rows.Scan(&id, &s, &e, &rq); err != nil {
			return nil, fmt.Errorf("scan user_subscription: %w", err)
		}
		sub.ID = id
		sub.StartDate = s
		sub.EndDate = e
		sub.RemainingQuota = rq
	}

	return sub, nil
}

func (r *PostgresSubscriptionRepository) GetUserSubscription(userID uuid.UUID) (*models.UserSubscription, error) {
	var sub models.UserSubscription
	query := `SELECT id, user_id, subscription_id, start_date, end_date, remaining_quota 
			  FROM user_subscription 
			  WHERE user_id=$1 AND end_date >= NOW() LIMIT 1`
	if err := r.db.Get(&sub, query, userID); err != nil {
		return nil, fmt.Errorf("get user subscription: %w", err)
	}
	return &sub, nil
}
