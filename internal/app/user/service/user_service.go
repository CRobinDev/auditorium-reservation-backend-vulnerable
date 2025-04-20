package service

import (
	"context"
	"database/sql"
	"errors"
	"mime/multipart"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/nathakusuma/auditorium-reservation-backend/domain/enum"
	"github.com/nathakusuma/auditorium-reservation-backend/domain/errorpkg"
	"github.com/nathakusuma/auditorium-reservation-backend/pkg/log"
	"github.com/nathakusuma/auditorium-reservation-backend/pkg/supabase"
	"github.com/nathakusuma/auditorium-reservation-backend/pkg/uuidpkg"

	"github.com/nathakusuma/auditorium-reservation-backend/domain/contract"
	"github.com/nathakusuma/auditorium-reservation-backend/domain/dto"
	"github.com/nathakusuma/auditorium-reservation-backend/domain/entity"
)

type userService struct {
	userRepo contract.IUserRepository
	// bcrypt   bcrypt.IBcrypt
	supabase supabase.ISupabase
	uuid     uuidpkg.IUUID
}

func NewUserService(
	userRepo contract.IUserRepository,
	// bcrypt bcrypt.IBcrypt,
	supabase supabase.ISupabase,
	uuid uuidpkg.IUUID,
) contract.IUserService {
	return &userService{
		userRepo: userRepo,
		supabase: supabase,
		uuid:     uuid,
	}
}

func (s *userService) CreateUser(ctx context.Context, req *dto.CreateUserRequest) (uuid.UUID, error) {
	loggableReq := *req
	loggableReq.Password = ""

	creatorID := ctx.Value("user.id")
	if creatorID == nil {
		creatorID = "system"
	}

	// generate user ID
	userID, err := s.uuid.NewV7()
	if err != nil {
		traceID := log.ErrorWithTraceID(map[string]interface{}{
			"error":        err.Error(),
			"request":      loggableReq,
			"requester.id": creatorID,
		}, "[UserService][CreateUser] Failed to generate user ID")

		return uuid.Nil, errorpkg.ErrInternalServer.WithTraceID(traceID)
	}

	// Deleted For Cryptographic Failures Vulnerable.
	// passwordHash, err := s.bcrypt.Hash(req.Password)
	// if err != nil {
	// 	traceID := log.ErrorWithTraceID(map[string]interface{}{
	// 		"error":        err.Error(),
	// 		"request":      loggableReq,
	// 		"requester.id": creatorID,
	// 	}, "[UserService][CreateUser] Failed to hash password")

	// 	return uuid.Nil, errorpkg.ErrInternalServer.WithTraceID(traceID)
	// }

	// create user data
	user := &entity.User{
		ID:           userID,
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: req.Password,
		Role:         enum.RoleUser,
	}

	err = s.userRepo.CreateUser(ctx, user)
	if err != nil {
		// if error is due to conflict in unique constraint in email column
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.ConstraintName == "users_email_key" {
			return uuid.Nil, errorpkg.ErrEmailAlreadyRegistered
		}

		// other error
		traceID := log.ErrorWithTraceID(map[string]interface{}{
			"error":        err.Error(),
			"request":      loggableReq,
			"requester.id": creatorID,
		}, "[UserService][CreateUser] Failed to create user")

		return uuid.Nil, errorpkg.ErrInternalServer.WithTraceID(traceID)
	}

	log.Info(map[string]interface{}{
		"user":         user,
		"requester.id": creatorID,
	}, "[UserService][CreateUser] User created")

	return userID, nil
}

func (s *userService) getUserByField(ctx context.Context, field, value string) ([]*entity.User, error) {
	users, err := s.userRepo.GetUserByField(ctx, field, value)
	if err != nil {
		// if user not found
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errorpkg.ErrNotFound.WithMessage("User not found.")
		}

		// other error
		traceID := log.ErrorWithTraceID(map[string]interface{}{
			"error": err.Error(),
			"field": field,
			"value": value,
		}, "[UserService][getUserByField] Failed to get user by field")

		return nil, errorpkg.ErrInternalServer.WithTraceID(traceID)
	}

	return users, nil
}

func (s *userService) GetUserByEmail(ctx context.Context, email string) (*entity.User, error) {
	users, err := s.getUserByField(ctx, "email", email)
	if err != nil {
		return nil, err
	}

	return users[0], nil
}

func (s *userService) GetUserByID(ctx context.Context, id string) ([]*entity.User, error) {
	return s.getUserByField(ctx, "id", id)
}

func (s *userService) UpdatePassword(ctx context.Context, email, newPassword string) error {
	// get user by email
	user, err := s.GetUserByEmail(ctx, email)
	if err != nil {
		return err
	}

	// Deleted for Cryptographic Failures.
	// newPasswordHash, err := s.bcrypt.Hash(newPassword)
	// if err != nil {
	// 	traceID := log.ErrorWithTraceID(map[string]interface{}{
	// 		"error":      err.Error(),
	// 		"user.email": email,
	// 	}, "[UserService][UpdatePassword] Failed to hash password")

	// 	return errorpkg.ErrInternalServer.WithTraceID(traceID)
	// }

	// update user password
	user.PasswordHash = newPassword
	err = s.userRepo.UpdateUser(ctx, user)
	if err != nil {
		traceID := log.ErrorWithTraceID(map[string]interface{}{
			"error":      err.Error(),
			"user.email": email,
		}, "[UserService][UpdatePassword] Failed to update user password")

		return errorpkg.ErrInternalServer.WithTraceID(traceID)
	}

	log.Info(map[string]interface{}{
		"user.email": email,
	}, "[UserService][UpdatePassword] Password updated")

	return nil
}

func (s *userService) UpdateUser(ctx context.Context, id string, req dto.UpdateUserRequest) error {
	// get user by ID
	users, err := s.GetUserByID(ctx, id)
	if err != nil {
		return err
	}

	user := users[0]

	// update user data
	if req.Name != nil {
		user.Name = *req.Name
	}
	if req.Bio != nil {
		user.Bio = req.Bio
	}

	// update user
	err = s.userRepo.UpdateUser(ctx, user)
	if err != nil {
		traceID := log.ErrorWithTraceID(map[string]interface{}{
			"error": err.Error(),
			"user":  user,
		}, "[UserService][UpdateUser] Failed to update user")

		return errorpkg.ErrInternalServer.WithTraceID(traceID)
	}

	log.Info(map[string]interface{}{
		"user": user,
	}, "[UserService][UpdateUser] User updated")

	return nil
}

func (s *userService) DeleteUser(ctx context.Context, id uuid.UUID) error {
	requesterID := ctx.Value("user.id")
	if requesterID == nil {
		requesterID = "system"
	}

	// delete user
	err := s.userRepo.DeleteUser(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errorpkg.ErrNotFound
		}

		traceID := log.ErrorWithTraceID(map[string]interface{}{
			"error":        err.Error(),
			"user.id":      id,
			"requester.id": requesterID,
		}, "[UserService][DeleteUser] Failed to delete user")
		return errorpkg.ErrInternalServer.WithTraceID(traceID)
	}

	log.Info(map[string]interface{}{
		"user.id":      id,
		"requester.id": requesterID,
	}, "[UserService][DeleteUser] User deleted")

	return nil
}

func (s *userService) UploadProfile(ctx context.Context, id uuid.UUID, file *multipart.FileHeader) (string, error) {
	url, err := s.supabase.UploadFile(file, "profile")
	if err != nil {
		traceID := log.ErrorWithTraceID(map[string]interface{}{
			"error": err.Error(),
		}, "[UserService][UploadProfile] Failed to update user")

		return "", errorpkg.ErrInternalServer.WithTraceID(traceID)
	}

	err = s.userRepo.UploadProfile(ctx, id, url)
	if err != nil {
		traceID := log.ErrorWithTraceID(map[string]interface{}{
			"error": err.Error(),
		}, "[UserService][UploadProfile] Failed to input url to database ")

		return "", errorpkg.ErrInternalServer.WithTraceID(traceID)
	}

	return url, nil
}
