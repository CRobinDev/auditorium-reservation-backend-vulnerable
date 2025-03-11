package entity

import (
	"github.com/google/uuid"
	"time"
)

type AuthSession struct {
	Token     string    `json:"token"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
}
