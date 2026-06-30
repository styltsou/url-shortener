package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/styltsou/url-shortener/server/pkg/db"
	"github.com/styltsou/url-shortener/server/pkg/dto"
	mw "github.com/styltsou/url-shortener/server/pkg/middleware"
)

func TestLinkHTTPIntegrationCreateWithTags(t *testing.T) {
	userID := "user_123"
	tagID := uuid.New()
	linkID := uuid.New()

	handler := NewLinkHandler(&mockLinkService{
		CreateShortLinkWithTagsFunc: func(ctx context.Context, gotUserID string, originalURL string, customShortcode *string, expiresAt *time.Time, tagIDs []uuid.UUID) (db.GetLinkByIdAndUserWithTagsRow, error) {
			if gotUserID != userID {
				t.Fatalf("userID = %q, want %q", gotUserID, userID)
			}
			if originalURL != "https://example.com" {
				t.Fatalf("originalURL = %q, want https://example.com", originalURL)
			}
			if customShortcode == nil || *customShortcode != "demo" {
				t.Fatalf("customShortcode = %v, want demo", customShortcode)
			}
			if len(tagIDs) != 1 || tagIDs[0] != tagID {
				t.Fatalf("tagIDs = %v, want [%s]", tagIDs, tagID)
			}

			now := time.Now().UTC()
			tagsJSON, err := json.Marshal([]dto.TagResponse{{
				ID:        tagID,
				Name:      "launch",
				CreatedAt: now,
			}})
			if err != nil {
				t.Fatalf("marshal tags: %v", err)
			}

			return db.GetLinkByIdAndUserWithTagsRow{
				ID:          linkID,
				Shortcode:   "demo",
				OriginalUrl: originalURL,
				IsActive:    true,
				CreatedAt:   pgtype.Timestamp{Time: now, Valid: true},
				UpdatedAt:   pgtype.Timestamp{Time: now, Valid: true},
				Tags:        tagsJSON,
			}, nil
		},
	}, "http://localhost:8080", createTestLogger())

	router := newLinkIntegrationRouter(userID, func(r chi.Router) {
		r.With(mw.RequestValidator[dto.CreateLink](createTestLogger())).Post("/api/v1/links", handler.CreateLink)
	})

	body := []byte(`{"url":"https://example.com","shortcode":"demo","tag_ids":["` + tagID.String() + `"]}`)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/links", bytes.NewReader(body))
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusCreated, rec.Body.String())
	}

	var response dto.SuccessResponse[dto.LinkResponse]
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Data.Shortcode != "demo" {
		t.Fatalf("shortcode = %q, want demo", response.Data.Shortcode)
	}
	if len(response.Data.Tags) != 1 || response.Data.Tags[0].ID != tagID {
		t.Fatalf("tags = %+v, want tag %s", response.Data.Tags, tagID)
	}
}

func TestLinkHTTPIntegrationRejectsInvalidCreateJSON(t *testing.T) {
	handler := NewLinkHandler(&mockLinkService{}, "http://localhost:8080", createTestLogger())
	router := newLinkIntegrationRouter("user_123", func(r chi.Router) {
		r.With(mw.RequestValidator[dto.CreateLink](createTestLogger())).Post("/api/v1/links", handler.CreateLink)
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/links", bytes.NewReader([]byte(`{"url":`)))
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestLinkHTTPIntegrationClearsExpiration(t *testing.T) {
	userID := "user_123"
	linkID := uuid.New()

	handler := NewLinkHandler(&mockLinkService{
		UpdateLinkFunc: func(ctx context.Context, gotUserID string, gotLinkID uuid.UUID, shortcode *string, isActive *bool, expiresAtSet bool, expiresAt *time.Time) (db.UpdateLinkRow, error) {
			if gotUserID != userID {
				t.Fatalf("userID = %q, want %q", gotUserID, userID)
			}
			if gotLinkID != linkID {
				t.Fatalf("linkID = %s, want %s", gotLinkID, linkID)
			}
			if !expiresAtSet {
				t.Fatalf("expiresAtSet = false, want true")
			}
			if expiresAt != nil {
				t.Fatalf("expiresAt = %v, want nil", expiresAt)
			}

			now := time.Now().UTC()
			return db.UpdateLinkRow{
				ID:          linkID,
				Shortcode:   "demo",
				OriginalUrl: "https://example.com",
				IsActive:    true,
				CreatedAt:   pgtype.Timestamp{Time: now, Valid: true},
				UpdatedAt:   pgtype.Timestamp{Time: now, Valid: true},
				ExpiresAt:   pgtype.Timestamp{Valid: false},
			}, nil
		},
	}, "http://localhost:8080", createTestLogger())

	router := newLinkIntegrationRouter(userID, func(r chi.Router) {
		r.With(mw.RequestValidator[dto.UpdateLink](createTestLogger())).Patch("/api/v1/links/{id}", handler.UpdateLink)
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/links/"+linkID.String(), bytes.NewReader([]byte(`{"expires_at":null}`)))
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
}

func newLinkIntegrationRouter(userID string, routes func(chi.Router)) http.Handler {
	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			next.ServeHTTP(w, req.WithContext(mw.WithUserID(req.Context(), userID)))
		})
	})
	routes(r)
	return r
}
