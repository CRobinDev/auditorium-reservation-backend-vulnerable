package contract

import (
	"context"
	"github.com/google/uuid"
	"github.com/nathakusuma/auditorium-reservation-backend/domain/dto"
	"github.com/nathakusuma/auditorium-reservation-backend/domain/entity"
)

type IFeedbackRepository interface {
	CreateFeedback(ctx context.Context, feedback *entity.Feedback) error
	GetFeedbacksByConferenceID(ctx context.Context, conferenceID uuid.UUID,
		lazyReq dto.LazyLoadQuery) ([]entity.Feedback, dto.LazyLoadResponse, error)
	DeleteFeedback(ctx context.Context, id uuid.UUID) error
	IsFeedbackGiven(ctx context.Context, userID, conferenceID uuid.UUID) (bool, error)
}

type IFeedbackService interface {
	CreateFeedback(ctx context.Context, userID, conferenceID uuid.UUID, comment string) (uuid.UUID, error)
	GetFeedbacksByConferenceID(ctx context.Context, conferenceID uuid.UUID,
		lazyReq dto.LazyLoadQuery) ([]dto.FeedbackResponse, dto.LazyLoadResponse, error)
	DeleteFeedback(ctx context.Context, id uuid.UUID) error
}
