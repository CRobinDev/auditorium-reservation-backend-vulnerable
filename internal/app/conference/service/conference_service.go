package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nathakusuma/auditorium-reservation-backend/domain/contract"
	"github.com/nathakusuma/auditorium-reservation-backend/domain/dto"
	"github.com/nathakusuma/auditorium-reservation-backend/domain/entity"
	"github.com/nathakusuma/auditorium-reservation-backend/domain/enum"
	"github.com/nathakusuma/auditorium-reservation-backend/domain/errorpkg"
	"github.com/nathakusuma/auditorium-reservation-backend/pkg/log"
	"github.com/nathakusuma/auditorium-reservation-backend/pkg/uuidpkg"
)

type conferenceService struct {
	r    contract.IConferenceRepository
	uuid uuidpkg.IUUID
}

func NewConferenceService(conferenceRepo contract.IConferenceRepository,
	uuid uuidpkg.IUUID) contract.IConferenceService {

	return &conferenceService{r: conferenceRepo, uuid: uuid}
}

func (s *conferenceService) CreateConferenceProposal(ctx context.Context,
	req *dto.CreateConferenceProposalRequest) (uuid.UUID, error) {
	// Create Conference Session Proposal

	requesterID, ok := ctx.Value("user.id").(uuid.UUID)
	if !ok {
		traceID := log.ErrorWithTraceID(map[string]interface{}{
			"error":        errors.New("failed to get user id from context"),
			"requester.id": requesterID,
		}, "[ConferenceService][CreateConferenceProposal] Failed to get user id from context")
		return uuid.Nil, errorpkg.ErrInternalServer.WithTraceID(traceID)
	}

	// Check if user has active proposal
	userConferences, _, err := s.GetConferences(ctx, &dto.GetConferenceQuery{
		Limit:       1,
		HostID:      &requesterID,
		Status:      enum.ConferencePending,
		IncludePast: false,
		OrderBy:     "created_at",
		Order:       "desc",
	})
	if err != nil {
		return uuid.Nil, err
	}

	if len(userConferences) > 0 {
		conflict := userConferences[0]
		return uuid.Nil, errorpkg.ErrUserHasActiveProposal.WithDetail(map[string]interface{}{
			"conference": dto.ConferenceResponse{
				ID:        conflict.ID,
				Title:     conflict.Title,
				Status:    conflict.Status,
				CreatedAt: conflict.CreatedAt,
			}})
	}

	// Check if there is a conference in the same time window
	conflicts, err := s.r.GetConferencesConflictingWithTime(ctx, req.StartsAt, req.EndsAt, uuid.Nil)
	if err != nil {
		traceID := log.ErrorWithTraceID(map[string]interface{}{
			"error":        err,
			"request":      req,
			"requester.id": requesterID,
		}, "[ConferenceService][CreateConferenceProposal] Failed to get conflicting conferences")
		return uuid.Nil, errorpkg.ErrInternalServer.WithTraceID(traceID)
	}

	if len(conflicts) > 0 {
		resp := make([]dto.ConferenceResponse, len(conflicts))
		for i, conflict := range conflicts {
			resp[i] = dto.ConferenceResponse{
				ID:       conflict.ID,
				Title:    conflict.Description,
				StartsAt: &conflict.StartsAt,
				EndsAt:   &conflict.EndsAt,
			}
		}

		return uuid.Nil, errorpkg.ErrTimeWindowConflict.WithDetail(map[string]interface{}{
			"conferences": resp,
		})
	}

	// Create conference
	conferenceID, err := s.uuid.NewV7()
	if err != nil {
		traceID := log.ErrorWithTraceID(map[string]interface{}{
			"error":        err,
			"request":      req,
			"requester.id": requesterID,
		}, "[ConferenceService][CreateConferenceProposal] Failed to generate conference ID")
		return uuid.Nil, errorpkg.ErrInternalServer.WithTraceID(traceID)
	}

	if req.StartsAt.Before(time.Now()) {
		return uuid.Nil, errorpkg.ErrTimeAlreadyPassed
	}

	if req.EndsAt.Before(req.StartsAt) {
		return uuid.Nil, errorpkg.ErrEndTimeBeforeStart
	}

	conference := entity.Conference{
		ID:             conferenceID,
		Title:          req.Title,
		Description:    req.Description,
		SpeakerName:    req.SpeakerName,
		SpeakerTitle:   req.SpeakerTitle,
		TargetAudience: req.TargetAudience,
		Prerequisites:  req.Prerequisites,
		Seats:          req.Seats,
		StartsAt:       req.StartsAt,
		EndsAt:         req.EndsAt,
		HostID:         requesterID,
		Status:         enum.ConferencePending,
	}

	if err = s.r.CreateConference(ctx, &conference); err != nil {
		traceID := log.ErrorWithTraceID(map[string]interface{}{
			"error":        err,
			"request":      req,
			"requester.id": requesterID,
		}, "[ConferenceService][CreateConferenceProposal] Failed to create conference")
		return uuid.Nil, errorpkg.ErrInternalServer.WithTraceID(traceID)
	}

	log.Info(map[string]interface{}{
		"conference":   conference,
		"requester.id": requesterID,
	}, "[ConferenceService][CreateConferenceProposal] Conference proposal created")

	return conferenceID, nil
}

func (s *conferenceService) GetConferenceByID(ctx context.Context, id uuid.UUID) (*dto.ConferenceResponse, error) {
	requesterID, _ := ctx.Value("user.id").(uuid.UUID)
	requesterRole, _ := ctx.Value("user.role").(enum.UserRole)

	conference, err := s.r.GetConferenceByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errorpkg.ErrNotFound
		}

		traceID := log.ErrorWithTraceID(map[string]interface{}{
			"error":        err.Error(),
			"requester.id": requesterID,
		}, "[ConferenceService][GetConferenceByID] Failed to get conference")
		return nil, errorpkg.ErrInternalServer.WithTraceID(traceID)
	}

	isRestrictedUser := requesterRole == enum.RoleUser && conference.HostID != requesterID

	if conference.Status != enum.ConferenceApproved && isRestrictedUser {
		return nil, errorpkg.ErrForbiddenUser
	}

	var resp dto.ConferenceResponse
	resp.PopulateFromEntity(conference)

	return &resp, nil
}

func (s *conferenceService) GetConferences(ctx context.Context,
	query *dto.GetConferenceQuery) ([]dto.ConferenceResponse, dto.LazyLoadResponse, error) {

	if query.AfterID != nil && query.BeforeID != nil {
		return nil, dto.LazyLoadResponse{}, errorpkg.ErrInvalidPagination
	}

	requesterID, _ := ctx.Value("user.id").(uuid.UUID)
	requesterRole, _ := ctx.Value("user.role").(enum.UserRole)

	// If requester is system, it will not enter this block because requesterRole is empty
	if query.Status != enum.ConferenceApproved && requesterRole == enum.RoleUser {
		if query.HostID == nil {
			query.HostID = &requesterID
		} else if *query.HostID != requesterID {
			return nil, dto.LazyLoadResponse{}, errorpkg.ErrForbiddenUser
		}
	}

	conferences, lazy, err := s.r.GetConferences(ctx, query)
	if err != nil {
		traceID := log.ErrorWithTraceID(map[string]interface{}{
			"error":        err.Error(),
			"requester.id": requesterID,
		}, "[ConferenceService][GetConferences] Failed to get conferences")
		return nil, dto.LazyLoadResponse{}, errorpkg.ErrInternalServer.WithTraceID(traceID)
	}

	resp := make([]dto.ConferenceResponse, len(conferences))
	for i, conference := range conferences {
		resp[i].PopulateFromEntity(&conference)
	}

	return resp, lazy, nil
}

func (s *conferenceService) UpdateConference(ctx context.Context, id uuid.UUID, req dto.UpdateConferenceRequest) error {
	requesterID, _ := ctx.Value("user.id").(uuid.UUID)
	requesterRole, _ := ctx.Value("user.role").(enum.UserRole)

	original, err := s.r.GetConferenceByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errorpkg.ErrNotFound
		}

		traceID := log.ErrorWithTraceID(map[string]interface{}{
			"error":        err,
			"requester.id": requesterID,
		}, "[ConferenceService][UpdateConference] Failed to get conference")
		return errorpkg.ErrInternalServer.WithTraceID(traceID)
	}

	conference := *original
	req.GenerateUpdateEntity(&conference)

	// Check if user is the host
	if requesterRole == enum.RoleUser && conference.HostID != requesterID {
		return errorpkg.ErrForbiddenUser
	}

	if original.EndsAt.Before(time.Now()) {
		return errorpkg.ErrUpdatePastConference
	}

	if original.Status != enum.ConferencePending {
		return errorpkg.ErrUpdateNotPendingConference
	}

	// Check if there is a conference in the same time window
	if req.StartsAt != nil || req.EndsAt != nil {
		if conference.StartsAt.Before(time.Now()) {
			return errorpkg.ErrTimeAlreadyPassed
		}

		if conference.EndsAt.Before(conference.StartsAt) {
			return errorpkg.ErrEndTimeBeforeStart
		}

		// Check if there is a conference in the same time window
		conflicts, err := s.r.GetConferencesConflictingWithTime(ctx, *req.StartsAt, *req.EndsAt, id)
		if err != nil {
			traceID := log.ErrorWithTraceID(map[string]interface{}{
				"error":        err,
				"requester.id": requesterID,
			}, "[ConferenceService][UpdateConference] Failed to get conflicting conferences")
			return errorpkg.ErrInternalServer.WithTraceID(traceID)
		}

		if len(conflicts) > 0 {
			resp := make([]dto.ConferenceResponse, len(conflicts))
			for i, conflict := range conflicts {
				resp[i] = dto.ConferenceResponse{
					ID:       conflict.ID,
					Title:    conflict.Description,
					StartsAt: &conflict.StartsAt,
					EndsAt:   &conflict.EndsAt,
				}
			}

			return errorpkg.ErrTimeWindowConflict.WithDetail(map[string]interface{}{
				"conferences": resp,
			})
		}
	}

	// update conference
	if err = s.r.UpdateConference(ctx, &conference); err != nil {
		traceID := log.ErrorWithTraceID(map[string]interface{}{
			"error":        err,
			"requester.id": requesterID,
		}, "[ConferenceService][UpdateConference] Failed to update conference")
		return errorpkg.ErrInternalServer.WithTraceID(traceID)
	}

	log.Info(map[string]interface{}{
		"conference":   conference,
		"requester.id": requesterID,
	}, "[ConferenceService][UpdateConference] Conference updated")

	return nil
}

func (s *conferenceService) DeleteConference(ctx context.Context, id uuid.UUID) error {
	requesterID, _ := ctx.Value("user.id").(uuid.UUID)
	requesterRole, _ := ctx.Value("user.role").(enum.UserRole)

	conference, err := s.r.GetConferenceByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errorpkg.ErrNotFound
		}

		traceID := log.ErrorWithTraceID(map[string]interface{}{
			"error":        err,
			"requester.id": requesterID,
		}, "[ConferenceService][DeleteConference] Failed to get conference")
		return errorpkg.ErrInternalServer.WithTraceID(traceID)
	}

	if requesterRole == enum.RoleUser && conference.HostID != requesterID {
		return errorpkg.ErrForbiddenUser
	}

	if err = s.r.DeleteConference(ctx, id); err != nil {
		traceID := log.ErrorWithTraceID(map[string]interface{}{
			"error":        err,
			"requester.id": ctx.Value("user.id"),
		}, "[ConferenceService][DeleteConference] Failed to delete conference")
		return errorpkg.ErrInternalServer.WithTraceID(traceID)
	}

	log.Info(map[string]interface{}{
		"conference":   conference,
		"requester.id": ctx.Value("user.id"),
	}, "[ConferenceService][DeleteConference] Conference deleted")

	return nil
}

func (s *conferenceService) UpdateConferenceStatus(ctx context.Context, id uuid.UUID,
	status enum.ConferenceStatus) error {

	conference, err := s.r.GetConferenceByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errorpkg.ErrNotFound
		}

		traceID := log.ErrorWithTraceID(map[string]interface{}{
			"error":        err,
			"requester.id": ctx.Value("user.id"),
		}, "[ConferenceService][UpdateConferenceStatus] Failed to get conference")
		return errorpkg.ErrInternalServer.WithTraceID(traceID)
	}

	if conference.Status != enum.ConferencePending {
		return errorpkg.ErrUpdateNotPendingConference
	}

	// Only allow transition from pending to rejected if conference is in the past
	if conference.StartsAt.Before(time.Now()) && status != enum.ConferenceRejected {
		return errorpkg.ErrUpdatePastConferenceStatus
	}

	if status == enum.ConferenceApproved {
		// Check for time conflicts only when approving
		conflicts, err2 := s.r.GetConferencesConflictingWithTime(ctx, conference.StartsAt, conference.EndsAt, id)
		if err2 != nil {
			traceID := log.ErrorWithTraceID(map[string]interface{}{
				"error":        err2,
				"requester.id": ctx.Value("user.id"),
			}, "[ConferenceService][UpdateConferenceStatus] Failed to get conflicting conferences")
			return errorpkg.ErrInternalServer.WithTraceID(traceID)
		}

		if len(conflicts) > 0 {
			resp := make([]dto.ConferenceResponse, len(conflicts))
			for i, conflict := range conflicts {
				resp[i] = dto.ConferenceResponse{
					ID:       conflict.ID,
					Title:    conflict.Description,
					StartsAt: &conflict.StartsAt,
					EndsAt:   &conflict.EndsAt,
				}
			}

			return errorpkg.ErrTimeWindowConflict.WithDetail(map[string]interface{}{
				"conferences": resp,
			})
		}
	}

	// update conference status
	conference.Status = status

	if err = s.r.UpdateConference(ctx, conference); err != nil {
		traceID := log.ErrorWithTraceID(map[string]interface{}{
			"error":        err,
			"requester.id": ctx.Value("user.id"),
		}, fmt.Sprintf("[ConferenceService][UpdateConferenceStatus] Failed to update conference status to %s", status))
		return errorpkg.ErrInternalServer.WithTraceID(traceID)
	}

	log.Info(map[string]interface{}{
		"conference":   conference,
		"requester.id": ctx.Value("user.id"),
	}, fmt.Sprintf("[ConferenceService][UpdateConferenceStatus] Conference status updated to %s", status))

	return nil
}
