package handler

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/nathakusuma/auditorium-reservation-backend/domain/contract"
	"github.com/nathakusuma/auditorium-reservation-backend/domain/dto"
	"github.com/nathakusuma/auditorium-reservation-backend/domain/enum"
	"github.com/nathakusuma/auditorium-reservation-backend/domain/errorpkg"
	"github.com/nathakusuma/auditorium-reservation-backend/internal/middleware"
	"github.com/nathakusuma/auditorium-reservation-backend/pkg/validator"
)

type conferenceHandler struct {
	val validator.IValidator
	svc contract.IConferenceService
}

func InitConferenceHandler(
	router fiber.Router,
	midw *middleware.Middleware,
	validator validator.IValidator,
	conferenceSvc contract.IConferenceService,
) {
	handler := conferenceHandler{
		svc: conferenceSvc,
		val: validator,
	}

	conferenceGroup := router.Group("/conferences")
	conferenceGroup.Use(midw.RequireAuthenticated())

	conferenceGroup.Post("",
		midw.RequireOneOfRoles(enum.RoleUser),
		handler.createConferenceProposal(),
	)
	conferenceGroup.Get("/:id",
		handler.getConferenceByID(),
	)
	conferenceGroup.Get("",
		handler.getConferences(),
	)
	conferenceGroup.Patch("/:id",
		midw.RequireOneOfRoles(enum.RoleUser),
		handler.updateConference(),
	)
	conferenceGroup.Delete("/:id",
		midw.RequireOneOfRoles(enum.RoleUser, enum.RoleEventCoordinator),
		handler.deleteConference(),
	)
	conferenceGroup.Patch("/:id/status",
		midw.RequireOneOfRoles(enum.RoleEventCoordinator),
		handler.updateConferenceStatus(),
	)
}

func (c *conferenceHandler) createConferenceProposal() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		type request struct {
			Title          string  `json:"title" validate:"required,min=3,max=100"`
			Description    string  `json:"description" validate:"required,min=3,max=1000"`
			SpeakerName    string  `json:"speaker_name" validate:"required,min=3,max=100"`
			SpeakerTitle   string  `json:"speaker_title" validate:"required,min=3,max=100"`
			TargetAudience string  `json:"target_audience" validate:"required,min=3,max=255"`
			Prerequisites  *string `json:"prerequisites" validate:"omitempty,max=255"`
			Seats          int     `json:"seats" validate:"required,min=1"`
			StartsAt       string  `json:"starts_at" validate:"required,datetime=2006-01-02T15:04:05Z07:00"`
			EndsAt         string  `json:"ends_at" validate:"required,datetime=2006-01-02T15:04:05Z07:00"`
		}

		var req request
		if err := ctx.BodyParser(&req); err != nil {
			return errorpkg.ErrFailParseRequest
		}

		if err := c.val.ValidateStruct(req); err != nil {
			return err
		}

		startsAt, err := time.Parse(time.RFC3339, req.StartsAt)
		endsAt, err2 := time.Parse(time.RFC3339, req.EndsAt)
		if err != nil || err2 != nil {
			return errorpkg.ErrFailParseRequest
		}

		proposal := dto.CreateConferenceProposalRequest{
			Title:          req.Title,
			Description:    req.Description,
			SpeakerName:    req.SpeakerName,
			SpeakerTitle:   req.SpeakerTitle,
			TargetAudience: req.TargetAudience,
			Prerequisites:  req.Prerequisites,
			Seats:          req.Seats,
			StartsAt:       startsAt,
			EndsAt:         endsAt,
		}

		conferenceID, err := c.svc.CreateConferenceProposal(ctx.Context(), &proposal)
		if err != nil {
			return err
		}

		return ctx.Status(fiber.StatusCreated).JSON(map[string]interface{}{
			"conference": dto.ConferenceResponse{ID: conferenceID},
		})
	}
}

func (c *conferenceHandler) getConferenceByID() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		conferenceID, err := uuid.Parse(ctx.Params("id"))
		if err != nil {
			return errorpkg.ErrFailParseRequest
		}

		conference, err := c.svc.GetConferenceByID(ctx.Context(), conferenceID)
		if err != nil {
			return err
		}

		return ctx.JSON(map[string]interface{}{
			"conference": conference,
		})
	}
}

func (c *conferenceHandler) getConferences() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		type request struct {
			AfterID      *uuid.UUID            `query:"after_id" validate:"omitempty,uuid"`
			BeforeID     *uuid.UUID            `query:"before_id" validate:"omitempty,uuid"`
			Limit        int                   `query:"limit" validate:"required,min=1,max=20"`
			HostID       *uuid.UUID            `query:"host_id" validate:"omitempty,uuid"`
			Status       enum.ConferenceStatus `query:"status" validate:"required,oneof=pending approved rejected"`
			StartsBefore *string               `query:"starts_before" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
			StartsAfter  *string               `query:"starts_after" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
			IncludePast  bool                  `query:"include_past" validate:"omitempty"`
			OrderBy      string                `query:"order_by" validate:"required,oneof=created_at starts_at"`
			Order        string                `query:"order" validate:"required,oneof=asc desc"`
			Title        *string               `query:"title" validate:"omitempty"`
		}

		var req request
		if err := ctx.QueryParser(&req); err != nil {
			return errorpkg.ErrFailParseRequest
		}

		if err := c.val.ValidateStruct(req); err != nil {
			return err
		}

		var startsBefore, startsAfter *time.Time
		if req.StartsBefore != nil {
			startsBeforeValue, err := time.Parse(time.RFC3339, *req.StartsBefore)
			if err != nil {
				return errorpkg.ErrFailParseRequest
			}
			startsBefore = &startsBeforeValue
		}

		if req.StartsAfter != nil {
			startsAfterValue, err := time.Parse(time.RFC3339, *req.StartsAfter)
			if err != nil {
				return errorpkg.ErrFailParseRequest
			}
			startsAfter = &startsAfterValue
		}

		query := dto.GetConferenceQuery{
			AfterID:      req.AfterID,
			BeforeID:     req.BeforeID,
			Limit:        req.Limit,
			HostID:       req.HostID,
			Status:       req.Status,
			StartsBefore: startsBefore,
			StartsAfter:  startsAfter,
			IncludePast:  req.IncludePast,
			OrderBy:      req.OrderBy,
			Order:        req.Order,
			Title:        req.Title,
		}

		conferences, lazy, err := c.svc.GetConferences(ctx.Context(), &query)
		if err != nil {
			return err
		}

		return ctx.JSON(map[string]interface{}{
			"conferences": conferences,
			"pagination":  lazy,
		})
	}
}

func (c *conferenceHandler) updateConference() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		type request struct {
			Title          *string `json:"title" validate:"omitempty,min=3,max=100"`
			Description    *string `json:"description" validate:"omitempty,min=3,max=1000"`
			SpeakerName    *string `json:"speaker_name" validate:"omitempty,min=3,max=100"`
			SpeakerTitle   *string `json:"speaker_title" validate:"omitempty,min=3,max=100"`
			TargetAudience *string `json:"target_audience" validate:"omitempty,min=3,max=255"`
			Prerequisites  *string `json:"prerequisites" validate:"omitempty,max=255"`
			StartsAt       *string `json:"starts_at" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
			EndsAt         *string `json:"ends_at" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
		}

		conferenceID, err := uuid.Parse(ctx.Params("id"))
		if err != nil {
			return errorpkg.ErrFailParseRequest
		}

		var req request
		if err2 := ctx.BodyParser(&req); err2 != nil {
			return errorpkg.ErrFailParseRequest
		}

		if err2 := c.val.ValidateStruct(req); err2 != nil {
			return err
		}

		var startsAt, endsAt *time.Time
		if req.StartsAt != nil {
			startsAtValue, err2 := time.Parse(time.RFC3339, *req.StartsAt)
			if err2 != nil {
				return errorpkg.ErrFailParseRequest
			}
			startsAt = &startsAtValue
		}

		if req.EndsAt != nil {
			endsAtValue, err2 := time.Parse(time.RFC3339, *req.EndsAt)
			if err2 != nil {
				return errorpkg.ErrFailParseRequest
			}
			endsAt = &endsAtValue
		}

		conference := dto.UpdateConferenceRequest{
			Title:          req.Title,
			Description:    req.Description,
			SpeakerName:    req.SpeakerName,
			SpeakerTitle:   req.SpeakerTitle,
			TargetAudience: req.TargetAudience,
			Prerequisites:  req.Prerequisites,
			StartsAt:       startsAt,
			EndsAt:         endsAt,
		}

		if err = c.svc.UpdateConference(ctx.Context(), conferenceID, conference); err != nil {
			return err
		}

		return ctx.SendStatus(fiber.StatusNoContent)
	}
}

func (c *conferenceHandler) deleteConference() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		conferenceID, err := uuid.Parse(ctx.Params("id"))
		if err != nil {
			return errorpkg.ErrFailParseRequest
		}

		if err = c.svc.DeleteConference(ctx.Context(), conferenceID); err != nil {
			return err
		}

		return ctx.SendStatus(fiber.StatusNoContent)
	}
}

func (c *conferenceHandler) updateConferenceStatus() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		type request struct {
			Status enum.ConferenceStatus `json:"status" validate:"required,oneof=pending approved rejected"`
		}

		conferenceID, err := uuid.Parse(ctx.Params("id"))
		if err != nil {
			return errorpkg.ErrFailParseRequest
		}

		var req request
		if err = ctx.BodyParser(&req); err != nil {
			return errorpkg.ErrFailParseRequest
		}

		if err = c.val.ValidateStruct(req); err != nil {
			return err
		}

		if err = c.svc.UpdateConferenceStatus(ctx.Context(), conferenceID, req.Status); err != nil {
			return err
		}

		return ctx.SendStatus(fiber.StatusNoContent)
	}
}
