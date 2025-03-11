package contract

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nathakusuma/auditorium-reservation-backend/domain/dto"
	"github.com/nathakusuma/auditorium-reservation-backend/domain/entity"
)

type IRegistrationRepository interface {
	CreateRegistration(ctx context.Context, registration *entity.Registration) error

	GetRegisteredUsersByConference(ctx context.Context, conferenceID uuid.UUID,
		lazyReq dto.LazyLoadQuery) ([]entity.User, dto.LazyLoadResponse, error)
	GetRegisteredConferencesByUser(ctx context.Context, userID uuid.UUID, includePast bool,
		lazyReq dto.LazyLoadQuery) ([]entity.Conference, dto.LazyLoadResponse, error)

	IsUserRegisteredToConference(ctx context.Context, conferenceID, userID uuid.UUID) (bool, error)
	GetConflictingRegistrations(ctx context.Context, userID uuid.UUID, startsAt,
		endsAt time.Time) ([]entity.Conference, error)
	CountRegistrationsByConference(ctx context.Context, conferenceID uuid.UUID) (int, error)
}

type IRegistrationService interface {
	Register(ctx context.Context, conferenceID, userID uuid.UUID) error

	GetRegisteredUsersByConference(ctx context.Context, conferenceID uuid.UUID,
		lazyReq dto.LazyLoadQuery) ([]dto.UserResponse, dto.LazyLoadResponse, error)
	GetRegisteredConferencesByUser(ctx context.Context, userID uuid.UUID,
		includePast bool, lazyReq dto.LazyLoadQuery) ([]dto.ConferenceResponse, dto.LazyLoadResponse, error)

	IsUserRegisteredToConference(ctx context.Context, conferenceID, userID uuid.UUID) (bool, error)
}
