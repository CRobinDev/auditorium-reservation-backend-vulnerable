package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/nathakusuma/auditorium-reservation-backend/domain/contract"
	"github.com/nathakusuma/auditorium-reservation-backend/domain/dto"
	"github.com/nathakusuma/auditorium-reservation-backend/domain/errorpkg"
	"github.com/nathakusuma/auditorium-reservation-backend/internal/middleware"
	"github.com/nathakusuma/auditorium-reservation-backend/pkg/validator"
	"net/url"
)

type userHandler struct {
	val validator.IValidator
	svc contract.IUserService
}

func InitUserHandler(
	router fiber.Router,
	midw *middleware.Middleware,
	validator validator.IValidator,
	userSvc contract.IUserService,
) {
	handler := userHandler{
		svc: userSvc,
		val: validator,
	}

	userGroup := router.Group("/users")
	userGroup.Post("",
		midw.RequireAuthenticated(),
		handler.createUser(),
	)
	userGroup.Get("/me",
		midw.RequireAuthenticated(),
		handler.getUser("me"),
	)
	userGroup.Get("/:id",
		midw.RequireAuthenticated(),
		handler.getUser("id"),
	)
	userGroup.Patch("/me",
		midw.RequireAuthenticated(),
		handler.updateUser(),
	)
	userGroup.Delete("/:id",
		midw.RequireAuthenticated(),
		handler.deleteUser(),
	)
}

func (c *userHandler) createUser() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		var req dto.CreateUserRequest
		if err := ctx.BodyParser(&req); err != nil {
			return errorpkg.ErrFailParseRequest
		}

		if err := c.val.ValidateStruct(req); err != nil {
			return err
		}

		userID, err := c.svc.CreateUser(ctx.Context(), &req)
		if err != nil {
			return err
		}

		return ctx.Status(fiber.StatusCreated).JSON(map[string]interface{}{
			"user": dto.UserResponse{ID: userID},
		})
	}
}

func (c *userHandler) getUser(param string) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		var userID string
		var err error
		if param == "me" {
			oldUserID := ctx.Locals("user.id").(uuid.UUID)
			userID = oldUserID.String()
		} else {
			userID, err = url.QueryUnescape(ctx.Params("id"))
			if err != nil {
				return errorpkg.ErrFailParseRequest
			}
		}

		users, err := c.svc.GetUserByID(ctx.Context(), userID)
		if err != nil {
			return err
		}

		var respUsers []dto.UserResponse
		for _, user := range users {
			resp := dto.UserResponse{}
			if param == "me" {
				resp.PopulateFromEntity(user)
			} else {
				resp.PopulateFromEntity(user)
			}
			respUsers = append(respUsers, resp)
		}

		// Return array of users
		return ctx.Status(fiber.StatusOK).JSON(map[string]interface{}{
			"users": respUsers,
			"user":  respUsers[0],
		})
	}
}

func (c *userHandler) updateUser() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		var req dto.UpdateUserRequest
		if err := ctx.BodyParser(&req); err != nil {
			return errorpkg.ErrFailParseRequest
		}

		if err := c.val.ValidateStruct(req); err != nil {
			return err
		}

		oldUserID := ctx.Locals("user.id").(uuid.UUID)

		if err := c.svc.UpdateUser(ctx.Context(), oldUserID.String(), req); err != nil {
			return err
		}

		return ctx.SendStatus(fiber.StatusNoContent)
	}
}

func (c *userHandler) deleteUser() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		userID, err := uuid.Parse(ctx.Params("id"))
		if err != nil {
			return errorpkg.ErrFailParseRequest
		}

		if err := c.svc.DeleteUser(ctx.Context(), userID); err != nil {
			return err
		}

		return ctx.SendStatus(fiber.StatusNoContent)
	}
}
