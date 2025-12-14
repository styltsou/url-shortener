package dto

import (
	apperrors "github.com/styltsou/url-shortener/server/pkg/errors"
)

// SuccessResponse represents a successful API response
// Pagination is optional - only included for paginated endpoints
type SuccessResponse[T any] struct {
	Data       T               `json:"data"`
	Pagination *PaginationMeta `json:"pagination,omitempty"`
}

// PaginatedResponse is deprecated - use SuccessResponse with Pagination field instead
// Kept for backwards compatibility if needed
type PaginatedResponse[T any] struct {
	Data       T              `json:"data"`
	Pagination PaginationMeta `json:"pagination"`
}

// PaginationMeta contains pagination metadata
type PaginationMeta struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// ErrorResponse represents an error API response
type ErrorResponse struct {
	Error ErrorObject `json:"error"`
}

// ErrorObject follows a simplified version of the RFC7807 spec (https://tools.ietf.org/html/rfc7807)
type ErrorObject struct {
	Code   apperrors.ErrorCode `json:"code"`
	Title  string              `json:"title"`
	Detail string              `json:"detail"`
}
