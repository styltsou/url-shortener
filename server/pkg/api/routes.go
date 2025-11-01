// Package api defines routes and middleware
package api

import (
	"fmt"
	"net/http"

	"github.com/MarceloPetrucio/go-scalar-api-reference"
	clerkhttp "github.com/clerk/clerk-sdk-go/v2/http"
	"github.com/go-chi/chi/v5"

	"github.com/styltsou/url-shortener/server/pkg/handlers"
)

func NewRouter(linkH *handlers.LinkHandler) *chi.Mux {
	r := chi.NewRouter()

	r.Get("/reference", func(w http.ResponseWriter, r *http.Request) {
		htmlContent, err := scalar.ApiReferenceHTML(&scalar.Options{
			// SpecURL: "https://generator3.swagger.io/openapi.json",// allow external URL or local path file
			SpecURL: "./docs/swagger.json",
			CustomOptions: scalar.CustomOptions{
				PageTitle: "Simple API",
			},
			DarkMode: true,
		})

		// TODO: I could add a bit better error handling here
		if err != nil {
			fmt.Printf("%v", err)
		}

		fmt.Println(w, htmlContent)
	})

	// Public redirect endpoint
	r.Get("/{code}", linkH.Redirect)

	// Versioned API
	r.Route("/api/v1", func(r chi.Router) {
		r.Use(clerkhttp.RequireHeaderAuthorization())

		r.Route("/links", func(r chi.Router) {
			r.Get("/", linkH.ListLinks)
			r.Post("/", linkH.CreateLink)
			r.Get("/{id}", linkH.GetLink)
			r.Patch("/{id}", linkH.UpdateLink)
			r.Delete("/{id}", linkH.DeleteLink)
		})
	})

	return r
}
