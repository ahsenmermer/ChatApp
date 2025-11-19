package models

import "github.com/google/uuid"

// SubscriptionQuota stores quota for a plan
type SubscriptionQuota struct {
	ID             uuid.UUID `db:"id" json:"id"`
	SubscriptionID uuid.UUID `db:"subscription_id" json:"subscription_id"`
	Quota          int       `db:"quota" json:"quota"`
}
