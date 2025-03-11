package server

import (
	"errors"
	"github.com/nathakusuma/auditorium-reservation-backend/domain/errorpkg"

	"github.com/gofiber/fiber/v2"
	"github.com/nathakusuma/auditorium-reservation-backend/pkg/validator"
)

func ErrorHandler() fiber.ErrorHandler {
	return func(ctx *fiber.Ctx, err error) error {
		var apiErr *errorpkg.ErrorResponse
		if errors.As(err, &apiErr) {
			return ctx.Status(apiErr.HttpStatusCode).JSON(apiErr)
		}

		var validationErr validator.ValidationErrorsResponse
		if errors.As(err, &validationErr) {
			return ctx.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
				"message":    "There are invalid fields in your request. Please check and try again",
				"detail":     validationErr,
				"error_code": "VALIDATION_ERROR",
			})
		}

		var fiberErr *fiber.Error
		if errors.As(err, &fiberErr) {
			return ctx.Status(fiberErr.Code).JSON(fiber.Map{
				"message": fiberErr.Message,
			})
		}

		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Internal server error",
		})
	}
}
