package entity

import (
	"time"

	"github.com/google/uuid"
)

type Registration struct {
	UserID       uuid.UUID `json:"user_id" db:"user_id"`
	ConferenceID uuid.UUID `json:"conference_id" db:"conference_id"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`

	User       *User       `json:"-" db:"-"`
	Conference *Conference `json:"-" db:"-"`
}
