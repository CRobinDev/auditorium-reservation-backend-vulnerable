package entity

import (
	"github.com/nathakusuma/auditorium-reservation-backend/domain/enum"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID     `json:"id"`
	Name         string        `json:"name"`
	Email        string        `json:"email"`
	PasswordHash string        `json:"-" db:"password_hash"`
	Role         enum.UserRole `json:"role"`
	Bio          *string       `json:"bio"`
	CreatedAt    time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at" db:"updated_at"`
	DeletedAt    *time.Time    `json:"-" db:"deleted_at"`
}
