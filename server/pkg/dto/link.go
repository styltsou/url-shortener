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
	ExpiresAt *time.Time `json:"expires_at"`
}

// TODO: is it better to use the error package instead of using fmt.Error()?
func (dto UpdateLink) Validate() error {
	if dto.Shortcode == nil && dto.ExpiresAt == nil {
		return errors.New("At least one of the following fields must be provided: shortcode | expires_at")
	}
	return nil
}
