package service

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/nathakusuma/auditorium-reservation-backend/domain/contract"
	"github.com/nathakusuma/auditorium-reservation-backend/domain/dto"
	"github.com/nathakusuma/auditorium-reservation-backend/domain/entity"
	"github.com/nathakusuma/auditorium-reservation-backend/domain/errorpkg"
	"github.com/nathakusuma/auditorium-reservation-backend/pkg/log"
	"github.com/nathakusuma/auditorium-reservation-backend/pkg/uuidpkg"
)

type feedbackService struct {
	repo            contract.IFeedbackRepository
	registrationSvc contract.IRegistrationService
	conferenceSvc   contract.IConferenceService
	uuid            uuidpkg.IUUID
}

func NewFeedbackService(
	feedbackRepository contract.IFeedbackRepository,
	registrationService contract.IRegistrationService,
	conferenceService contract.IConferenceService,
	uuid uuidpkg.IUUID,
) contract.IFeedbackService {
	return &feedbackService{
		repo:            feedbackRepository,
		registrationSvc: registrationService,
		conferenceSvc:   conferenceService,
		uuid:            uuid,
	}
}

func (s *feedbackService) CreateFeedback(ctx context.Context, userID, conferenceID uuid.UUID,
	comment string) (uuid.UUID, error) {

	isRegistered, err := s.registrationSvc.IsUserRegisteredToConference(ctx, conferenceID, userID)
	if err != nil {
		return uuid.Nil, err
	}
	if !isRegistered {
		return uuid.Nil, errorpkg.ErrUserNotRegisteredToConference
	}

	hasGivenFeedback, err := s.repo.IsFeedbackGiven(ctx, userID, conferenceID)
	if err != nil {
		traceID := log.ErrorWithTraceID(map[string]interface{}{
			"error":        err,
			"userID":       userID,
			"conferenceID": conferenceID,
		}, "[FeedbackService][CreateFeedback] Failed to check if feedback is given")
		return uuid.Nil, errorpkg.ErrInternalServer.WithTraceID(traceID)
	}
	if hasGivenFeedback {
		return uuid.Nil, errorpkg.ErrFeedbackAlreadyGiven
	}

	conference, err := s.conferenceSvc.GetConferenceByID(ctx, conferenceID)
	if err != nil {
		return uuid.Nil, err
	}

	if conference.Host.ID == userID {
		return uuid.Nil, errorpkg.ErrHostCannotGiveFeedback
	}

	if conference.EndsAt.After(time.Now()) {
		return uuid.Nil, errorpkg.ErrConferenceNotEnded
	}

	feedbackID, err := s.uuid.NewV7()
	if err != nil {
		traceID := log.ErrorWithTraceID(map[string]interface{}{
			"error":        err,
			"userID":       userID,
			"conferenceID": conferenceID,
		}, "[FeedbackService][CreateFeedback] Failed to generate UUID")
		return uuid.Nil, errorpkg.ErrInternalServer.WithTraceID(traceID)
	}

	feedback := &entity.Feedback{
		ID:           feedbackID,
		UserID:       userID,
		ConferenceID: conferenceID,
		Comment:      comment,
	}

	if err := s.repo.CreateFeedback(ctx, feedback); err != nil {
		traceID := log.ErrorWithTraceID(map[string]interface{}{
			"error":    err,
			"feedback": feedback,
		}, "[FeedbackService][CreateFeedback] Failed to create feedback")
		return uuid.Nil, errorpkg.ErrInternalServer.WithTraceID(traceID)
	}

	log.Info(map[string]interface{}{
		"feedback": feedback,
	}, "[FeedbackService][CreateFeedback] Feedback created successfully")

	return feedbackID, nil
}

func (s *feedbackService) GetFeedbacksByConferenceID(ctx context.Context, conferenceID uuid.UUID,
	lazyReq dto.LazyLoadQuery) ([]dto.FeedbackResponse, dto.LazyLoadResponse, error) {

	feedbacks, lazyResp, err := s.repo.GetFeedbacksByConferenceID(ctx, conferenceID, lazyReq)
	if err != nil {
		traceID := log.ErrorWithTraceID(map[string]interface{}{
			"error":        err,
			"conferenceID": conferenceID,
		}, "[FeedbackService][GetFeedbacksByConferenceID] Failed to get feedbacks by conference ID")
		return nil, dto.LazyLoadResponse{}, errorpkg.ErrInternalServer.WithTraceID(traceID)
	}

	resp := make([]dto.FeedbackResponse, len(feedbacks))
	for i, feedback := range feedbacks {
		resp[i].PopulateFromEntity(&feedback)
	}

	return resp, lazyResp, nil
}

func (s *feedbackService) DeleteFeedback(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.DeleteFeedback(ctx, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errorpkg.ErrNotFound
		}

		traceID := log.ErrorWithTraceID(map[string]interface{}{
			"error":        err,
			"feedback.id":  id,
			"requester.id": ctx.Value("user.id"),
		}, "[FeedbackService][DeleteFeedback] Failed to delete feedback")
		return errorpkg.ErrInternalServer.WithTraceID(traceID)
	}

	log.Info(map[string]interface{}{
		"feedback.id":  id,
		"requester.id": ctx.Value("user.id"),
	}, "[FeedbackService][DeleteFeedback] Feedback deleted successfully")

	return nil
}
