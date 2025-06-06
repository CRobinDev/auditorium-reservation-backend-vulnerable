package contract

import (
	"context"
	"mime/multipart"

	"github.com/google/uuid"
	"github.com/nathakusuma/auditorium-reservation-backend/domain/dto"
	"github.com/nathakusuma/auditorium-reservation-backend/domain/entity"
)

type IUserRepository interface {
	CreateUser(ctx context.Context, user *entity.User) error
	GetUserByField(ctx context.Context, field, value string) (*entity.User, error)
	UpdateUser(ctx context.Context, user *entity.User) error
	DeleteUser(ctx context.Context, id uuid.UUID) error
	UploadProfile(ctx context.Context, id uuid.UUID, url string) error
}

type IUserService interface {
	CreateUser(ctx context.Context, req *dto.CreateUserRequest) (uuid.UUID, error)
	GetUserByEmail(ctx context.Context, email string) (*entity.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	UpdatePassword(ctx context.Context, email, newPassword string) error
	UpdateUser(ctx context.Context, id uuid.UUID, req dto.UpdateUserRequest) error
	DeleteUser(ctx context.Context, id uuid.UUID) error
	UploadProfile(ctx context.Context, id uuid.UUID, file *multipart.FileHeader) (string, error)
}
