package middleware

import (
	"context"
	"net/http"

	"github.com/clerk/clerk-sdk-go/v2"
	clerkhttp "github.com/clerk/clerk-sdk-go/v2/http"
	"github.com/go-chi/render"
	"github.com/styltsou/url-shortener/server/pkg/dto"
	apperrors "github.com/styltsou/url-shortener/server/pkg/errors"
	"github.com/styltsou/url-shortener/server/pkg/logger"
	"go.uber.org/zap"
)

type contextKey string

const userIDKey contextKey = "user_id"

// authFailureHandler returns an HTTP handler that writes authentication failure
// responses using our API error schema format.
func authFailureHandler(log logger.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Warn("Authentication failed",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
		)

		render.Status(r, http.StatusUnauthorized)
		render.JSON(w, r, dto.ErrorResponse{
			Error: dto.ErrorObject{
				Code:   apperrors.CodeAuthRequired,
				Title:  apperrors.AuthRequired.Error(),
				Detail: "You need to be authenticated to perform this action",
			},
		})
	})
}

/*
RequireAuth is a middleware that:
1. Requires and validates the Authorization header (via Clerk's RequireHeaderAuthorization)
2. Extracts session claims from context (added by Clerk)
3. Adds userID to context for handlers to use
*/
func RequireAuth(log logger.Logger) func(http.Handler) http.Handler {
	clerkAuth := clerkhttp.RequireHeaderAuthorization(
		clerkhttp.AuthorizationFailureHandler(authFailureHandler(log)),
	)

	return func(next http.Handler) http.Handler {
		return clerkAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := clerk.SessionClaimsFromContext(r.Context())

			if !ok || claims == nil {
				log.Error("Session claims missing after successful authentication",
					zap.String("method", r.Method),
					zap.String("path", r.URL.Path),
				)

				render.Status(r, http.StatusInternalServerError)
				render.JSON(w, r, dto.ErrorResponse{
					Error: dto.ErrorObject{
						Code:   apperrors.CodeInternalError,
						Title:  apperrors.InternalError.Error(),
						Detail: "An internal error occurred while processing your request",
					},
				})
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, claims.Subject)
			next.ServeHTTP(w, r.WithContext(ctx))
		}))
	}
}

// GetUserID extracts the user ID from the request context.
func GetUserIDFromContext(ctx context.Context) string {
	userID, ok := ctx.Value(userIDKey).(string)

	if !ok {
		panic("user ID not found in context: make sure that the handler is authenticated")
	}

	return userID
}

/*
WithUserID adds the user ID to the context.

This is primarily used for testing, but can also be used when you need
to manually set the user ID in context (e.g., in tests or when bypassing auth).
*/
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}
