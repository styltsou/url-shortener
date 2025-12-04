package errors

import (
	"errors"
)

type ErrorCode string

const (
	CodeInvalidRequest ErrorCode = "invalid_request"

	CodeAuthRequired ErrorCode = "authentication_required"
	CodeAuthFailed   ErrorCode = "authentication_failed"

	CodeLinkNotFound ErrorCode = "link_not_found"
	CodeInvalidURL   ErrorCode = "invalid_url"
	CodeLinkExpired  ErrorCode = "link_expired"
	CodeCodeTaken    ErrorCode = "code_taken"

	CodeInternalError ErrorCode = "internal_server_error"
)

// Sentinel errors - use these in services, check with errors.Is()
var (
	AuthRequired = errors.New("Authentication required")
	AuthFailed   = errors.New("Authentication failed")

	LinkNotFound = errors.New("Link not found")
	InvalidURL   = errors.New("Invalid URL")
	LinkExpired  = errors.New("Link expired")

	InternalError = errors.New("Internal server error")
)
