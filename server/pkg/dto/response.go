package dto

import (
	apperrors "github.com/styltsou/url-shortener/server/pkg/errors"
)

// SuccessResponse represents a successful API response
type SuccessResponse[T any] struct {
	Data T `json:"data"`
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
