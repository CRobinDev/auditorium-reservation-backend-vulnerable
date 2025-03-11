package dto

import (
	"github.com/google/uuid"
	"github.com/nathakusuma/auditorium-reservation-backend/domain/entity"
	"time"
)

type FeedbackResponse struct {
	ID        uuid.UUID     `json:"id"`
	Comment   string        `json:"comment,omitempty"`
	CreatedAt *time.Time    `json:"created_at,omitempty"`
	User      *UserResponse `json:"user,omitempty"`
}

func (f *FeedbackResponse) PopulateFromEntity(feedback *entity.Feedback) *FeedbackResponse {
	f.ID = feedback.ID
	f.Comment = feedback.Comment
	f.CreatedAt = &feedback.CreatedAt
	f.User = &UserResponse{
		ID:   feedback.UserID,
		Name: feedback.User.Name,
	}
	return f
}
