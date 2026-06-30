package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/google/uuid"
	"github.com/styltsou/url-shortener/server/pkg/dto"
	apperrors "github.com/styltsou/url-shortener/server/pkg/errors"
	"github.com/styltsou/url-shortener/server/pkg/logger"
	"go.uber.org/zap"
)

func parseUUIDParam(w http.ResponseWriter, r *http.Request, log logger.Logger) (uuid.UUID, bool) {
	idStr := chi.URLParam(r, "id")
	parsed, err := uuid.Parse(idStr)
	if err != nil {
		log.Warn("Invalid ID format",
			zap.Error(err),
			zap.String("provided_id", idStr),
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
		)

		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, dto.ErrorResponse{
			Error: dto.ErrorObject{
				Code:   apperrors.CodeInvalidID,
				Title:  "Invalid ID format",
				Detail: "ID must be a valid UUID format",
			},
		})
		return uuid.UUID{}, false
	}
	return parsed, true
}


