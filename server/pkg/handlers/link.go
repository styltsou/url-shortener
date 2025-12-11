package handlers

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"time"

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

// TODO: Generally walk through every implemented handler and make sure
// they follow our newly established patterns
// Also make use of the logger interface as params for logger
// It generally needs refactory especially when it comes to logging and error handling

// LinkServiceInterface defines the service methods needed by LinkHandler
type LinkService interface {
	GetOriginalURL(ctx context.Context, code string) (db.GetLinkForRedirectRow, error)
	CreateShortLink(ctx context.Context, userID string, originalURL string) (db.TryCreateLinkRow, error)
	ListAllLinks(ctx context.Context, userID string) ([]db.ListUserLinksRow, error)
	GetLinkByShortcode(ctx context.Context, userID string, shortcode string) (db.GetLinkByShortcodeAndUserRow, error)
	UpdateLink(ctx context.Context, userID string, id uuid.UUID, shortcode *string, isActive *bool, expiresAt *time.Time) (db.UpdateLinkRow, error)
	DeleteLink(ctx context.Context, userID string, id uuid.UUID) (db.DeleteLinkRow, error)
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
		// TODO: Log error
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

	createdLink, err := h.LinkService.CreateShortLink(r.Context(), userID, reqBody.URL)
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

// List links: GET /api/v1/links
func (h *LinkHandler) ListLinks(w http.ResponseWriter, r *http.Request) {
	userID := mw.GetUserIDFromContext(r.Context())

	// TODO: Is this redundant in production?
	h.logger.Info("Listing user links",
		zap.String("user_id", userID),
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
	)

	links, err := h.LinkService.ListAllLinks(r.Context(), userID)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	// Ensure we always return an empty array (not null) when there are no links
	if links == nil {
		links = []db.ListUserLinksRow{}
	}

	// TODO: I think that maybe this is too redundant.
	// // Empty slice is a valid response - user simply has no links yet
	// // This is not an error condition
	// if len(links) == 0 {
	// 	h.logger.Info("User has no links",
	// 		zap.String("user_id", userID),
	// 	)
	// } else {
	// 	h.logger.Info("User links retrieved successfully",
	// 		zap.String("user_id", userID),
	// 		zap.Int("link_count", len(links)),
	// 	)
	// }

	render.Status(r, http.StatusOK)
	render.JSON(w, r, &dto.SuccessResponse[[]db.ListUserLinksRow]{
		Data: links,
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
		// TODO: Logging and also check if handler handles all types of possible errors
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
