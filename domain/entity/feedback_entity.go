package entity

import (
	"time"

	"github.com/google/uuid"
)

type Feedback struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	UserID       uuid.UUID  `json:"user_id" db:"user_id"`
	ConferenceID uuid.UUID  `json:"conference_id" db:"conference_id"`
	Comment      string     `json:"comment" db:"comment"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	DeletedAt    *time.Time `json:"deleted_at" db:"deleted_at"`

	User       *User       `json:"-" db:"-"`
	Conference *Conference `json:"-" db:"-"`
}
