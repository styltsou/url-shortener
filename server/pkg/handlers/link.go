package handlers

import (
	"context"
	"database/sql"
	"errors"
	"net/http"

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
	CreateShortLink(ctx context.Context, userID string, originalURL string) (db.Link, error)
	ListAllLinks(ctx context.Context, userID string) ([]db.Link, error)
	GetLinkByID(ctx context.Context, id uuid.UUID, userID string) (db.Link, error)
	GetOriginalURL(ctx context.Context, code string) (db.GetLinkForRedirectRow, error)
	DeleteLink(ctx context.Context, id uuid.UUID, userID string) error
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

// Public redirect: GET /{code}
func (h *LinkHandler) Redirect(w http.ResponseWriter, r *http.Request) {}

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
	render.JSON(w, r, &dto.SuccessResponse[db.Link]{
		Data:    createdLink,
		Message: "Short Link created successfully",
	})
}

// List links: GET /api/v1/links
func (h *LinkHandler) ListLinks(w http.ResponseWriter, r *http.Request) {
	userID := mw.GetUserIDFromContext(r.Context())
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
		links = []db.Link{}
	}

	// Empty slice is a valid response - user simply has no links yet
	// This is not an error condition
	if len(links) == 0 {
		h.logger.Info("User has no links",
			zap.String("user_id", userID),
		)
	} else {
		h.logger.Info("User links retrieved successfully",
			zap.String("user_id", userID),
			zap.Int("link_count", len(links)),
		)
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, &dto.SuccessResponse[[]db.Link]{
		Data:    links, // Will be empty slice [] if no links exist
		Message: "User links retrieved successfully",
	})
}

// Get link by ID: GET /api/v1/links/{id}
func (h *LinkHandler) GetLink(w http.ResponseWriter, r *http.Request) {}

// Update link (PATCH code/expiry): PATCH /api/v1/links/{id}
func (h *LinkHandler) UpdateLink(w http.ResponseWriter, r *http.Request) {}

// Delete link by ID: DELETE /api/v1/links/{id}
func (h *LinkHandler) DeleteLink(w http.ResponseWriter, r *http.Request) {}

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
