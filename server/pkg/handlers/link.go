package handlers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/google/uuid"
	"github.com/skip2/go-qrcode"
	"github.com/styltsou/url-shortener/server/pkg/analytics"
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
	CreateShortLinkWithTags(ctx context.Context, userID string, originalURL string, customShortcode *string, expiresAt *time.Time, tagIDs []uuid.UUID) (db.GetLinkByIdAndUserWithTagsRow, error)
	ListAllLinks(ctx context.Context, userID string, isActive *bool, tagIDs []uuid.UUID, page, limit int) (*service.ListLinksResult, error)
	GetLinkByShortcode(ctx context.Context, userID string, shortcode string) (db.GetLinkByShortcodeAndUserRow, error)
	UpdateLink(ctx context.Context, userID string, id uuid.UUID, shortcode *string, isActive *bool, expiresAtSet bool, expiresAt *time.Time) (db.UpdateLinkRow, error)
	DeleteLink(ctx context.Context, userID string, id uuid.UUID) (db.DeleteLinkRow, error)
	AddTagsToLink(ctx context.Context, userID string, linkID uuid.UUID, tagIDs []uuid.UUID) (db.GetLinkByIdAndUserWithTagsRow, error)
	RemoveTagsFromLink(ctx context.Context, userID string, linkID uuid.UUID, tagIDs []uuid.UUID) (db.GetLinkByIdAndUserWithTagsRow, error)
	RecordClick(ctx context.Context, linkID uuid.UUID, ip, userAgent, referrer string)
	GetLinkAnalytics(ctx context.Context, userID string, shortcode string) (*analytics.LinkAnalytics, error)
	GetDashboardStats(ctx context.Context, userID string) (*service.DashboardStats, error)
}

type LinkHandler struct {
	LinkService LinkService
	baseURL     string
	logger      logger.Logger
}

func NewLinkHandler(linkService LinkService, baseURL string, logger logger.Logger) *LinkHandler {
	return &LinkHandler{
		LinkService: linkService,
		baseURL:     baseURL,
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

	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	h.LinkService.RecordClick(r.Context(), link.ID, ip, r.UserAgent(), r.Referer())

	http.Redirect(w, r, link.OriginalUrl, http.StatusFound)
}

// Create link: POST /api/v1/links
func (h *LinkHandler) CreateLink(w http.ResponseWriter, r *http.Request) {
	reqBody := mw.GetRequestBodyFromContext[dto.CreateLink](r.Context())
	userID := mw.GetUserIDFromContext(r.Context())

	createdLink, err := h.LinkService.CreateShortLinkWithTags(
		r.Context(),
		userID,
		reqBody.URL,
		reqBody.Shortcode,
		reqBody.ExpiresAt,
		reqBody.TagIDs,
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
	response, err := dto.LinkFromWithTagsRow(createdLink)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	render.JSON(w, r, &dto.SuccessResponse[dto.LinkResponse]{
		Data: response,
	})
}

// List links: GET /api/v1/links?tags=id1,id2&status=active|inactive|all
func (h *LinkHandler) ListLinks(w http.ResponseWriter, r *http.Request) {
	userID := mw.GetUserIDFromContext(r.Context())

	query, parseErr := parseListLinksQuery(r)
	if parseErr != nil {
		h.logger.Warn("Invalid list links query",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.String("detail", parseErr.Error.Detail),
		)
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, parseErr)
		return
	}

	h.logger.Info("Listing user links",
		zap.String("user_id", userID),
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
		zap.Any("is_active", query.IsActive),
		zap.Any("tag_ids", query.TagIDs),
		zap.Int("page", query.Page),
		zap.Int("limit", query.Limit),
	)

	result, err := h.LinkService.ListAllLinks(r.Context(), userID, query.IsActive, query.TagIDs, query.Page, query.Limit)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	links, err := dto.LinksFromListRows(result.Links)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, &dto.SuccessResponse[[]dto.LinkResponse]{
		Data: links,
		Pagination: &dto.PaginationMeta{
			Page:       result.Page,
			Limit:      result.Limit,
			Total:      result.Total,
			TotalPages: result.TotalPages,
		},
	})
}

type listLinksQuery struct {
	IsActive *bool
	TagIDs   []uuid.UUID
	Page     int
	Limit    int
}

func parseListLinksQuery(r *http.Request) (listLinksQuery, *dto.ErrorResponse) {
	query := listLinksQuery{Page: 1, Limit: 5}
	values := r.URL.Query()

	status := values.Get("status")
	switch status {
	case "", "all":
	case "active":
		active := true
		query.IsActive = &active
	case "inactive":
		active := false
		query.IsActive = &active
	default:
		return listLinksQuery{}, invalidQuery("invalid_request", "Invalid status filter", "status must be one of: active, inactive, all")
	}

	tagsParam := values.Get("tags")
	if tagsParam != "" {
		for _, tagStr := range strings.Split(tagsParam, ",") {
			tagStr = strings.TrimSpace(tagStr)
			if tagStr == "" {
				continue
			}
			tagID, err := uuid.Parse(tagStr)
			if err != nil {
				return listLinksQuery{}, invalidQuery(string(apperrors.CodeInvalidID), "Invalid tag ID format", fmt.Sprintf("Tag ID '%s' is not a valid UUID", tagStr))
			}
			query.TagIDs = append(query.TagIDs, tagID)
		}
	}

	if pageStr := values.Get("page"); pageStr != "" {
		page, err := strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			return listLinksQuery{}, invalidQuery("invalid_request", "Invalid page", "page must be a positive integer")
		}
		query.Page = page
	}
	if limitStr := values.Get("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 1 {
			return listLinksQuery{}, invalidQuery("invalid_request", "Invalid limit", "limit must be a positive integer")
		}
		query.Limit = limit
	}

	return query, nil
}

func invalidQuery(code string, title string, detail string) *dto.ErrorResponse {
	return &dto.ErrorResponse{
		Error: dto.ErrorObject{
			Code:   apperrors.ErrorCode(code),
			Title:  title,
			Detail: detail,
		},
	}
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

	response, err := dto.LinkFromShortcodeRow(link)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, &dto.SuccessResponse[dto.LinkResponse]{Data: response})
}

// Update link (PATCH code/expiry): PATCH /api/v1/links/{id}
func (h *LinkHandler) UpdateLink(w http.ResponseWriter, r *http.Request) {
	userID := mw.GetUserIDFromContext(r.Context())

	linkID, ok := parseUUIDParam(w, r, h.logger)
	if !ok {
		return
	}

	body := mw.GetRequestBodyFromContext[dto.UpdateLink](r.Context())

	updatedLink, err := h.LinkService.UpdateLink(
		r.Context(),
		userID,
		linkID,
		body.Shortcode,
		body.IsActive,
		body.ExpiresAt.Set,
		body.ExpiresAt.Value,
	)

	if err != nil {
		// handleError logs the error and maps it to appropriate HTTP response
		h.handleError(w, r, err)
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, dto.SuccessResponse[dto.LinkResponse]{
		Data: dto.LinkFromUpdateRow(updatedLink),
	})
}

// Delete link by ID: DELETE /api/v1/links/{id}
func (h *LinkHandler) DeleteLink(w http.ResponseWriter, r *http.Request) {
	userID := mw.GetUserIDFromContext(r.Context())

	linkID, ok := parseUUIDParam(w, r, h.logger)
	if !ok {
		return
	}

	// Here call the actuall delete service, handle any error, return the deleted entity
	deletedLink, err := h.LinkService.DeleteLink(r.Context(), userID, linkID)

	if err != nil {
		h.handleError(w, r, err)
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, &dto.SuccessResponse[dto.LinkResponse]{
		Data: dto.LinkFromDeleteRow(deletedLink),
	})
}

// AddTagsToLink: POST /api/v1/links/{id}/tags
func (h *LinkHandler) AddTagsToLink(w http.ResponseWriter, r *http.Request) {
	userID := mw.GetUserIDFromContext(r.Context())

	linkID, ok := parseUUIDParam(w, r, h.logger)
	if !ok {
		return
	}

	reqBody := mw.GetRequestBodyFromContext[dto.AddTagsToLink](r.Context())

	updatedLink, err := h.LinkService.AddTagsToLink(r.Context(), userID, linkID, reqBody.TagIDs)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	render.Status(r, http.StatusOK)
	response, err := dto.LinkFromWithTagsRow(updatedLink)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	render.JSON(w, r, &dto.SuccessResponse[dto.LinkResponse]{
		Data: response,
	})
}

// RemoveTagsFromLink: DELETE /api/v1/links/{id}/tags
func (h *LinkHandler) RemoveTagsFromLink(w http.ResponseWriter, r *http.Request) {
	userID := mw.GetUserIDFromContext(r.Context())

	linkID, ok := parseUUIDParam(w, r, h.logger)
	if !ok {
		return
	}

	reqBody := mw.GetRequestBodyFromContext[dto.RemoveTagsFromLink](r.Context())

	updatedLink, err := h.LinkService.RemoveTagsFromLink(r.Context(), userID, linkID, reqBody.TagIDs)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	render.Status(r, http.StatusOK)
	response, err := dto.LinkFromWithTagsRow(updatedLink)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	render.JSON(w, r, &dto.SuccessResponse[dto.LinkResponse]{
		Data: response,
	})
}

// GetQRCode: GET /api/v1/links/{shortcode}/qrcode
func (h *LinkHandler) GetQRCode(w http.ResponseWriter, r *http.Request) {
	shortcode := chi.URLParam(r, "shortcode")
	sizeStr := r.URL.Query().Get("size")
	size := 256
	if sizeStr != "" {
		if s, err := strconv.Atoi(sizeStr); err == nil && s >= 64 && s <= 1024 {
			size = s
		}
	}

	shortURL := fmt.Sprintf("%s/%s", h.baseURL, shortcode)

	png, err := qrcode.Encode(shortURL, qrcode.Medium, size)
	if err != nil {
		h.logger.Error("Failed to generate QR code",
			zap.Error(err),
			zap.String("shortcode", shortcode),
		)
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, dto.ErrorResponse{
			Error: dto.ErrorObject{
				Code:   apperrors.CodeInternalError,
				Title:  apperrors.InternalError.Error(),
				Detail: "Failed to generate QR code",
			},
		})
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`inline; filename="qr-%s.png"`, shortcode))
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(png); err != nil {
		h.logger.Error("Failed to write QR code response",
			zap.Error(err),
			zap.String("shortcode", shortcode),
		)
	}
}

// GetDashboard: GET /api/v1/dashboard
func (h *LinkHandler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	userID := mw.GetUserIDFromContext(r.Context())

	stats, err := h.LinkService.GetDashboardStats(r.Context(), userID)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, dto.SuccessResponse[*service.DashboardStats]{
		Data: stats,
	})
}

// GetLinkAnalytics: GET /api/v1/links/{shortcode}/analytics
func (h *LinkHandler) GetLinkAnalytics(w http.ResponseWriter, r *http.Request) {
	userID := mw.GetUserIDFromContext(r.Context())
	shortcode := chi.URLParam(r, "shortcode")

	analyticsData, err := h.LinkService.GetLinkAnalytics(r.Context(), userID, shortcode)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, dto.SuccessResponse[*analytics.LinkAnalytics]{
		Data: analyticsData,
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
				Detail: "One or more tags were not found",
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
