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

func New(linkH *handlers.LinkHandler, tagH *handlers.TagHandler, logger logger.Logger) *chi.Mux {
	r := chi.NewRouter()

	// Set custom NotFound handler
	r.NotFound(notFoundHandler(logger))

	// Set custom MethodNotAllowed handler
	r.MethodNotAllowed(methodNotAllowedHandler(logger))

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
		r.Use(mw.RequireAuth(logger))

		r.Route("/links", func(r chi.Router) {
			r.With(mw.RequestValidator[dto.CreateLink](logger)).Post("/", linkH.CreateLink)
			r.Get("/", linkH.ListLinks)
			r.Get("/{shortcode}", linkH.GetLink)
			r.With(mw.RequestValidator[dto.UpdateLink](logger)).Patch("/{id}", linkH.UpdateLink)
			r.Delete("/{id}", linkH.DeleteLink)

			// Tag assignment endpoints
			r.With(mw.RequestValidator[dto.AddTagsToLink](logger)).Post("/{id}/tags", linkH.AddTagsToLink)
			r.With(mw.RequestValidator[dto.RemoveTagsFromLink](logger)).Post("/{id}/tags/remove", linkH.RemoveTagsFromLink)
		})

		r.Route("/tags", func(r chi.Router) {
			r.Get("/", tagH.ListTags)
			r.With(mw.RequestValidator[dto.CreateTag](logger)).Post("/", tagH.CreateTag)
			r.With(mw.RequestValidator[dto.DeleteTags](logger)).Post("/bulk-delete", tagH.DeleteTags)
			r.With(mw.RequestValidator[dto.UpdateTag](logger)).Patch("/{id}", tagH.UpdateTag)
			r.Delete("/{id}", tagH.DeleteTag)
		})
	})

	return r
}

// notFoundHandler returns a handler for 404 Not Found errors
func notFoundHandler(logger logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Warn("Route not found",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
		)

		render.Status(r, http.StatusNotFound)
		render.JSON(w, r, dto.ErrorResponse{
			Error: dto.ErrorObject{
				Code:   apperrors.CodeNotFound,
				Title:  "Not Found",
				Detail: "The requested resource could not be found",
			},
		})
	}
}

// methodNotAllowedHandler returns a handler for 405 Method Not Allowed errors
func methodNotAllowedHandler(logger logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Warn("Method not allowed",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
		)

		render.Status(r, http.StatusMethodNotAllowed)
		render.JSON(w, r, dto.ErrorResponse{
			Error: dto.ErrorObject{
				Code:   apperrors.CodeMethodNotAllowed,
				Title:  "Method Not Allowed",
				Detail: "The requested HTTP method is not allowed for this resource",
			},
		})
	}
}
