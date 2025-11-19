package repository

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"auth_service/internal/models"
)

// UserRepository defines operations we need for users.
type UserRepository interface {
	CreateUser(u *models.User) (*models.User, error)
	GetByEmail(email string) (*models.User, error)
	GetByID(id uuid.UUID) (*models.User, error)
}

// PostgresUserRepository is a Postgres implementation of UserRepository.
type PostgresUserRepository struct {
	db *sqlx.DB
}

// NewPostgresUserRepository creates a new PostgresUserRepository
func NewPostgresUserRepository(db *sqlx.DB) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

// CreateUser inserts a new user into the database.
// Expects u.PasswordHash to be already set (hashed).
func (r *PostgresUserRepository) CreateUser(u *models.User) (*models.User, error) {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}

	// Ensure CreatedAt is set
	if u.CreatedAt.IsZero() {
		u.CreatedAt = time.Now().UTC()
	}

	query := `
		INSERT INTO users (id, username, email, password_hash, created_at)
		VALUES (:id, :username, :email, :password_hash, :created_at)
		RETURNING id, created_at
	`

	// Use NamedQuery to bind struct fields to named params
	rows, err := r.db.NamedQuery(query, u)
	if err != nil {
		return nil, fmt.Errorf("insert user: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		var returnedID uuid.UUID
		var createdAt time.Time
		if err := rows.Scan(&returnedID, &createdAt); err != nil {
			return nil, fmt.Errorf("scan returned id/created_at: %w", err)
		}
		u.ID = returnedID
		u.CreatedAt = createdAt
	}

	return u, nil
}

// GetByEmail fetches a user by email
func (r *PostgresUserRepository) GetByEmail(email string) (*models.User, error) {
	var u models.User
	query := `SELECT id, username, email, password_hash, created_at FROM users WHERE email=$1 LIMIT 1`
	if err := r.db.Get(&u, query, email); err != nil {
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	return &u, nil
}

// GetByID fetches a user by id
func (r *PostgresUserRepository) GetByID(id uuid.UUID) (*models.User, error) {
	var u models.User
	query := `SELECT id, username, email, password_hash, created_at FROM users WHERE id=$1 LIMIT 1`
	if err := r.db.Get(&u, query, id); err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return &u, nil
}
