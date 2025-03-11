package contract

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nathakusuma/auditorium-reservation-backend/domain/dto"
	"github.com/nathakusuma/auditorium-reservation-backend/domain/entity"
	"github.com/nathakusuma/auditorium-reservation-backend/domain/enum"
)

type IConferenceService interface {
	CreateConferenceProposal(ctx context.Context, req *dto.CreateConferenceProposalRequest) (uuid.UUID, error)
	GetConferenceByID(ctx context.Context, id uuid.UUID) (*dto.ConferenceResponse, error)
	GetConferences(ctx context.Context,
		query *dto.GetConferenceQuery) ([]dto.ConferenceResponse, dto.LazyLoadResponse, error)
	UpdateConference(ctx context.Context, id uuid.UUID, req dto.UpdateConferenceRequest) error
	DeleteConference(ctx context.Context, id uuid.UUID) error

	UpdateConferenceStatus(ctx context.Context, id uuid.UUID, status enum.ConferenceStatus) error
}

type IConferenceRepository interface {
	CreateConference(ctx context.Context, conference *entity.Conference) error
	GetConferenceByID(ctx context.Context, id uuid.UUID) (*entity.Conference, error)
	GetConferences(ctx context.Context, query *dto.GetConferenceQuery) ([]entity.Conference, dto.LazyLoadResponse, error)
	UpdateConference(ctx context.Context, conference *entity.Conference) error
	DeleteConference(ctx context.Context, id uuid.UUID) error

	GetConferencesConflictingWithTime(ctx context.Context, startsAt, endsAt time.Time,
		excludeID uuid.UUID) ([]entity.Conference, error)
}
