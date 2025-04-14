package service

import (
	"context"
	"database/sql"
	"errors"
	"net/url"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/nathakusuma/auditorium-reservation-backend/domain/contract"
	"github.com/nathakusuma/auditorium-reservation-backend/domain/dto"
	"github.com/nathakusuma/auditorium-reservation-backend/domain/entity"
	"github.com/nathakusuma/auditorium-reservation-backend/domain/enum"
	"github.com/nathakusuma/auditorium-reservation-backend/domain/errorpkg"
	"github.com/nathakusuma/auditorium-reservation-backend/internal/infra/env"
	"github.com/nathakusuma/auditorium-reservation-backend/pkg/jwt"
	"github.com/nathakusuma/auditorium-reservation-backend/pkg/log"
	"github.com/nathakusuma/auditorium-reservation-backend/pkg/mail"
	"github.com/nathakusuma/auditorium-reservation-backend/pkg/randgen"
	"github.com/nathakusuma/auditorium-reservation-backend/pkg/uuidpkg"
	"github.com/redis/go-redis/v9"
)

type authService struct {
	repo    contract.IAuthRepository
	userSvc contract.IUserService
	// bcrypt  bcrypt.IBcrypt
	jwt    jwt.IJwt
	mailer mail.IMailer
	uuid   uuidpkg.IUUID
}

func NewAuthService(
	authRepo contract.IAuthRepository,
	userSvc contract.IUserService,
	// bcrypt bcrypt.IBcrypt,
	jwt jwt.IJwt,
	mailer mail.IMailer,
	uuid uuidpkg.IUUID,
) contract.IAuthService {
	return &authService{
		repo:    authRepo,
		userSvc: userSvc,
		// bcrypt:  bcrypt,
		jwt:    jwt,
		mailer: mailer,
		uuid:   uuid,
	}
}

func (s *authService) RequestOTPRegisterUser(ctx context.Context, email string) error {
	// check if email is already registered
	_, err := s.userSvc.GetUserByEmail(ctx, email)
	if err == nil {
		return errorpkg.ErrEmailAlreadyRegistered
	}

	if !errors.Is(err, errorpkg.ErrNotFound) {
		traceID := log.ErrorWithTraceID(map[string]interface{}{
			"error":      err.Error(),
			"user.email": email,
		}, "[AuthService][RequestOTPRegisterUser] failed to get user by email")

		return errorpkg.ErrInternalServer.WithTraceID(traceID)
	}

	// generate otp
	otp := strconv.Itoa(randgen.RandomNumber(6))

	// save otp
	err = s.repo.SetOTPRegisterUser(ctx, email, otp)
	if err != nil {
		traceID := log.ErrorWithTraceID(map[string]interface{}{
			"error":      err.Error(),
			"user.email": email,
		}, "[AuthService][RequestOTPRegisterUser] failed to save otp")

		return errorpkg.ErrInternalServer.WithTraceID(traceID)
	}

	// send otp to email
	go func() {
		err = s.mailer.Send(
			email,
			"[Auditorium Reservation] Verify Your Account",
			"otp_register_user.html",
			map[string]interface{}{
				"otp":  otp,
				"href": env.GetEnv().FrontendURL + "/verify-email?email=" + url.QueryEscape(email) + "&otp=" + otp,
			})

		if err != nil {
			log.Error(map[string]interface{}{
				"error": err.Error(),
			}, "[AuthService][RequestOTPRegisterUser] failed to send email")
		}
	}()

	log.Info(map[string]interface{}{
		"user.email": email,
	}, "[AuthService][RequestOTPRegisterUser] otp requested")

	return nil
}

func (s *authService) CheckOTPRegisterUser(ctx context.Context, email, otp string) error {
	savedOtp, err := s.repo.GetOTPRegisterUser(ctx, email)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return errorpkg.ErrInvalidOTP
		}

		traceID := log.ErrorWithTraceID(map[string]interface{}{
			"error":      err.Error(),
			"user.email": email,
		}, "[AuthService][CheckOTPRegisterUser] failed to get otp")

		return errorpkg.ErrInternalServer.WithTraceID(traceID)
	}

	if savedOtp != otp {
		return errorpkg.ErrInvalidOTP
	}

	return nil
}

func (s *authService) RegisterUser(ctx context.Context,
	req dto.RegisterUserRequest) (dto.LoginResponse, error) {

	var resp dto.LoginResponse

	// req without Password and OTP
	loggableReq := req
	loggableReq.Password = ""
	loggableReq.OTP = ""

	// get otp
	savedOtp, err := s.repo.GetOTPRegisterUser(ctx, req.Email)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return resp, errorpkg.ErrInvalidOTP
		}

		traceID := log.ErrorWithTraceID(map[string]interface{}{
			"error":   err.Error(),
			"request": loggableReq,
		}, "[AuthService][RegisterUser] failed to get otp")

		return resp, errorpkg.ErrInternalServer.WithTraceID(traceID)
	}

	if savedOtp != req.OTP {
		return resp, errorpkg.ErrInvalidOTP
	}

	// delete otp
	err = s.repo.DeleteOTPRegisterUser(ctx, req.Email)
	if err != nil {
		traceID := log.ErrorWithTraceID(map[string]interface{}{
			"error":   err.Error(),
			"request": loggableReq,
		}, "[AuthService][RegisterUser] failed to delete otp")

		return resp, errorpkg.ErrInternalServer.WithTraceID(traceID)
	}

	// save user
	_, err = s.userSvc.CreateUser(ctx, &dto.CreateUserRequest{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
		Role:     enum.RoleUser,
	})
	if err != nil {
		traceID := log.ErrorWithTraceID(map[string]interface{}{
			"error":   err.Error(),
			"request": loggableReq,
		}, "[AuthService][RegisterUser] failed to create user")

		return resp, errorpkg.ErrInternalServer.WithTraceID(traceID)
	}

	log.Info(map[string]interface{}{
		"user.email": req.Email,
	}, "[AuthService][RegisterUser] user registered")

	// login user
	return s.Login(ctx, dto.LoginUserRequest{
		Email:    req.Email,
		Password: req.Password,
	})
}

func (s *authService) Login(ctx context.Context, req dto.LoginUserRequest) (dto.LoginResponse, error) {
	var resp dto.LoginResponse

	// get user by email
	user, err := s.userSvc.GetUserByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, errorpkg.ErrNotFound) {
			return resp, errorpkg.ErrNotFound.WithMessage("User not found. Please register first.")
		}

		traceID := log.ErrorWithTraceID(map[string]interface{}{
			"error":      err.Error(),
			"user.email": req.Email,
		}, "[AuthService][Login] failed to get user by email")

		return resp, errorpkg.ErrInternalServer.WithTraceID(traceID)
	}

	// check password (plain text)
	ok := req.Password == user.PasswordHash
	if !ok {
		return resp, errorpkg.ErrCredentialsNotMatch
	}

	// Generate access token first
	accessToken, err := s.jwt.Create(user.ID, user.Role)
	if err != nil {
		traceID := log.ErrorWithTraceID(map[string]interface{}{
			"error":      err.Error(),
			"user.email": req.Email,
		}, "[AuthService][Login] failed to generate access token")
		return resp, errorpkg.ErrInternalServer.WithTraceID(traceID)
	}

	// Generate and store refresh token
	refreshToken := randgen.RandomString(32)
	err = s.repo.CreateAuthSession(ctx, &entity.AuthSession{
		Token:     refreshToken,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(env.GetEnv().JwtRefreshExpireDuration),
	})
	if err != nil {
		traceID := log.ErrorWithTraceID(map[string]interface{}{
			"error":      err.Error(),
			"user.email": req.Email,
		}, "[AuthService][Login] failed to store auth session")
		return resp, errorpkg.ErrInternalServer.WithTraceID(traceID)
	}

	userResp := dto.UserResponse{}
	userResp.PopulateFromEntity(user)
	resp = dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         &userResp,
	}

	log.Info(map[string]interface{}{
		"user.email": req.Email,
	}, "[AuthService][Login] user logged in")

	return resp, nil
}

func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (dto.LoginResponse, error) {
	var resp dto.LoginResponse

	authSession, err := s.repo.GetAuthSessionByToken(ctx, refreshToken)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return resp, errorpkg.ErrInvalidRefreshToken
		}

		traceID := log.ErrorWithTraceID(map[string]interface{}{
			"error": err.Error(),
		}, "[AuthService][RefreshToken] failed to get auth session by token")

		return resp, errorpkg.ErrInternalServer.WithTraceID(traceID)
	}

	if authSession.ExpiresAt.Before(time.Now()) {
		return resp, errorpkg.ErrInvalidRefreshToken
	}

	// get user by authSession
	user, err := s.userSvc.GetUserByID(ctx, authSession.UserID)
	if err != nil {
		if errors.Is(err, errorpkg.ErrNotFound) {
			return resp, errorpkg.ErrNotFound.WithMessage("User not found. Please register first.")
		}

		traceID := log.ErrorWithTraceID(map[string]interface{}{
			"error":   err.Error(),
			"user.id": authSession.UserID,
		}, "[AuthService][RefreshToken] failed to get user by ID")

		return resp, errorpkg.ErrInternalServer.WithTraceID(traceID)
	}

	accessToken, err := s.jwt.Create(user.ID, user.Role)
	if err != nil {
		traceID := log.ErrorWithTraceID(map[string]interface{}{
			"error":   err.Error(),
			"user.id": user.ID,
		}, "[AuthService][RefreshToken] failed to generate access token")

		return resp, errorpkg.ErrInternalServer.WithTraceID(traceID)
	}

	userResp := dto.UserResponse{}
	userResp.PopulateFromEntity(user)
	resp = dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         &userResp,
	}

	log.Info(map[string]interface{}{
		"user.id":    user.ID,
		"user.email": user.Email,
	}, "[AuthService][RefreshToken] token refreshed")

	return resp, nil
}

func (s *authService) Logout(ctx context.Context) error {
	userID := ctx.Value("user.id").(uuid.UUID)

	err := s.repo.DeleteAuthSession(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errorpkg.ErrInvalidBearerToken
		}

		traceID := log.ErrorWithTraceID(map[string]interface{}{
			"error":   err.Error(),
			"user.id": userID,
		}, "[AuthService][Logout] failed to delete auth session")

		return errorpkg.ErrInternalServer.WithTraceID(traceID)
	}

	log.Info(map[string]interface{}{
		"user.id": userID,
	}, "[AuthService][Logout] user logged out")

	return nil
}

func (s *authService) RequestOTPResetPassword(ctx context.Context, email string) error {
	// check if email is registered
	_, err := s.userSvc.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, errorpkg.ErrNotFound) {
			return errorpkg.ErrNotFound.WithMessage("User not found. Please register.")
		}

		traceID := log.ErrorWithTraceID(map[string]interface{}{
			"error":      err.Error(),
			"user.email": email,
		}, "[AuthService][RequestOTPResetPassword] failed to get user by email")

		return errorpkg.ErrInternalServer.WithTraceID(traceID)
	}

	// generate otp
	otp := strconv.Itoa(randgen.RandomNumber(6))

	// save otp
	err = s.repo.SetOTPResetPassword(ctx, email, otp)
	if err != nil {
		traceID := log.ErrorWithTraceID(map[string]interface{}{
			"error":      err.Error(),
			"user.email": email,
		}, "[AuthService][RequestOTPResetPassword] failed to save otp")

		return errorpkg.ErrInternalServer.WithTraceID(traceID)
	}

	// send otp to email
	go func() {
		err = s.mailer.Send(
			email,
			"[Auditorium Reservation] Reset Password",
			"otp_reset_password.html",
			map[string]interface{}{
				"otp": otp,
			})

		if err != nil {
			log.Error(map[string]interface{}{
				"error": err.Error(),
			}, "[AuthService][RequestOTPResetPassword] failed to send email")
		}
	}()

	log.Info(map[string]interface{}{
		"user.email": email,
	}, "[AuthService][RequestOTPResetPassword] otp requested")

	return nil
}

func (s *authService) ResetPassword(ctx context.Context, req dto.ResetPasswordRequest) (dto.LoginResponse, error) {
	// get otp
	savedOtp, err := s.repo.GetOTPResetPassword(ctx, req.Email)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return dto.LoginResponse{}, errorpkg.ErrInvalidOTP
		}

		traceID := log.ErrorWithTraceID(map[string]interface{}{
			"error":      err.Error(),
			"user.email": req.Email,
		}, "[AuthService][ResetPassword] failed to get otp")

		return dto.LoginResponse{}, errorpkg.ErrInternalServer.WithTraceID(traceID)
	}

	if savedOtp != req.OTP {
		return dto.LoginResponse{}, errorpkg.ErrInvalidOTP
	}

	// delete otp
	err = s.repo.SetOTPResetPassword(ctx, req.Email, "")
	if err != nil {
		traceID := log.ErrorWithTraceID(map[string]interface{}{
			"error":      err.Error(),
			"user.email": req.Email,
		}, "[AuthService][ResetPassword] failed to delete otp")

		return dto.LoginResponse{}, errorpkg.ErrInternalServer.WithTraceID(traceID)
	}

	// update user password
	if err = s.userSvc.UpdatePassword(ctx, req.Email, req.NewPassword); err != nil {
		if errors.Is(err, errorpkg.ErrNotFound) {
			// Small chance, since we've already checked it on RequestOTPResetPassword
			return dto.LoginResponse{}, err
		}

		traceID := log.ErrorWithTraceID(map[string]interface{}{
			"error":      err.Error(),
			"user.email": req.Email,
		}, "[AuthService][ResetPassword] failed to update user password")

		return dto.LoginResponse{}, errorpkg.ErrInternalServer.WithTraceID(traceID)
	}

	log.Info(map[string]interface{}{
		"user.email": req.Email,
	}, "[AuthService][ResetPassword] password reset")

	return s.Login(ctx, dto.LoginUserRequest{
		Email:    req.Email,
		Password: req.NewPassword,
	})
}
