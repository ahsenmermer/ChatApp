package models

import "github.com/google/uuid"

// SubscriptionPlan represents a plan like Free, Premium
type SubscriptionPlan struct {
	ID          uuid.UUID `db:"id" json:"id"`
	Name        string    `db:"name" json:"name"`
	Description string    `db:"description" json:"description"`
}
