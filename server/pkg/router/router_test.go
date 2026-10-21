package router

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/styltsou/url-shortener/server/pkg/logger"
)

func TestOpenAPIReferenceServesDocument(t *testing.T) {
	log, err := logger.New("test")
	if err != nil {
		t.Fatal(err)
	}
	r := New(nil, nil, nil, log)
	for _, test := range []struct {
		path string
		want string
	}{
		{"/api/v1/openapi.yaml", "openapi: 3.1.0"},
		{"/api/v1/reference", "/api/v1/openapi.yaml"},
	} {
		response := httptest.NewRecorder()
		r.ServeHTTP(response, httptest.NewRequest(http.MethodGet, test.path, nil))
		if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), test.want) {
			t.Fatalf("GET %s: status=%d, expected %q in body", test.path, response.Code, test.want)
		}
	}
}

func TestRedirectRouteUsesHandlerShortcodeParameter(t *testing.T) {
	log, err := logger.New("test")
	if err != nil {
		t.Fatal(err)
	}
	found := false
	err = chi.Walk(New(nil, nil, nil, log), func(method, route string, _ http.Handler, _ ...func(http.Handler) http.Handler) error {
		if method == http.MethodGet && route == "/{shortcode}" {
			found = true
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if !found {
		t.Fatal("public redirect route does not provide the shortcode parameter")
	}
}
