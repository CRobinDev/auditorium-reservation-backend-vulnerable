package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/nathakusuma/auditorium-reservation-backend/domain/contract"
	"github.com/nathakusuma/auditorium-reservation-backend/domain/dto"
	"github.com/nathakusuma/auditorium-reservation-backend/domain/errorpkg"
	"github.com/nathakusuma/auditorium-reservation-backend/internal/middleware"
	"github.com/nathakusuma/auditorium-reservation-backend/pkg/validator"
)

type feedbackHandler struct {
	svc contract.IFeedbackService
	val validator.IValidator
}

func InitFeedbackHandler(
	router fiber.Router,
	midw *middleware.Middleware,
	val validator.IValidator,
	feedbackSvc contract.IFeedbackService,

) {
	handler := feedbackHandler{
		svc: feedbackSvc,
		val: val,
	}

	feedbackGroup := router.Group("/feedbacks")
	feedbackGroup.Use(midw.RequireAuthenticated())

	feedbackGroup.Post("",
		handler.createFeedback(),
	)

	feedbackGroup.Get("/conferences/:id",
		handler.getFeedbacksByConferenceID(),
	)

	feedbackGroup.Delete("/:id",
		handler.deleteFeedback(),
	)
}

func (h *feedbackHandler) createFeedback() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		type request struct {
			ConferenceID uuid.UUID `json:"conference_id" validate:"required,uuid"`
			Comment      string    `json:"comment" validate:"required,min=3,max=1000"`
		}

		var req request
		if err := ctx.BodyParser(&req); err != nil {
			return errorpkg.ErrFailParseRequest
		}

		if err := h.val.ValidateStruct(req); err != nil {
			return err
		}

		userID, _ := ctx.Locals("user.id").(uuid.UUID)

		feedbackID, err := h.svc.CreateFeedback(ctx.Context(), userID, req.ConferenceID, req.Comment)
		if err != nil {
			return err
		}

		return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{
			"feedback": dto.FeedbackResponse{ID: feedbackID},
		})
	}
}

func (h *feedbackHandler) getFeedbacksByConferenceID() fiber.Handler {
	return func(c *fiber.Ctx) error {
		conferenceID, err := uuid.Parse(c.Params("id"))
		if err != nil {
			return errorpkg.ErrFailParseRequest
		}

		var lazyReq dto.LazyLoadQuery
		if err := c.QueryParser(&lazyReq); err != nil {
			return errorpkg.ErrFailParseRequest
		}

		if err := h.val.ValidateStruct(lazyReq); err != nil {
			return err
		}

		feedbacks, lazyResp, err := h.svc.GetFeedbacksByConferenceID(c.Context(), conferenceID, lazyReq)
		if err != nil {
			return err
		}

		return c.JSON(map[string]interface{}{
			"feedbacks":  feedbacks,
			"pagination": lazyResp,
		})
	}
}

func (h *feedbackHandler) deleteFeedback() fiber.Handler {
	return func(c *fiber.Ctx) error {
		feedbackID, err := uuid.Parse(c.Params("id"))
		if err != nil {
			return errorpkg.ErrFailParseRequest
		}

		if err := h.svc.DeleteFeedback(c.Context(), feedbackID); err != nil {
			return err
		}

		return c.SendStatus(fiber.StatusNoContent)
	}
}
