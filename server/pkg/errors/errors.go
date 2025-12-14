package errors

import (
	"errors"
)

type ErrorCode string

const (
	CodeInvalidRequest ErrorCode = "invalid_request"

	CodeAuthRequired ErrorCode = "authentication_required"
	CodeAuthFailed   ErrorCode = "authentication_failed"

	CodeInvalidID ErrorCode = "invalid id"

	CodeLinkNotFound ErrorCode = "link_not_found"
	CodeInvalidURL   ErrorCode = "invalid_url"
	CodeLinkExpired  ErrorCode = "link_expired"
	CodeCodeTaken    ErrorCode = "code_taken"
	CodeTagNotFound  ErrorCode = "tag_not_found"
	CodeTagNameTaken ErrorCode = "tag_name_taken"

	CodeNotFound         ErrorCode = "not_found"
	CodeMethodNotAllowed ErrorCode = "method_not_allowed"

	CodeInternalError ErrorCode = "internal_server_error"
)

// Sentinel errors - use these in services, check with errors.Is()
var (
	AuthRequired = errors.New("Authentication required")
	AuthFailed   = errors.New("Authentication failed")

	LinkNotFound       = errors.New("Link not found")
	InvalidURL         = errors.New("Invalid URL")
	LinkExpired        = errors.New("Link expired")
	LinkShortcodeTaken = errors.New("Shortcode already taken")
	TagNotFound        = errors.New("Tag not found")
	TagNameTaken       = errors.New("Tag name already taken")

	InternalError = errors.New("Internal server error")
)
