package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"github.com/styltsou/url-shortener/server/pkg/dto"
	apperrors "github.com/styltsou/url-shortener/server/pkg/errors"
	"github.com/styltsou/url-shortener/server/pkg/logger"
	"go.uber.org/zap"
)

var validate = validator.New()

const reqBodyKey contextKey = "request_body"

// ReqBodyKey returns the context key used for request body storage.
// This is exported for testing purposes.
func ReqBodyKey() contextKey {
	return reqBodyKey
}

// Validator defines the interface for custom validation logic on DTOs.
// DTOs can implement this interface to add custom validation beyond struct tags.
type Validator interface {
	Validate() error
}

const maxBodySize = 1 << 20 // 1MB - prevents memory exhaustion from large request bodies

func RequestValidator[T any](logger logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Limit request body size to prevent memory exhaustion attacks
			r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)

			var bodyDTO T

			if err := json.NewDecoder(r.Body).Decode(&bodyDTO); err != nil {
				// Check if error is due to request body being too large
				var maxBytesError *http.MaxBytesError
				if errors.As(err, &maxBytesError) {
					logger.Warn("Request body too large",
						zap.Error(err),
						zap.String("method", r.Method),
						zap.String("path", r.URL.Path),
						zap.Int64("max_size", maxBodySize),
					)
					render.Status(r, http.StatusRequestEntityTooLarge)
					render.JSON(w, r, dto.ErrorResponse{
						Error: dto.ErrorObject{
							Code:   apperrors.CodeInvalidRequest,
							Title:  "Request body too large",
							Detail: fmt.Sprintf("Request body exceeds maximum size of %d bytes", maxBodySize),
						},
					})
					return
				}

				// Handle other decode errors (invalid JSON, etc.)
				if err == io.EOF {
					logger.Warn("Empty request body",
						zap.String("method", r.Method),
						zap.String("path", r.URL.Path),
					)
					render.Status(r, http.StatusBadRequest)
					render.JSON(w, r, dto.ErrorResponse{
						Error: dto.ErrorObject{
							Code:   apperrors.CodeInvalidRequest,
							Title:  "Invalid request body",
							Detail: "Request body is required",
						},
					})
					return
				}

				logger.Warn("Failed to decode request body",
					zap.Error(err),
					zap.String("method", r.Method),
					zap.String("path", r.URL.Path),
				)
				render.Status(r, http.StatusBadRequest)
				render.JSON(w, r, dto.ErrorResponse{
					Error: dto.ErrorObject{
						Code:   apperrors.CodeInvalidRequest,
						Title:  "Invalid request body",
						Detail: "Request payload is not valid JSON or does not match the expected schema",
					},
				})
				return
			}

			if err := validate.Struct(&bodyDTO); err != nil {
				validationErrors, ok := err.(validator.ValidationErrors)

				if !ok {
					logger.Error("Unexpected validation error type",
						zap.Error(err),
						zap.String("method", r.Method),
						zap.String("path", r.URL.Path),
					)
					render.Status(r, http.StatusBadRequest)
					render.JSON(w, r, dto.ErrorResponse{
						Error: dto.ErrorObject{
							Code:   apperrors.CodeInvalidRequest,
							Title:  "Invalid request body",
							Detail: "Request validation failed",
						},
					})
					return
				}

				// Build user-friendly error message from validation errors
				var errorMessages []string
				for _, fieldErr := range validationErrors {
					fieldName := fieldErr.Field()
					if fieldName == "" {
						fieldName = fieldErr.StructField()
					}
					errorMessages = append(errorMessages, fmt.Sprintf("%s: %s", fieldName, fieldErr.Error()))
				}

				logger.Warn("Request validation failed",
					zap.Strings("errors", errorMessages),
					zap.String("method", r.Method),
					zap.String("path", r.URL.Path),
				)

				render.Status(r, http.StatusBadRequest)
				render.JSON(w, r, dto.ErrorResponse{
					Error: dto.ErrorObject{
						Code:   apperrors.CodeInvalidRequest,
						Title:  "Invalid request body",
						Detail: strings.Join(errorMessages, "; "),
					},
				})
				return
			}

			if v, ok := any(&bodyDTO).(Validator); ok {
				if err := v.Validate(); err != nil {
					logger.Warn("Request validation failed",
						zap.Error(err),
						zap.String("method", r.Method),
						zap.String("path", r.URL.Path),
					)

					render.Status(r, http.StatusBadRequest)
					render.JSON(w, r, dto.ErrorResponse{
						Error: dto.ErrorObject{
							Code:   apperrors.CodeInvalidRequest,
							Title:  "Invalid request body",
							Detail: err.Error(),
						},
					})
					return
				}
			}

			ctx := context.WithValue(r.Context(), reqBodyKey, bodyDTO)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

/*
GetRequestBodyFromContext extracts the validated request body from the request context.

This provides a type-safe way to retrieve the request body without exposing
the context key implementation details.

This function panics if the request body is not found in context, which
indicates a programming error (e.g., handler called without RequestValidator
middleware, or wrong DTO type specified). The panic will be recovered by the
Recoverer middleware and logged appropriately.
*/
func GetRequestBodyFromContext[T any](ctx context.Context) T {
	val := ctx.Value(reqBodyKey)
	if val == nil {
		panic("request body not found in context: RequestValidator middleware must be applied before this handler")
	}

	dto, ok := val.(T)
	if !ok {
		panic("request body type mismatch: ensure RequestValidator middleware uses the same DTO type")
	}

	return dto
}
