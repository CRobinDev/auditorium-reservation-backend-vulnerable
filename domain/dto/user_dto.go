package dto

import (
	"github.com/google/uuid"
	"github.com/nathakusuma/auditorium-reservation-backend/domain/entity"
	"github.com/nathakusuma/auditorium-reservation-backend/domain/enum"
	"time"
)

type UserResponse struct {
	ID        uuid.UUID     `json:"id,omitempty"`
	Name      string        `json:"name,omitempty"`
	Email     string        `json:"email,omitempty"`
	Role      enum.UserRole `json:"role,omitempty,omitempty"`
	Bio       *string       `json:"bio,omitempty"`
	CreatedAt *time.Time    `json:"created_at,omitempty"`
	UpdatedAt *time.Time    `json:"updated_at,omitempty"`
}

func (u *UserResponse) PopulateFromEntity(user *entity.User) *UserResponse {
	u.ID = user.ID
	u.Name = user.Name
	u.Email = user.Email
	u.Role = user.Role
	u.Bio = user.Bio
	u.CreatedAt = &user.CreatedAt
	u.UpdatedAt = &user.UpdatedAt
	return u
}

func (u *UserResponse) PopulateMinimalFromEntity(user *entity.User) *UserResponse {
	u.ID = user.ID
	u.Name = user.Name
	u.Role = user.Role
	u.Bio = user.Bio
	return u
}

type CreateUserRequest struct {
	Name     string        `json:"name" validate:"required,min=3,max=100,ascii"`
	Email    string        `json:"email" validate:"required,email,max=320"`
	Password string        `json:"password" validate:"required,min=8,max=72,ascii"`
	Role     enum.UserRole `json:"role" validate:"required"`
}

type UpdateUserRequest struct {
	Name *string `json:"name" validate:"omitempty,min=3,max=100,ascii"`
	Bio  *string `json:"bio" validate:"omitempty,max=500"`
}
