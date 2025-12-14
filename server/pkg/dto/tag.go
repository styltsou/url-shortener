package dto

import (
	"errors"
	"strings"

	"github.com/google/uuid"
)

// For custom validation logic, implement the Validator interface
// defined in pkg/middleware/request_validator.go

type CreateTag struct {
	Name string `json:"name" validate:"required,min=1,max=30"`
}

func (dto *CreateTag) Validate() error {
	// Trim whitespace from tag name
	dto.Name = strings.TrimSpace(dto.Name)

	// Validate name is not empty after trimming
	if dto.Name == "" {
		return errors.New("tag name cannot be empty")
	}

	return nil
}

type UpdateTag struct {
	Name string `json:"name" validate:"required,min=1,max=30"`
}

func (dto *UpdateTag) Validate() error {
	// Trim whitespace from tag name
	dto.Name = strings.TrimSpace(dto.Name)

	// Validate name is not empty after trimming
	if dto.Name == "" {
		return errors.New("tag name cannot be empty")
	}

	return nil
}

type DeleteTags struct {
	TagIDs []uuid.UUID `json:"tag_ids" validate:"required,min=1"`
}

func (dto *DeleteTags) Validate() error {
	if len(dto.TagIDs) == 0 {
		return errors.New("tag_ids cannot be empty")
	}

	// Validate all UUIDs are not nil
	for _, id := range dto.TagIDs {
		if id == uuid.Nil {
			return errors.New("all tag_ids must be valid UUIDs")
		}
	}

	return nil
}
