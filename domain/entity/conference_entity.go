package entity

import (
	"github.com/google/uuid"
	"github.com/nathakusuma/auditorium-reservation-backend/domain/enum"
	"time"
)

type Conference struct {
	ID             uuid.UUID             `json:"id" db:"id"`
	Title          string                `json:"title" db:"title"`
	Description    string                `json:"description" db:"description"`
	SpeakerName    string                `json:"speaker_name" db:"speaker_name"`
	SpeakerTitle   string                `json:"speaker_title" db:"speaker_title"`
	TargetAudience string                `json:"target_audience" db:"target_audience"`
	Prerequisites  *string               `json:"prerequisites" db:"prerequisites"`
	Seats          int                   `json:"seats" db:"seats"`
	StartsAt       time.Time             `json:"starts_at" db:"starts_at"`
	EndsAt         time.Time             `json:"ends_at" db:"ends_at"`
	HostID         uuid.UUID             `json:"host_id" db:"host_id"`
	Status         enum.ConferenceStatus `json:"status" db:"status"`
	CreatedAt      time.Time             `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time             `json:"updated_at" db:"updated_at"`
	DeletedAt      *time.Time            `json:"deleted_at" db:"deleted_at"`

	Host              User `json:"-" db:"-"`
	RegistrationCount int  `json:"-" db:"-"`
}
