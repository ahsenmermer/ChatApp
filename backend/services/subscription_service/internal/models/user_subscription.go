package models

import (
	"time"

	"github.com/google/uuid"
)

type UserSubscription struct {
	ID             uuid.UUID `db:"id" json:"id"`
	UserID         uuid.UUID `db:"user_id" json:"user_id"`
	SubscriptionID uuid.UUID `db:"subscription_id" json:"subscription_id"`
	StartDate      time.Time `db:"start_date" json:"start_date"`
	EndDate        time.Time `db:"end_date" json:"end_date"`
	RemainingQuota int       `db:"remaining_quota" json:"remaining_quota"`
}
