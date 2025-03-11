package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/nathakusuma/auditorium-reservation-backend/domain/entity"
	"github.com/nathakusuma/auditorium-reservation-backend/domain/enum"
)

type ConferenceResponse struct {
	ID             uuid.UUID             `json:"id"`
	Title          string                `json:"title,omitempty"`
	Description    string                `json:"description,omitempty"`
	SpeakerName    string                `json:"speaker_name,omitempty"`
	SpeakerTitle   string                `json:"speaker_title,omitempty"`
	TargetAudience string                `json:"target_audience,omitempty"`
	Prerequisites  *string               `json:"prerequisites,omitempty"`
	Seats          int                   `json:"seats,omitempty"`
	StartsAt       *time.Time            `json:"starts_at,omitempty"`
	EndsAt         *time.Time            `json:"ends_at,omitempty"`
	Host           *UserResponse         `json:"host,omitempty"`
	Status         enum.ConferenceStatus `json:"status,omitempty"`
	CreatedAt      *time.Time            `json:"created_at,omitempty"`
	UpdatedAt      *time.Time            `json:"updated_at,omitempty"`
	SeatsTaken     *int                  `json:"seats_taken,omitempty"`
}

func (c *ConferenceResponse) PopulateFromEntity(conference *entity.Conference) *ConferenceResponse {
	c.ID = conference.ID
	c.Title = conference.Title
	c.Description = conference.Description
	c.SpeakerName = conference.SpeakerName
	c.SpeakerTitle = conference.SpeakerTitle
	c.TargetAudience = conference.TargetAudience
	c.Prerequisites = conference.Prerequisites
	c.Seats = conference.Seats
	c.StartsAt = &conference.StartsAt
	c.EndsAt = &conference.EndsAt
	c.Status = conference.Status
	c.CreatedAt = &conference.CreatedAt
	c.UpdatedAt = &conference.UpdatedAt

	c.SeatsTaken = &conference.RegistrationCount
	c.Host = new(UserResponse).PopulateMinimalFromEntity(&conference.Host)
	return c
}

type CreateConferenceProposalRequest struct {
	Title          string
	Description    string
	SpeakerName    string
	SpeakerTitle   string
	TargetAudience string
	Prerequisites  *string
	Seats          int
	StartsAt       time.Time
	EndsAt         time.Time
}

type GetConferenceQuery struct {
	AfterID      *uuid.UUID
	BeforeID     *uuid.UUID
	Limit        int
	HostID       *uuid.UUID
	Status       enum.ConferenceStatus
	StartsBefore *time.Time
	StartsAfter  *time.Time
	IncludePast  bool
	OrderBy      string
	Order        string
	Title        *string
}

type UpdateConferenceRequest struct {
	Title          *string
	Description    *string
	SpeakerName    *string
	SpeakerTitle   *string
	TargetAudience *string
	Prerequisites  *string
	StartsAt       *time.Time
	EndsAt         *time.Time
}

func (p *UpdateConferenceRequest) GenerateUpdateEntity(original *entity.Conference) *entity.Conference {
	if p.Title != nil {
		original.Title = *p.Title
	}
	if p.Description != nil {
		original.Description = *p.Description
	}
	if p.SpeakerName != nil {
		original.SpeakerName = *p.SpeakerName
	}
	if p.SpeakerTitle != nil {
		original.SpeakerTitle = *p.SpeakerTitle
	}
	if p.TargetAudience != nil {
		original.TargetAudience = *p.TargetAudience
	}
	if p.Prerequisites != nil {
		original.Prerequisites = p.Prerequisites
	}
	if p.StartsAt != nil {
		original.StartsAt = *p.StartsAt
	}
	if p.EndsAt != nil {
		original.EndsAt = *p.EndsAt
	}

	return original
}
