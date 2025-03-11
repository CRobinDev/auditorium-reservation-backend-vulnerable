package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/nathakusuma/auditorium-reservation-backend/domain/entity"
	"github.com/nathakusuma/auditorium-reservation-backend/domain/enum"
)

type ConferenceJoinUserRow struct {
	ID             uuid.UUID             `db:"id"`
	Title          string                `db:"title"`
	Description    string                `db:"description"`
	SpeakerName    string                `db:"speaker_name"`
	SpeakerTitle   string                `db:"speaker_title"`
	TargetAudience string                `db:"target_audience"`
	Prerequisites  *string               `db:"prerequisites"`
	Seats          int                   `db:"seats"`
	StartsAt       time.Time             `db:"starts_at"`
	EndsAt         time.Time             `db:"ends_at"`
	HostID         uuid.UUID             `db:"host_id"`
	Status         enum.ConferenceStatus `db:"status"`
	CreatedAt      time.Time             `db:"created_at"`
	UpdatedAt      time.Time             `db:"updated_at"`

	HostName          string `db:"host_name"`
	RegistrationCount int    `db:"registration_count"`
}

func (r *ConferenceJoinUserRow) ToEntity() entity.Conference {
	return entity.Conference{
		ID:             r.ID,
		Title:          r.Title,
		Description:    r.Description,
		SpeakerName:    r.SpeakerName,
		SpeakerTitle:   r.SpeakerTitle,
		TargetAudience: r.TargetAudience,
		Prerequisites:  r.Prerequisites,
		Seats:          r.Seats,
		StartsAt:       r.StartsAt,
		EndsAt:         r.EndsAt,
		HostID:         r.HostID,
		Status:         r.Status,
		CreatedAt:      r.CreatedAt,
		UpdatedAt:      r.UpdatedAt,
		Host: entity.User{
			ID:   r.HostID,
			Name: r.HostName,
		},
		RegistrationCount: r.RegistrationCount,
	}
}
