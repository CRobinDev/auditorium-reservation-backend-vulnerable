package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nathakusuma/auditorium-reservation-backend/domain/contract"
	"github.com/nathakusuma/auditorium-reservation-backend/domain/dto"
	"github.com/nathakusuma/auditorium-reservation-backend/domain/errorpkg"
	"github.com/nathakusuma/auditorium-reservation-backend/internal/middleware"
	"github.com/nathakusuma/auditorium-reservation-backend/pkg/validator"
	"net/http"
)

type authHandler struct {
	val validator.IValidator
	svc contract.IAuthService
}

func InitAuthHandler(
	router fiber.Router,
	middlewareInstance *middleware.Middleware,
	validator validator.IValidator,
	authSvc contract.IAuthService,
) {

	handler := authHandler{
		svc: authSvc,
		val: validator,
	}

	authGroup := router.Group("/auth")
	authGroup.Post("/register/otp", handler.requestOTPRegisterUser())
	authGroup.Post("/register/otp/check", handler.checkOTPRegisterUser())
	authGroup.Post("/register", handler.registerUser())
	authGroup.Post("/login", handler.loginUser())
	authGroup.Post("/refresh", handler.refreshToken())
	authGroup.Post("/logout", middlewareInstance.RequireAuthenticated(), handler.logout())
	authGroup.Post("/reset-password/otp", handler.requestOTPResetPassword())
	authGroup.Post("/reset-password", handler.resetPassword())
}

func (c *authHandler) requestOTPRegisterUser() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		var req dto.RequestOTPRegisterUserRequest
		if err := ctx.BodyParser(&req); err != nil {
			return errorpkg.ErrFailParseRequest
		}

		if err := c.val.ValidateStruct(req); err != nil {
			return err
		}

		err := c.svc.RequestOTPRegisterUser(ctx.Context(), req.Email)
		if err != nil {
			return err
		}

		return ctx.SendStatus(http.StatusNoContent)
	}
}

func (c *authHandler) checkOTPRegisterUser() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		var req dto.CheckOTPRegisterUserRequest
		if err := ctx.BodyParser(&req); err != nil {
			return errorpkg.ErrFailParseRequest
		}

		if err := c.val.ValidateStruct(req); err != nil {
			return err
		}

		err := c.svc.CheckOTPRegisterUser(ctx.Context(), req.Email, req.OTP)
		if err != nil {
			return err
		}

		return ctx.SendStatus(http.StatusNoContent)
	}
}

func (c *authHandler) registerUser() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		var req dto.RegisterUserRequest
		if err := ctx.BodyParser(&req); err != nil {
			return errorpkg.ErrFailParseRequest
		}

		if err := c.val.ValidateStruct(req); err != nil {
			return err
		}

		resp, err := c.svc.RegisterUser(ctx.Context(), req)
		if err != nil {
			return err
		}

		return ctx.Status(http.StatusCreated).JSON(resp)
	}
}

func (c *authHandler) loginUser() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		var req dto.LoginUserRequest
		if err := ctx.BodyParser(&req); err != nil {
			return errorpkg.ErrFailParseRequest
		}

		if err := c.val.ValidateStruct(req); err != nil {
			return err
		}

		resp, err := c.svc.Login(ctx.Context(), req)
		if err != nil {
			return err
		}

		return ctx.Status(http.StatusOK).JSON(resp)
	}
}

func (c *authHandler) refreshToken() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		var req dto.RefreshTokenRequest
		if err := ctx.BodyParser(&req); err != nil {
			return errorpkg.ErrFailParseRequest
		}

		if err := c.val.ValidateStruct(req); err != nil {
			return err
		}

		resp, err := c.svc.RefreshToken(ctx.Context(), req.RefreshToken)
		if err != nil {
			return err
		}

		return ctx.Status(http.StatusOK).JSON(resp)
	}
}

func (c *authHandler) logout() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		err := c.svc.Logout(ctx.Context())
		if err != nil {
			return err
		}

		return ctx.SendStatus(http.StatusNoContent)
	}
}

func (c *authHandler) requestOTPResetPassword() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		var req dto.RequestOTPResetPasswordRequest
		if err := ctx.BodyParser(&req); err != nil {
			return errorpkg.ErrFailParseRequest
		}

		if err := c.val.ValidateStruct(req); err != nil {
			return err
		}

		err := c.svc.RequestOTPResetPassword(ctx.Context(), req.Email)
		if err != nil {
			return err
		}

		return ctx.SendStatus(http.StatusNoContent)
	}
}

func (c *authHandler) resetPassword() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		var req dto.ResetPasswordRequest
		if err := ctx.BodyParser(&req); err != nil {
			return errorpkg.ErrFailParseRequest
		}

		if err := c.val.ValidateStruct(req); err != nil {
			return err
		}

		resp, err := c.svc.ResetPassword(ctx.Context(), req)
		if err != nil {
			return err
		}

		return ctx.Status(http.StatusOK).JSON(resp)
	}
}
