package dto

import "github.com/google/uuid"

type LazyLoadResponse struct {
	HasMore bool        `json:"has_more"`
	FirstID interface{} `json:"first_id"`
	LastID  interface{} `json:"last_id"`
}

type LazyLoadQuery struct {
	AfterID  uuid.UUID `query:"after_id"`
	BeforeID uuid.UUID `query:"before_id"`
	Limit    int       `query:"limit" validate:"required,min=1,max=20"`
}
