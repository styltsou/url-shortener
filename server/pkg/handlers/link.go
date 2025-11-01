package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/go-chi/render"
	"github.com/styltsou/url-shortener/server/pkg/db"
	"github.com/styltsou/url-shortener/server/pkg/dto"
	"github.com/styltsou/url-shortener/server/pkg/service"
)

type LinkHandler struct {
	LinkService *service.LinkService
}

func NewLinkHandler(linkService *service.LinkService) *LinkHandler {
	return &LinkHandler{linkService}
}

// Public redirect: GET /{code}
func (h *LinkHandler) Redirect(w http.ResponseWriter, r *http.Request) {}

// Create link: POST /api/v1/links
func (h *LinkHandler) CreateLink(w http.ResponseWriter, r *http.Request) {
	var reqBody dto.CreateLinkRequest

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, &dto.ErrorResponse{
			Error: dto.ErrorDetail{
				Code:    dto.ErrorBadRequest,
				Message: "Unable to create a new shortened URL",
			},
		})
	}

	claims, _ := clerk.SessionClaimsFromContext(r.Context())

	createdLlink, err := h.LinkService.CreateShortLink(r.Context(), claims.Subject, reqBody.URL)

	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, &dto.ErrorResponse{
			Error: dto.ErrorDetail{
				Code:    dto.ErrorInternal,
				Message: "Unable to create a new shortened URL",
			},
		})
	}

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, &dto.SuccessReponse[db.Link]{
		Data:    createdLlink,
		Message: "Short Link created successfully",
	})
}

// List links: GET /api/v1/links
func (h *LinkHandler) ListLinks(w http.ResponseWriter, r *http.Request) {
	claims, _ := clerk.SessionClaimsFromContext(r.Context())

	links, err := h.LinkService.ListAllLinks(r.Context(), claims.Subject)

	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, &dto.ErrorResponse{
			Error: dto.ErrorDetail{
				Code:    dto.ErrorInternal,
				Message: "Unable to create a new shortened URL",
			},
		})
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, &dto.SuccessReponse[[]db.Link]{
		Data:    links,
		Message: "User links retrieved successfully",
	})
}

// Get link by ID: GET /api/v1/links/{id}
func (h *LinkHandler) GetLink(w http.ResponseWriter, r *http.Request) {}

// Update link (PATCH code/expiry): PATCH /api/v1/links/{id}
func (h *LinkHandler) UpdateLink(w http.ResponseWriter, r *http.Request) {}

// Delete link by ID: DELETE /api/v1/links/{id}
func (h *LinkHandler) DeleteLink(w http.ResponseWriter, r *http.Request) {}
