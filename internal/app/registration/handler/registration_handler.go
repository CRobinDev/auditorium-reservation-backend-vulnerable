package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/nathakusuma/auditorium-reservation-backend/domain/contract"
	"github.com/nathakusuma/auditorium-reservation-backend/domain/dto"
	"github.com/nathakusuma/auditorium-reservation-backend/domain/enum"
	"github.com/nathakusuma/auditorium-reservation-backend/domain/errorpkg"
	"github.com/nathakusuma/auditorium-reservation-backend/internal/middleware"
	"github.com/nathakusuma/auditorium-reservation-backend/pkg/validator"
)

type registrationHandler struct {
	svc contract.IRegistrationService
	val validator.IValidator
}

func InitRegistrationHandler(
	router fiber.Router,
	middleware *middleware.Middleware,
	validator validator.IValidator,
	registrationService contract.IRegistrationService) {

	handler := registrationHandler{
		svc: registrationService,
		val: validator,
	}

	registrationGroup := router.Group("/registrations")
	registrationGroup.Use(middleware.RequireAuthenticated())

	registrationGroup.Post("",
		middleware.RequireOneOfRoles(enum.RoleUser),
		handler.register(),
	)

	registrationGroup.Get("/conferences/:id",
		handler.getRegisteredUsersByConference(),
	)

	registrationGroup.Get("/users/:id",
		handler.getRegisteredConferencesByUser(),
	)
}

func (h *registrationHandler) register() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var request struct {
			ConferenceID string `json:"conference_id" validate:"required,uuid"`
		}

		if err := c.BodyParser(&request); err != nil {
			return errorpkg.ErrFailParseRequest
		}

		if err := h.val.ValidateStruct(request); err != nil {
			return err
		}

		userID, _ := c.Locals("user.id").(uuid.UUID)

		return h.svc.Register(c.Context(), uuid.MustParse(request.ConferenceID), userID)
	}
}

func (h *registrationHandler) getRegisteredUsersByConference() fiber.Handler {
	return func(c *fiber.Ctx) error {
		conferenceID, err := uuid.Parse(c.Params("id"))
		if err != nil {
			return errorpkg.ErrFailParseRequest
		}

		var lazyReq dto.LazyLoadQuery
		if err2 := c.QueryParser(&lazyReq); err2 != nil {
			return errorpkg.ErrFailParseRequest
		}

		if err2 := h.val.ValidateStruct(lazyReq); err2 != nil {
			return err2
		}

		users, lazyResp, err := h.svc.GetRegisteredUsersByConference(c.Context(), conferenceID, lazyReq)
		if err != nil {
			return err
		}

		return c.JSON(map[string]interface{}{
			"users":      users,
			"pagination": lazyResp,
		})
	}
}

func (h *registrationHandler) getRegisteredConferencesByUser() fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, err := uuid.Parse(c.Params("id"))
		if err != nil {
			return errorpkg.ErrFailParseRequest
		}

		var lazyReq dto.LazyLoadQuery
		if err2 := c.QueryParser(&lazyReq); err2 != nil {
			return errorpkg.ErrFailParseRequest
		}

		if err2 := h.val.ValidateStruct(lazyReq); err2 != nil {
			return err2
		}

		includePast := c.Query("include_past") == "true"

		conferences, lazyResp, err := h.svc.GetRegisteredConferencesByUser(c.Context(), userID, includePast, lazyReq)
		if err != nil {
			return err
		}

		return c.JSON(map[string]interface{}{
			"conferences": conferences,
			"pagination":  lazyResp,
		})
	}
}
