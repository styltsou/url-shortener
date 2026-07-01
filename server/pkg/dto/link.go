package dto

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/styltsou/url-shortener/server/pkg/db"
)

// For custom validation logic, implement the Validator interface
// defined in pkg/middleware/request_validator.go

type CreateLink struct {
	URL       string      `json:"url" validate:"required"`
	Shortcode *string     `json:"shortcode" validate:"omitempty,min=1"`
	ExpiresAt *time.Time  `json:"expires_at" validate:"omitempty"`
	TagIDs    []uuid.UUID `json:"tag_ids" validate:"omitempty,dive"`
}

func (dto *CreateLink) Validate() error {
	if dto.ExpiresAt != nil && dto.ExpiresAt.Before(time.Now()) {
		return errors.New("expires_at must be set to a future time")
	}
	return nil
}

type UpdateLink struct {
	Shortcode *string      `json:"shortcode"`
	IsActive  *bool        `json:"is_active"`
	ExpiresAt OptionalTime `json:"expires_at"`
}

func (dto *UpdateLink) Validate() error {
	if dto.Shortcode == nil && dto.IsActive == nil && !dto.ExpiresAt.Set {
		return errors.New("At least one of the following fields must be provided: shortcode | is_active | expires_at")
	}

	if dto.ExpiresAt.Set && dto.ExpiresAt.Value != nil && dto.ExpiresAt.Value.Before(time.Now()) {
		return errors.New("expires_at must be set to a future time")
	}

	return nil
}

type OptionalTime struct {
	Set   bool
	Value *time.Time
}

func (t *OptionalTime) UnmarshalJSON(data []byte) error {
	t.Set = true
	if string(data) == "null" {
		t.Value = nil
		return nil
	}

	var value time.Time
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	t.Value = &value
	return nil
}

type AddTagsToLink struct {
	TagIDs []uuid.UUID `json:"tag_ids" validate:"required,min=1"`
}

type RemoveTagsFromLink struct {
	TagIDs []uuid.UUID `json:"tag_ids" validate:"required,min=1"`
}

type TagResponse struct {
	ID        uuid.UUID  `json:"id"`
	Name      string     `json:"name"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}

type LinkResponse struct {
	ID          uuid.UUID     `json:"id"`
	Shortcode   string        `json:"shortcode"`
	OriginalURL string        `json:"original_url"`
	ExpiresAt   *time.Time    `json:"expires_at"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   *time.Time    `json:"updated_at"`
	IsActive    bool          `json:"is_active"`
	Tags        []TagResponse `json:"tags"`
}

func LinkFromCreateRow(row db.TryCreateLinkRow) LinkResponse {
	return LinkResponse{
		ID:          row.ID,
		Shortcode:   row.Shortcode,
		OriginalURL: row.OriginalUrl,
		ExpiresAt:   timestampPtr(row.ExpiresAt),
		CreatedAt:   timestampValue(row.CreatedAt),
		UpdatedAt:   timestampPtr(row.UpdatedAt),
		IsActive:    row.IsActive,
		Tags:        []TagResponse{},
	}
}

func LinkFromListRow(row db.ListUserLinksRow) (LinkResponse, error) {
	return linkWithTags(
		row.ID,
		row.Shortcode,
		row.OriginalUrl,
		row.ExpiresAt,
		row.CreatedAt,
		row.UpdatedAt,
		row.IsActive,
		row.Tags,
	)
}

func LinkFromShortcodeRow(row db.GetLinkByShortcodeAndUserRow) (LinkResponse, error) {
	return linkWithTags(
		row.ID,
		row.Shortcode,
		row.OriginalUrl,
		row.ExpiresAt,
		row.CreatedAt,
		row.UpdatedAt,
		row.IsActive,
		row.Tags,
	)
}

func LinkFromWithTagsRow(row db.GetLinkByIdAndUserWithTagsRow) (LinkResponse, error) {
	return linkWithTags(
		row.ID,
		row.Shortcode,
		row.OriginalUrl,
		row.ExpiresAt,
		row.CreatedAt,
		row.UpdatedAt,
		row.IsActive,
		row.Tags,
	)
}

func LinkFromUpdateRow(row db.UpdateLinkRow) LinkResponse {
	return LinkResponse{
		ID:          row.ID,
		Shortcode:   row.Shortcode,
		OriginalURL: row.OriginalUrl,
		ExpiresAt:   timestampPtr(row.ExpiresAt),
		CreatedAt:   timestampValue(row.CreatedAt),
		UpdatedAt:   timestampPtr(row.UpdatedAt),
		IsActive:    row.IsActive,
		Tags:        []TagResponse{},
	}
}

func LinkFromDeleteRow(row db.DeleteLinkRow) LinkResponse {
	return LinkResponse{
		ID:          row.ID,
		Shortcode:   row.Shortcode,
		OriginalURL: row.OriginalUrl,
		ExpiresAt:   timestampPtr(row.ExpiresAt),
		CreatedAt:   timestampValue(row.CreatedAt),
		UpdatedAt:   timestampPtr(row.UpdatedAt),
		IsActive:    row.IsActive,
		Tags:        []TagResponse{},
	}
}

func LinkFromRecentRow(row db.GetRecentLinksRow) LinkResponse {
	return LinkResponse{
		ID:          row.ID,
		Shortcode:   row.Shortcode,
		OriginalURL: row.OriginalUrl,
		ExpiresAt:   timestampPtr(row.ExpiresAt),
		CreatedAt:   timestampValue(row.CreatedAt),
		UpdatedAt:   timestampPtr(row.UpdatedAt),
		IsActive:    row.IsActive,
		Tags:        []TagResponse{},
	}
}

func LinksFromListRows(rows []db.ListUserLinksRow) ([]LinkResponse, error) {
	links := make([]LinkResponse, 0, len(rows))
	for _, row := range rows {
		link, err := LinkFromListRow(row)
		if err != nil {
			return nil, err
		}
		links = append(links, link)
	}
	return links, nil
}

func LinksFromRecentRows(rows []db.GetRecentLinksRow) []LinkResponse {
	links := make([]LinkResponse, 0, len(rows))
	for _, row := range rows {
		links = append(links, LinkFromRecentRow(row))
	}
	return links
}

func linkWithTags(id uuid.UUID, shortcode, originalURL string, expiresAt, createdAt, updatedAt pgtype.Timestamp, isActive bool, tagsValue interface{}) (LinkResponse, error) {
	tags, err := decodeTags(tagsValue)
	if err != nil {
		return LinkResponse{}, err
	}
	return LinkResponse{
		ID:          id,
		Shortcode:   shortcode,
		OriginalURL: originalURL,
		ExpiresAt:   timestampPtr(expiresAt),
		CreatedAt:   timestampValue(createdAt),
		UpdatedAt:   timestampPtr(updatedAt),
		IsActive:    isActive,
		Tags:        tags,
	}, nil
}

func decodeTags(value interface{}) ([]TagResponse, error) {
	if value == nil {
		return []TagResponse{}, nil
	}
	if tags, ok := value.([]TagResponse); ok {
		return tags, nil
	}

	var raw []byte
	switch v := value.(type) {
	case []byte:
		raw = v
	case string:
		raw = []byte(v)
	default:
		raw, _ = json.Marshal(v)
	}
	if len(raw) == 0 {
		return []TagResponse{}, nil
	}

	var tags []TagResponse
	if err := json.Unmarshal(raw, &tags); err != nil {
		return nil, err
	}
	if tags == nil {
		return []TagResponse{}, nil
	}
	return tags, nil
}

func timestampPtr(ts pgtype.Timestamp) *time.Time {
	if !ts.Valid {
		return nil
	}
	return &ts.Time
}

func timestampValue(ts pgtype.Timestamp) time.Time {
	if !ts.Valid {
		return time.Time{}
	}
	return ts.Time
}
