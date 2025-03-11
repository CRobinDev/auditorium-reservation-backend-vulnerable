package contract

import (
	"context"
	"github.com/google/uuid"
	"github.com/nathakusuma/auditorium-reservation-backend/domain/dto"
	"github.com/nathakusuma/auditorium-reservation-backend/domain/entity"
)

type IAuthRepository interface {
	SetOTPRegisterUser(ctx context.Context, email, otp string) error
	GetOTPRegisterUser(ctx context.Context, email string) (string, error)
	DeleteOTPRegisterUser(ctx context.Context, email string) error

	CreateAuthSession(ctx context.Context, authSession *entity.AuthSession) error
	GetAuthSessionByToken(ctx context.Context, token string) (*entity.AuthSession, error)
	DeleteAuthSession(ctx context.Context, userID uuid.UUID) error

	SetOTPResetPassword(ctx context.Context, email, otp string) error
	GetOTPResetPassword(ctx context.Context, email string) (string, error)
	DeleteOTPResetPassword(ctx context.Context, email string) error
}

type IAuthService interface {
	RequestOTPRegisterUser(ctx context.Context, email string) error
	CheckOTPRegisterUser(ctx context.Context, email, otp string) error
	RegisterUser(ctx context.Context, req dto.RegisterUserRequest) (dto.LoginResponse, error)
	Login(ctx context.Context, req dto.LoginUserRequest) (dto.LoginResponse, error)

	RefreshToken(ctx context.Context, refreshToken string) (dto.LoginResponse, error)
	Logout(ctx context.Context) error

	RequestOTPResetPassword(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, req dto.ResetPasswordRequest) (dto.LoginResponse, error)
}
