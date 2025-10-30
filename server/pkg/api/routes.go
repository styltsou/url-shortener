// Package api defines routes and middleware
package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

/*
  - GET /{code}
	- GET, POST /api/v1/links
	- GET, PATCH, DELETE /api/v1/links/{code}
	- GET, POST /api/v1/me
*/

func NewRouter() *chi.Mux {
	r := chi.NewRouter()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("Hello, World!"))
		if err != nil {
			panic("Paniced")
		}
	})

	return r
}
