package router

import (
	"fmt"
	"net/http"

	"github.com/MarceloPetrucio/go-scalar-api-reference"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/styltsou/url-shortener/server/pkg/dto"
	apperrors "github.com/styltsou/url-shortener/server/pkg/errors"
	"github.com/styltsou/url-shortener/server/pkg/handlers"
	"github.com/styltsou/url-shortener/server/pkg/logger"
	mw "github.com/styltsou/url-shortener/server/pkg/middleware"
	"go.uber.org/zap"
)

func New(linkH *handlers.LinkHandler, logger logger.Logger) *chi.Mux {
	r := chi.NewRouter()

	r.Get("/{code}", linkH.Redirect)

	r.Get("/api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		render.Status(r, http.StatusOK)
		render.JSON(w, r, map[string]string{
			"status":  "ok",
			"service": "URL Shortener API",
		})
	})

	r.Get("/api/v1/reference", func(w http.ResponseWriter, r *http.Request) {
		htmlContent, err := scalar.ApiReferenceHTML(&scalar.Options{
			// SpecURL: "https://generator3.swagger.io/openapi.json",
			SpecURL: "./docs/openapi.yaml",
			CustomOptions: scalar.CustomOptions{
				PageTitle: "URL Shortener API",
			},
			DarkMode: true,
		})

		if err != nil {
			logger.Error("Failed to generate API reference",
				zap.Error(err),
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
			)

			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, dto.ErrorResponse{
				Error: dto.ErrorObject{
					Code:   apperrors.CodeInternalError,
					Title:  apperrors.InternalError.Error(),
					Detail: "Failed to generate API reference",
				},
			})
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, htmlContent)
	})

	r.Route("/api/v1", func(r chi.Router) {
		// TODO: Remove this bypass - development/testing only
		r.Use(mw.BypassAuth(logger))
		// r.Use(mw.RequireAuth(logger))  // Uncomment when done testing

		r.Route("/links", func(r chi.Router) {
			r.Get("/", linkH.ListLinks)
			r.With(mw.RequestValidator[dto.CreateLink](logger)).Post("/", linkH.CreateLink)
			r.Get("/{id}", linkH.GetLink)
			r.Patch("/{id}", linkH.UpdateLink)
			r.Delete("/{id}", linkH.DeleteLink)
		})
	})

	return r
}
