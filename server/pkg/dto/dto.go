package dto

type CreateLinkRequest struct {
	URL string `json:"url"`
}

// TODO: Will add more request DTOs later

type SuccessReponse[T any] struct {
	Data    T      `json:"data"`
	Message string `json:"message,omitempty"`
}

type ErrorCode string

// * These error code stuff might be a little to much but nvm
const (
	// Generic errors
	ErrorBadRequest   ErrorCode = "BAD_REQUEST"
	ErrorUnauthorized ErrorCode = "UNAUTHORIZED"
	ErrorForbidden    ErrorCode = "FORBIDDEN"
	ErrorNotFound     ErrorCode = "NOT_FOUND"
	ErrorInternal     ErrorCode = "INTERNAL_ERROR"

	// Domain-specific: Link resource
	ErrorLinkNotFound   ErrorCode = "LINK_NOT_FOUND"
	ErrorInvalidLinkURL ErrorCode = "INVALID_LINK_URL"

	// Domain-specific: User resource
	ErrorUserNotFound ErrorCode = "USER_NOT_FOUND"
)

type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
}
