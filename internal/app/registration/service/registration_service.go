package service

import (
	"context"
	"github.com/google/uuid"
	"github.com/nathakusuma/auditorium-reservation-backend/domain/contract"
	"github.com/nathakusuma/auditorium-reservation-backend/domain/dto"
	"github.com/nathakusuma/auditorium-reservation-backend/domain/entity"
	"github.com/nathakusuma/auditorium-reservation-backend/domain/enum"
	"github.com/nathakusuma/auditorium-reservation-backend/domain/errorpkg"
	"github.com/nathakusuma/auditorium-reservation-backend/pkg/log"
	"time"
)

type registrationService struct {
	r             contract.IRegistrationRepository
	conferenceSvc contract.IConferenceService
}

func NewRegistrationService(registrationRepository contract.IRegistrationRepository,
	conferenceService contract.IConferenceService) contract.IRegistrationService {

	return &registrationService{
		r:             registrationRepository,
		conferenceSvc: conferenceService,
	}
}

func (s *registrationService) Register(ctx context.Context, conferenceID, userID uuid.UUID) error {
	// Get conference by ID
	conference, err := s.conferenceSvc.GetConferenceByID(ctx, conferenceID)
	if err != nil {
		return err
	}
	// At this point, conference must have been approved, because it's checked in the GetConferenceByID method

	// Is user host of conference?
	if conference.Host.ID == userID {
		return errorpkg.ErrHostCannotRegister
	}

	// Is the conference already ended?
	if conference.EndsAt.Before(time.Now()) {
		return errorpkg.ErrConferenceEnded
	}

	// Is user registered to conference?
	isRegistered, err := s.r.IsUserRegisteredToConference(ctx, conferenceID, userID)
	if err != nil {
		traceID := log.ErrorWithTraceID(map[string]interface{}{
			"error":        err,
			"conferenceID": conferenceID,
			"userID":       userID,
		}, "[RegistrationService][Register] Failed to check if user is registered to conference")
		return errorpkg.ErrInternalServer.WithTraceID(traceID)
	}
	if isRegistered {
		return errorpkg.ErrUserAlreadyRegisteredToConference
	}

	// Is it full?
	taken, err := s.r.CountRegistrationsByConference(ctx, conferenceID)
	if err != nil {
		traceID := log.ErrorWithTraceID(map[string]interface{}{
			"error":        err,
			"conferenceID": conferenceID,
		}, "[RegistrationService][Register] Failed to count registrations by conference")
		return errorpkg.ErrInternalServer.WithTraceID(traceID)
	}
	if taken >= conference.Seats {
		return errorpkg.ErrConferenceFull
	}

	// Get conflicting registrations
	// This is unlikely to happen. There will never be conflicting approved conferences (checked in conference service)
	// Just to be safe
	conflictingRegistrations, err := s.r.GetConflictingRegistrations(ctx, userID, *conference.StartsAt,
		*conference.EndsAt)
	if err != nil {
		traceID := log.ErrorWithTraceID(map[string]interface{}{
			"error":        err,
			"conferenceID": conferenceID,
			"userID":       userID,
		}, "[RegistrationService][Register] Failed to get conflicting registrations")
		return errorpkg.ErrInternalServer.WithTraceID(traceID)
	}
	if len(conflictingRegistrations) > 0 {
		return errorpkg.ErrConflictingRegistrations.WithDetail(map[string]interface{}{
			"conferences": conflictingRegistrations,
		})
	}

	// Create registration
	if err := s.r.CreateRegistration(ctx, &entity.Registration{
		ConferenceID: conferenceID,
		UserID:       userID,
	}); err != nil {
		traceID := log.ErrorWithTraceID(map[string]interface{}{
			"error":        err,
			"conferenceID": conferenceID,
			"userID":       userID,
		}, "[RegistrationService][Register] Failed to create registration")
		return errorpkg.ErrInternalServer.WithTraceID(traceID)
	}

	return nil
}

func (s *registrationService) GetRegisteredUsersByConference(ctx context.Context,
	conferenceID uuid.UUID, lazyReq dto.LazyLoadQuery) ([]dto.UserResponse, dto.LazyLoadResponse, error) {
	if lazyReq.AfterID != uuid.Nil && lazyReq.BeforeID != uuid.Nil {
		return nil, dto.LazyLoadResponse{}, errorpkg.ErrInvalidPagination
	}

	requesterID, _ := ctx.Value("user.id").(uuid.UUID)
	requesterRole, _ := ctx.Value("user.role").(enum.UserRole)

	conference, err := s.conferenceSvc.GetConferenceByID(ctx, conferenceID)
	if err != nil {
		return nil, dto.LazyLoadResponse{}, err
	}

	if requesterRole == enum.RoleUser && requesterID != conference.Host.ID {
		return nil, dto.LazyLoadResponse{}, errorpkg.ErrForbiddenUser
	}

	users, lazyResp, err := s.r.GetRegisteredUsersByConference(ctx, conferenceID, lazyReq)
	if err != nil {
		traceID := log.ErrorWithTraceID(map[string]interface{}{
			"error":        err,
			"conferenceID": conferenceID,
			"requester.id": requesterID,
		}, "[RegistrationService][GetRegisteredUsersByConference] Failed to get registered users by conference")
		return nil, dto.LazyLoadResponse{}, errorpkg.ErrInternalServer.WithTraceID(traceID)
	}

	resp := make([]dto.UserResponse, len(users))
	for i, user := range users {
		resp[i] = dto.UserResponse{
			ID:    user.ID,
			Name:  user.Name,
			Email: user.Email,
		}
	}

	return resp, lazyResp, nil
}

func (s *registrationService) GetRegisteredConferencesByUser(ctx context.Context, userID uuid.UUID,
	includePast bool, lazyReq dto.LazyLoadQuery) ([]dto.ConferenceResponse, dto.LazyLoadResponse, error) {

	if lazyReq.AfterID != uuid.Nil && lazyReq.BeforeID != uuid.Nil {
		return nil, dto.LazyLoadResponse{}, errorpkg.ErrInvalidPagination
	}

	requesterID, _ := ctx.Value("user.id").(uuid.UUID)
	requesterRole, _ := ctx.Value("user.role").(enum.UserRole)

	if requesterRole == enum.RoleUser && requesterID != userID {
		return nil, dto.LazyLoadResponse{}, errorpkg.ErrForbiddenUser
	}

	conferences, lazyResp, err := s.r.GetRegisteredConferencesByUser(ctx, userID, includePast, lazyReq)
	if err != nil {
		traceID := log.ErrorWithTraceID(map[string]interface{}{
			"error":        err,
			"userID":       userID,
			"requester.id": requesterID,
		}, "[RegistrationService][GetRegisteredConferencesByUser] Failed to get registered conferences by user")
		return nil, dto.LazyLoadResponse{}, errorpkg.ErrInternalServer.WithTraceID(traceID)
	}

	resp := make([]dto.ConferenceResponse, len(conferences))
	for i, conference := range conferences {
		var temp dto.ConferenceResponse
		temp.PopulateFromEntity(&conference)
		resp[i] = temp
	}

	return resp, lazyResp, nil
}

func (s *registrationService) IsUserRegisteredToConference(ctx context.Context, conferenceID,
	userID uuid.UUID) (bool, error) {

	ok, err := s.r.IsUserRegisteredToConference(ctx, conferenceID, userID)
	if err != nil {
		traceID := log.ErrorWithTraceID(map[string]interface{}{
			"error":        err,
			"conferenceID": conferenceID,
			"userID":       userID,
		}, "[RegistrationService][IsUserRegisteredToConference] Failed to check if user is registered to conference")
		return false, errorpkg.ErrInternalServer.WithTraceID(traceID)
	}

	return ok, nil
}
