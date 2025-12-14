package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/google/uuid"
	"github.com/styltsou/url-shortener/server/pkg/db"
	"github.com/styltsou/url-shortener/server/pkg/dto"
	apperrors "github.com/styltsou/url-shortener/server/pkg/errors"
	"github.com/styltsou/url-shortener/server/pkg/logger"
	mw "github.com/styltsou/url-shortener/server/pkg/middleware"
	"go.uber.org/zap"
)

// TagService defines the service methods needed by TagHandler
type TagService interface {
	ListAllTags(ctx context.Context, userID string) ([]db.ListUserTagsRow, error)
	CreateTag(ctx context.Context, userID string, name string) (db.CreateTagRow, error)
	UpdateTag(ctx context.Context, userID string, tagID uuid.UUID, name string) (db.UpdateTagRow, error)
	DeleteTag(ctx context.Context, userID string, tagID uuid.UUID) (db.DeleteTagRow, error)
	DeleteTags(ctx context.Context, userID string, tagIDs []uuid.UUID) ([]db.DeleteTagsRow, error)
}

type TagHandler struct {
	TagService TagService
	logger     logger.Logger
}

func NewTagHandler(tagService TagService, logger logger.Logger) *TagHandler {
	return &TagHandler{
		TagService: tagService,
		logger:     logger,
	}
}

// ListTags: GET /api/v1/tags
func (h *TagHandler) ListTags(w http.ResponseWriter, r *http.Request) {
	userID := mw.GetUserIDFromContext(r.Context())

	tags, err := h.TagService.ListAllTags(r.Context(), userID)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, &dto.SuccessResponse[[]db.ListUserTagsRow]{
		Data: tags,
	})
}

// CreateTag: POST /api/v1/tags
func (h *TagHandler) CreateTag(w http.ResponseWriter, r *http.Request) {
	reqBody := mw.GetRequestBodyFromContext[dto.CreateTag](r.Context())
	userID := mw.GetUserIDFromContext(r.Context())

	createdTag, err := h.TagService.CreateTag(r.Context(), userID, reqBody.Name)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	h.logger.Info("Tag created successfully",
		zap.String("user_id", userID),
		zap.String("tag_id", createdTag.ID.String()),
		zap.String("tag_name", createdTag.Name),
	)

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, &dto.SuccessResponse[db.CreateTagRow]{
		Data: createdTag,
	})
}

// UpdateTag: PATCH /api/v1/tags/{id}
func (h *TagHandler) UpdateTag(w http.ResponseWriter, r *http.Request) {
	userID := mw.GetUserIDFromContext(r.Context())

	tagID, uuidErr := uuid.Parse(chi.URLParam(r, "id"))
	if uuidErr != nil {
		h.logger.Warn("Invalid ID format",
			zap.Error(uuidErr),
			zap.String("provided_id", chi.URLParam(r, "id")),
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
		return
	}

	reqBody := mw.GetRequestBodyFromContext[dto.UpdateTag](r.Context())

	updatedTag, err := h.TagService.UpdateTag(r.Context(), userID, tagID, reqBody.Name)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, &dto.SuccessResponse[db.UpdateTagRow]{
		Data: updatedTag,
	})
}

// DeleteTag: DELETE /api/v1/tags/{id}
func (h *TagHandler) DeleteTag(w http.ResponseWriter, r *http.Request) {
	userID := mw.GetUserIDFromContext(r.Context())

	tagID, uuidErr := uuid.Parse(chi.URLParam(r, "id"))
	if uuidErr != nil {
		h.logger.Warn("Invalid ID format",
			zap.Error(uuidErr),
			zap.String("provided_id", chi.URLParam(r, "id")),
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
		return
	}

	deletedTag, err := h.TagService.DeleteTag(r.Context(), userID, tagID)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, &dto.SuccessResponse[db.DeleteTagRow]{
		Data: deletedTag,
	})
}

// DeleteTags: DELETE /api/v1/tags/bulk
func (h *TagHandler) DeleteTags(w http.ResponseWriter, r *http.Request) {
	userID := mw.GetUserIDFromContext(r.Context())

	reqBody := mw.GetRequestBodyFromContext[dto.DeleteTags](r.Context())

	deletedTags, err := h.TagService.DeleteTags(r.Context(), userID, reqBody.TagIDs)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, &dto.SuccessResponse[[]db.DeleteTagsRow]{
		Data: deletedTags,
	})
}

// handleError maps errors to HTTP responses and writes them directly
func (h *TagHandler) handleError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, apperrors.TagNotFound):
		h.logger.Warn("Tag not found",
			zap.Error(err),
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
		)
		render.Status(r, http.StatusNotFound)
		render.JSON(w, r, dto.ErrorResponse{
			Error: dto.ErrorObject{
				Code:   apperrors.CodeTagNotFound,
				Title:  apperrors.TagNotFound.Error(),
				Detail: "Unable to find tag with the provided ID",
			},
		})

	case errors.Is(err, apperrors.TagNameTaken):
		h.logger.Warn("Tag name already taken",
			zap.Error(err),
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
		)
		render.Status(r, http.StatusConflict) // 409 Conflict
		render.JSON(w, r, dto.ErrorResponse{
			Error: dto.ErrorObject{
				Code:   apperrors.CodeTagNameTaken,
				Title:  apperrors.TagNameTaken.Error(),
				Detail: "A tag with this name already exists",
			},
		})

	default:
		h.logger.Error("Internal server error",
			zap.Error(err),
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
		)
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, dto.ErrorResponse{
			Error: dto.ErrorObject{
				Code:   apperrors.CodeInternalError,
				Title:  apperrors.InternalError.Error(),
				Detail: "",
			},
		})
	}
}
