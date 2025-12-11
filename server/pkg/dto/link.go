package dto

import (
	"errors"
	"time"
)

// For custom validation logic, implement the Validator interface
// defined in pkg/middleware/request_validator.go

type CreateLink struct {
	URL       string     `json:"url" validate:"required"`
	Shortcode *string    `json:"shortcode" validate:"omitempty,min=1"`
	ExpiresAt *time.Time `json:"expires_at" validate:"omitempty"`
}

type UpdateLink struct {
	Shortcode *string    `json:"shortcode"`
	IsActive  *bool      `json:"is_active"`
	ExpiresAt *time.Time `json:"expires_at"`
}

func (dto UpdateLink) Validate() error {
	if dto.Shortcode == nil && dto.IsActive == nil && dto.ExpiresAt == nil {
		return errors.New("At least one of the following fields must be provided: shortcode | is_active | expires_at")
	}

	if dto.ExpiresAt != nil && dto.ExpiresAt.Before(time.Now()) {
		return errors.New("expires_at must be set to a future time")
	}

	return nil
}
