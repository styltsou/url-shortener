package handlers

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/google/uuid"
	"github.com/styltsou/url-shortener/server/pkg/db"
	"github.com/styltsou/url-shortener/server/pkg/dto"
	apperrors "github.com/styltsou/url-shortener/server/pkg/errors"
	"github.com/styltsou/url-shortener/server/pkg/logger"
	mw "github.com/styltsou/url-shortener/server/pkg/middleware"
	"github.com/styltsou/url-shortener/server/pkg/service"
	"go.uber.org/zap"
)

// All handlers follow established patterns:
// - Use handleError() for consistent error logging and HTTP response mapping
// - Log errors with appropriate levels (Warn for client errors, Error for server errors)
// - Include context (method, path, user_id, etc.) in log entries
// - Use structured logging with zap fields

// LinkServiceInterface defines the service methods needed by LinkHandler
type LinkService interface {
	GetOriginalURL(ctx context.Context, code string) (db.GetLinkForRedirectRow, error)
	CreateShortLink(ctx context.Context, userID string, originalURL string, customShortcode *string, expiresAt *time.Time) (db.TryCreateLinkRow, error)
	ListAllLinks(ctx context.Context, userID string, isActive *bool, tagIDs []uuid.UUID, page, limit int) (*service.ListLinksResult, error)
	GetLinkByShortcode(ctx context.Context, userID string, shortcode string) (db.GetLinkByShortcodeAndUserRow, error)
	UpdateLink(ctx context.Context, userID string, id uuid.UUID, shortcode *string, isActive *bool, expiresAt *time.Time) (db.UpdateLinkRow, error)
	DeleteLink(ctx context.Context, userID string, id uuid.UUID) (db.DeleteLinkRow, error)
	AddTagsToLink(ctx context.Context, userID string, linkID uuid.UUID, tagIDs []uuid.UUID) (db.GetLinkByIdAndUserWithTagsRow, error)
	RemoveTagsFromLink(ctx context.Context, userID string, linkID uuid.UUID, tagIDs []uuid.UUID) (db.GetLinkByIdAndUserWithTagsRow, error)
}

type LinkHandler struct {
	LinkService LinkService
	logger      logger.Logger
}

func NewLinkHandler(linkService LinkService, logger logger.Logger) *LinkHandler {
	return &LinkHandler{
		LinkService: linkService,
		logger:      logger,
	}
}

// Public redirect: GET /{shortcode}
func (h *LinkHandler) Redirect(w http.ResponseWriter, r *http.Request) {
	shortcode := chi.URLParam(r, "shortcode")

	link, err := h.LinkService.GetOriginalURL(r.Context(), shortcode)
	if err != nil {
		h.logger.Warn("Link not found for redirect",
			zap.Error(err),
			zap.String("shortcode", shortcode),
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.String("remote_addr", r.RemoteAddr),
		)
		render.Status(r, http.StatusNotFound)
		render.HTML(w, r, `<!DOCTYPE html>
<html>
	<head><title>Link Not Found</title></head>
	<body>
		<h1>404 - Link Not Found</h1>
		<p>This link may have expired or been deleted.</p>
	</body>
</html>`)
		return
	}

	http.Redirect(w, r, link.OriginalUrl, http.StatusFound)
}

// Create link: POST /api/v1/links
func (h *LinkHandler) CreateLink(w http.ResponseWriter, r *http.Request) {
	reqBody := mw.GetRequestBodyFromContext[dto.CreateLink](r.Context())
	userID := mw.GetUserIDFromContext(r.Context())

	createdLink, err := h.LinkService.CreateShortLink(
		r.Context(),
		userID,
		reqBody.URL,
		reqBody.Shortcode,
		reqBody.ExpiresAt,
	)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	h.logger.Info("Short link created successfully",
		zap.String("user_id", userID),
		zap.String("link_id", createdLink.ID.String()),
		zap.String("short_code", createdLink.Shortcode),
		zap.String("original_url", createdLink.OriginalUrl),
	)

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, &dto.SuccessResponse[db.TryCreateLinkRow]{
		Data: createdLink,
	})
}

// List links: GET /api/v1/links?tags=id1,id2&status=active|inactive|all
func (h *LinkHandler) ListLinks(w http.ResponseWriter, r *http.Request) {
	userID := mw.GetUserIDFromContext(r.Context())

	// Parse query parameters
	var isActive *bool
	var tagIDs []uuid.UUID

	// Parse status filter: ?status=active|inactive|all
	status := r.URL.Query().Get("status")
	if status != "" && status != "all" {
		switch status {
		case "active":
			val := true
			isActive = &val
		case "inactive":
			val := false
			isActive = &val
		}
	}

	// Parse tag IDs: ?tags=id1,id2,id3
	tagsParam := r.URL.Query().Get("tags")
	if tagsParam != "" {
		tagStrs := strings.Split(tagsParam, ",")
		for _, tagStr := range tagStrs {
			tagStr = strings.TrimSpace(tagStr)
			if tagStr == "" {
				continue
			}
			tagID, err := uuid.Parse(tagStr)
			if err != nil {
				h.logger.Warn("Invalid tag ID in query parameter",
					zap.String("tag_id", tagStr),
					zap.Error(err),
				)
				continue
			}
			tagIDs = append(tagIDs, tagID)
		}
	}

	// Parse pagination parameters: ?page=1&limit=5
	page := 1
	limit := 5
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	h.logger.Info("Listing user links",
		zap.String("user_id", userID),
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
		zap.Any("is_active", isActive),
		zap.Any("tag_ids", tagIDs),
		zap.Int("page", page),
		zap.Int("limit", limit),
	)

	result, err := h.LinkService.ListAllLinks(r.Context(), userID, isActive, tagIDs, page, limit)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	// Ensure we always return an empty array (not null) when there are no links
	if result.Links == nil {
		result.Links = []db.ListUserLinksRow{}
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, &dto.SuccessResponse[[]db.ListUserLinksRow]{
		Data: result.Links,
		Pagination: &dto.PaginationMeta{
			Page:       result.Page,
			Limit:      result.Limit,
			Total:      result.Total,
			TotalPages: result.TotalPages,
		},
	})
}

// Get link by shortcode: GET /api/v1/links/{shortcode}
func (h *LinkHandler) GetLink(w http.ResponseWriter, r *http.Request) {
	userID := mw.GetUserIDFromContext(r.Context())
	shortcode := chi.URLParam(r, "shortcode")

	link, err := h.LinkService.GetLinkByShortcode(r.Context(), userID, shortcode)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, &dto.SuccessResponse[db.GetLinkByShortcodeAndUserRow]{
		Data: link,
	})
}

// Update link (PATCH code/expiry): PATCH /api/v1/links/{id}
func (h *LinkHandler) UpdateLink(w http.ResponseWriter, r *http.Request) {
	userID := mw.GetUserIDFromContext(r.Context())

	linkID, uuidErr := uuid.Parse(chi.URLParam(r, "id"))
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

	body := mw.GetRequestBodyFromContext[dto.UpdateLink](r.Context())

	updatedLink, err := h.LinkService.UpdateLink(
		r.Context(),
		userID,
		linkID,
		body.Shortcode,
		body.IsActive,
		body.ExpiresAt,
	)

	if err != nil {
		// handleError logs the error and maps it to appropriate HTTP response
		h.handleError(w, r, err)
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, dto.SuccessResponse[db.UpdateLinkRow]{
		Data: updatedLink,
	})
}

// Delete link by ID: DELETE /api/v1/links/{id}
func (h *LinkHandler) DeleteLink(w http.ResponseWriter, r *http.Request) {
	userID := mw.GetUserIDFromContext(r.Context())

	linkID, uuidErr := uuid.Parse(chi.URLParam(r, "id"))
	if uuidErr != nil {
		h.logger.Warn("Invalid ID format",
			zap.Error(uuidErr),
			zap.String("provided_id", chi.URLParam(r, "id")), // Log for debugging
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

	// Here call the actuall delete service, handle any error, return the deleted entity
	deletedLink, err := h.LinkService.DeleteLink(r.Context(), userID, linkID)

	if err != nil {
		h.handleError(w, r, err)
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, &dto.SuccessResponse[db.DeleteLinkRow]{
		Data: deletedLink,
	})
}

// AddTagsToLink: POST /api/v1/links/{id}/tags
func (h *LinkHandler) AddTagsToLink(w http.ResponseWriter, r *http.Request) {
	userID := mw.GetUserIDFromContext(r.Context())

	linkID, uuidErr := uuid.Parse(chi.URLParam(r, "id"))
	if uuidErr != nil {
		h.logger.Warn("Invalid link ID format",
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
				Detail: "Link ID must be a valid UUID format",
			},
		})
		return
	}

	reqBody := mw.GetRequestBodyFromContext[dto.AddTagsToLink](r.Context())

	updatedLink, err := h.LinkService.AddTagsToLink(r.Context(), userID, linkID, reqBody.TagIDs)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, &dto.SuccessResponse[db.GetLinkByIdAndUserWithTagsRow]{
		Data: updatedLink,
	})
}

// RemoveTagsFromLink: DELETE /api/v1/links/{id}/tags
func (h *LinkHandler) RemoveTagsFromLink(w http.ResponseWriter, r *http.Request) {
	userID := mw.GetUserIDFromContext(r.Context())

	linkID, uuidErr := uuid.Parse(chi.URLParam(r, "id"))
	if uuidErr != nil {
		h.logger.Warn("Invalid link ID format",
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
				Detail: "Link ID must be a valid UUID format",
			},
		})
		return
	}

	reqBody := mw.GetRequestBodyFromContext[dto.RemoveTagsFromLink](r.Context())

	updatedLink, err := h.LinkService.RemoveTagsFromLink(r.Context(), userID, linkID, reqBody.TagIDs)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, &dto.SuccessResponse[db.GetLinkByIdAndUserWithTagsRow]{
		Data: updatedLink,
	})
}

// handleError maps errors to HTTP responses and writes them directly
func (h *LinkHandler) handleError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, apperrors.LinkNotFound):
		h.logger.Warn("Link not found",
			zap.Error(err),
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
		)
		render.Status(r, http.StatusNotFound)
		render.JSON(w, r, dto.ErrorResponse{
			Error: dto.ErrorObject{
				Code:   apperrors.CodeLinkNotFound,
				Title:  apperrors.LinkNotFound.Error(),
				Detail: "Unable to find link with shortcode",
			},
		})

	case errors.Is(err, apperrors.InvalidURL):
		h.logger.Warn("Invalid URL",
			zap.Error(err),
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
		)
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, dto.ErrorResponse{
			Error: dto.ErrorObject{
				Code:   apperrors.CodeInvalidURL,
				Title:  apperrors.InvalidURL.Error(),
				Detail: "",
			},
		})

	case errors.Is(err, sql.ErrNoRows):
		h.logger.Warn("Resource not found",
			zap.Error(err),
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
		)
		render.Status(r, http.StatusNotFound)
		render.JSON(w, r, dto.ErrorResponse{
			Error: dto.ErrorObject{
				Code:   apperrors.CodeLinkNotFound,
				Title:  "Resource not found",
				Detail: "",
			},
		})

	case errors.Is(err, apperrors.LinkShortcodeTaken):
		h.logger.Warn("Shortcode already taken",
			zap.Error(err),
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
		)
		render.Status(r, http.StatusConflict) // 409 Conflict
		render.JSON(w, r, dto.ErrorResponse{
			Error: dto.ErrorObject{
				Code:   apperrors.CodeCodeTaken,
				Title:  apperrors.LinkShortcodeTaken.Error(),
				Detail: "The provided shortcode is already in use",
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
