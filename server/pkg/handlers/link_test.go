package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/styltsou/url-shortener/server/pkg/db"
	"github.com/styltsou/url-shortener/server/pkg/dto"
	apperrors "github.com/styltsou/url-shortener/server/pkg/errors"
	"github.com/styltsou/url-shortener/server/pkg/logger"
	"github.com/styltsou/url-shortener/server/pkg/middleware"
)

// mockLinkService is a mock implementation of LinkServiceInterface
type mockLinkService struct {
	CreateShortLinkFunc    func(ctx context.Context, userID string, originalURL string) (db.TryCreateLinkRow, error)
	ListAllLinksFunc       func(ctx context.Context, userID string) ([]db.ListUserLinksRow, error)
	GetLinkByShortcodeFunc func(ctx context.Context, userID string, shortcode string) (db.GetLinkByShortcodeAndUserRow, error)
	GetOriginalURLFunc     func(ctx context.Context, code string) (db.GetLinkForRedirectRow, error)
	UpdateLinkFunc         func(ctx context.Context, userID string, id uuid.UUID, shortcode *string, isActive *bool, expiresAt *time.Time) (db.UpdateLinkRow, error)
	DeleteLinkFunc         func(ctx context.Context, userID string, id uuid.UUID) (db.DeleteLinkRow, error)
}

func (m *mockLinkService) CreateShortLink(ctx context.Context, userID string, originalURL string) (db.TryCreateLinkRow, error) {
	if m.CreateShortLinkFunc != nil {
		return m.CreateShortLinkFunc(ctx, userID, originalURL)
	}
	return db.TryCreateLinkRow{}, errors.New("not implemented")
}

func (m *mockLinkService) ListAllLinks(ctx context.Context, userID string) ([]db.ListUserLinksRow, error) {
	if m.ListAllLinksFunc != nil {
		return m.ListAllLinksFunc(ctx, userID)
	}
	return nil, errors.New("not implemented")
}

func (m *mockLinkService) GetLinkByShortcode(ctx context.Context, userID string, shortcode string) (db.GetLinkByShortcodeAndUserRow, error) {
	if m.GetLinkByShortcodeFunc != nil {
		return m.GetLinkByShortcodeFunc(ctx, userID, shortcode)
	}
	return db.GetLinkByShortcodeAndUserRow{}, errors.New("not implemented")
}

func (m *mockLinkService) GetOriginalURL(ctx context.Context, code string) (db.GetLinkForRedirectRow, error) {
	if m.GetOriginalURLFunc != nil {
		return m.GetOriginalURLFunc(ctx, code)
	}
	return db.GetLinkForRedirectRow{}, errors.New("not implemented")
}

func (m *mockLinkService) UpdateLink(ctx context.Context, userID string, id uuid.UUID, shortcode *string, isActive *bool, expiresAt *time.Time) (db.UpdateLinkRow, error) {
	if m.UpdateLinkFunc != nil {
		return m.UpdateLinkFunc(ctx, userID, id, shortcode, isActive, expiresAt)
	}
	return db.UpdateLinkRow{}, errors.New("not implemented")
}

func (m *mockLinkService) DeleteLink(ctx context.Context, userID string, id uuid.UUID) (db.DeleteLinkRow, error) {
	if m.DeleteLinkFunc != nil {
		return m.DeleteLinkFunc(ctx, userID, id)
	}
	return db.DeleteLinkRow{}, errors.New("not implemented")
}


func createTestLogger() logger.Logger {
	log, err := logger.New("test")
	if err != nil {
		panic("failed to create test logger: " + err.Error())
	}
	return log
}

func createTestLink(id uuid.UUID, shortcode, originalURL, userID string) db.Link {
	return db.Link{
		ID:          id,
		Shortcode:   shortcode,
		OriginalUrl: originalURL,
		UserID:      userID,
		ExpiresAt:   pgtype.Timestamp{Valid: false},
		CreatedAt:   pgtype.Timestamp{Valid: false},
		UpdatedAt:   pgtype.Timestamp{Valid: false},
	}
}

func TestLinkHandler_CreateLink(t *testing.T) {
	tests := []struct {
		name             string
		requestBody      dto.CreateLink
		userID           string
		mockService      *mockLinkService
		expectedStatus   int
		validateResponse func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name: "successful creation",
			requestBody: dto.CreateLink{
				URL: "https://example.com",
			},
			userID: "user_123",
			mockService: &mockLinkService{
				CreateShortLinkFunc: func(ctx context.Context, userID string, originalURL string) (db.TryCreateLinkRow, error) {
					if userID != "user_123" {
						t.Errorf("CreateShortLink called with wrong userID: got %s, want user_123", userID)
					}
					if originalURL != "https://example.com" {
						t.Errorf("CreateShortLink called with wrong URL: got %s, want https://example.com", originalURL)
					}
					return db.TryCreateLinkRow{
						ID:          uuid.New(),
						Shortcode:   "abc123",
						OriginalUrl: originalURL,
						ExpiresAt:   pgtype.Timestamp{Valid: false},
						IsActive:    true,
						CreatedAt:   pgtype.Timestamp{Valid: false},
						UpdatedAt:   pgtype.Timestamp{Valid: false},
					}, nil
				},
			},
			expectedStatus: http.StatusCreated,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response dto.SuccessResponse[db.TryCreateLinkRow]
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if response.Data.OriginalUrl != "https://example.com" {
					t.Errorf("Response OriginalUrl = %s, want https://example.com", response.Data.OriginalUrl)
				}
			},
		},
		// Note: "invalid JSON body" test is not applicable for handler unit tests
		// as the RequestValidator middleware would reject it before reaching the handler.
		// This should be tested in integration tests that include the middleware.
		{
			name: "invalid URL",
			requestBody: dto.CreateLink{
				URL: "invalid-url",
			},
			userID: "user_123",
			mockService: &mockLinkService{
				CreateShortLinkFunc: func(ctx context.Context, userID string, originalURL string) (db.TryCreateLinkRow, error) {
					return db.TryCreateLinkRow{}, apperrors.InvalidURL
				},
			},
			expectedStatus: http.StatusBadRequest,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response dto.ErrorResponse
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if response.Error.Code != apperrors.CodeInvalidURL {
					t.Errorf("Response Error.Code = %s, want %s", response.Error.Code, apperrors.CodeInvalidURL)
				}
			},
		},
		{
			name: "service error",
			requestBody: dto.CreateLink{
				URL: "https://example.com",
			},
			userID: "user_123",
			mockService: &mockLinkService{
				CreateShortLinkFunc: func(ctx context.Context, userID string, originalURL string) (db.TryCreateLinkRow, error) {
					return db.TryCreateLinkRow{}, errors.New("database error")
				},
			},
			expectedStatus: http.StatusInternalServerError,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response dto.ErrorResponse
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if response.Error.Code != apperrors.CodeInternalError {
					t.Errorf("Response Error.Code = %s, want %s", response.Error.Code, apperrors.CodeInternalError)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := &LinkHandler{
				LinkService: tt.mockService,
				logger:      createTestLogger(),
			}

			var reqBody []byte
			var err error
			if tt.name == "invalid JSON body" {
				reqBody = []byte("{ invalid json }")
			} else {
				reqBody, err = json.Marshal(tt.requestBody)
				if err != nil {
					t.Fatalf("Failed to marshal request body: %v", err)
				}
			}

			req := httptest.NewRequest(http.MethodPost, "/api/v1/links", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")
			ctx := middleware.WithUserID(req.Context(), tt.userID)
			// Set request body in context (same way RequestValidator middleware does)
			// Only set if not testing invalid JSON (that would fail validation middleware)
			if tt.name != "invalid JSON body" {
				ctx = context.WithValue(ctx, middleware.ReqBodyKey(), tt.requestBody)
			}
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			handler.CreateLink(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("CreateLink() status = %d, want %d", w.Code, tt.expectedStatus)
			}

			if tt.validateResponse != nil {
				tt.validateResponse(t, w)
			}
		})
	}
}

func TestLinkHandler_ListLinks(t *testing.T) {
	tests := []struct {
		name             string
		userID           string
		mockService      *mockLinkService
		expectedStatus   int
		validateResponse func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name:   "successful list with links",
			userID: "user_123",
			mockService: &mockLinkService{
				ListAllLinksFunc: func(ctx context.Context, userID string) ([]db.ListUserLinksRow, error) {
					if userID != "user_123" {
						t.Errorf("ListAllLinks called with wrong userID: got %s, want user_123", userID)
					}
					return []db.ListUserLinksRow{
						{
							ID:          uuid.New(),
							Shortcode:   "abc123",
							OriginalUrl: "https://example.com",
							ExpiresAt:   pgtype.Timestamp{Valid: false},
							IsActive:    true,
							CreatedAt:   pgtype.Timestamp{Valid: false},
							UpdatedAt:   pgtype.Timestamp{Valid: false},
							Tags:        nil,
						},
						{
							ID:          uuid.New(),
							Shortcode:   "xyz789",
							OriginalUrl: "https://example.org",
							ExpiresAt:   pgtype.Timestamp{Valid: false},
							IsActive:    true,
							CreatedAt:   pgtype.Timestamp{Valid: false},
							UpdatedAt:   pgtype.Timestamp{Valid: false},
							Tags:        nil,
						},
					}, nil
				},
			},
			expectedStatus: http.StatusOK,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response dto.SuccessResponse[[]db.ListUserLinksRow]
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if len(response.Data) != 2 {
					t.Errorf("Response Data length = %d, want 2", len(response.Data))
				}
			},
		},
		{
			name:   "successful list with no links",
			userID: "user_123",
			mockService: &mockLinkService{
				ListAllLinksFunc: func(ctx context.Context, userID string) ([]db.ListUserLinksRow, error) {
					return []db.ListUserLinksRow{}, nil
				},
			},
			expectedStatus: http.StatusOK,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response dto.SuccessResponse[[]db.ListUserLinksRow]
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if len(response.Data) != 0 {
					t.Errorf("Response Data length = %d, want 0", len(response.Data))
				}
			},
		},
		{
			name:   "service error",
			userID: "user_123",
			mockService: &mockLinkService{
				ListAllLinksFunc: func(ctx context.Context, userID string) ([]db.ListUserLinksRow, error) {
					return nil, errors.New("database error")
				},
			},
			expectedStatus: http.StatusInternalServerError,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response dto.ErrorResponse
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if response.Error.Code != apperrors.CodeInternalError {
					t.Errorf("Response Error.Code = %s, want %s", response.Error.Code, apperrors.CodeInternalError)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := &LinkHandler{
				LinkService: tt.mockService,
				logger:      createTestLogger(),
			}

			req := httptest.NewRequest(http.MethodGet, "/api/v1/links", nil)
			ctx := middleware.WithUserID(req.Context(), tt.userID)
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			handler.ListLinks(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("ListLinks() status = %d, want %d", w.Code, tt.expectedStatus)
			}

			if tt.validateResponse != nil {
				tt.validateResponse(t, w)
			}
		})
	}
}

func TestLinkHandler_UpdateLink(t *testing.T) {
	linkID := uuid.New()
	userID := "user_123"
	newShortcode := "newcode"
	isActive := false

	tests := []struct {
		name            string
		linkID          string
		userID          string
		requestBody     dto.UpdateLink
		mockService     *mockLinkService
		expectedStatus  int
		validateResponse func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name:   "successful update shortcode",
			linkID: linkID.String(),
			userID: userID,
			requestBody: dto.UpdateLink{
				Shortcode: &newShortcode,
			},
			mockService: &mockLinkService{
				UpdateLinkFunc: func(ctx context.Context, userIDParam string, id uuid.UUID, shortcode *string, isActive *bool, expiresAt *time.Time) (db.UpdateLinkRow, error) {
					if id != linkID {
						t.Errorf("UpdateLink called with wrong ID")
					}
					if userIDParam != userID {
						t.Errorf("UpdateLink called with wrong userID: got %s, want %s", userIDParam, userID)
					}
					return db.UpdateLinkRow{
						ID:          id,
						Shortcode:   *shortcode,
						OriginalUrl: "https://example.com",
						IsActive:    true,
						ExpiresAt:   pgtype.Timestamp{Valid: false},
						CreatedAt:   pgtype.Timestamp{Valid: false},
						UpdatedAt:   pgtype.Timestamp{Valid: false},
					}, nil
				},
			},
			expectedStatus: http.StatusOK,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response dto.SuccessResponse[db.UpdateLinkRow]
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if response.Data.Shortcode != newShortcode {
					t.Errorf("Response Data.Shortcode = %s, want %s", response.Data.Shortcode, newShortcode)
				}
			},
		},
		{
			name:   "successful update is_active",
			linkID: linkID.String(),
			userID: userID,
			requestBody: dto.UpdateLink{
				IsActive: &isActive,
			},
			mockService: &mockLinkService{
				UpdateLinkFunc: func(ctx context.Context, userID string, id uuid.UUID, shortcode *string, isActive *bool, expiresAt *time.Time) (db.UpdateLinkRow, error) {
					return db.UpdateLinkRow{
						ID:          id,
						Shortcode:   "oldcode",
						OriginalUrl: "https://example.com",
						IsActive:    *isActive,
						ExpiresAt:   pgtype.Timestamp{Valid: false},
						CreatedAt:   pgtype.Timestamp{Valid: false},
						UpdatedAt:   pgtype.Timestamp{Valid: false},
					}, nil
				},
			},
			expectedStatus: http.StatusOK,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response dto.SuccessResponse[db.UpdateLinkRow]
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if response.Data.IsActive != false {
					t.Errorf("Response Data.IsActive = %v, want false", response.Data.IsActive)
				}
			},
		},
		{
			name:   "invalid UUID format",
			linkID: "invalid-uuid",
			userID: userID,
			requestBody: dto.UpdateLink{
				Shortcode: &newShortcode,
			},
			mockService:    &mockLinkService{},
			expectedStatus: http.StatusBadRequest,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response dto.ErrorResponse
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if response.Error.Code != apperrors.CodeInvalidID {
					t.Errorf("Response Error.Code = %s, want %s", response.Error.Code, apperrors.CodeInvalidID)
				}
			},
		},
		{
			name:   "link not found",
			linkID: linkID.String(),
			userID: userID,
			requestBody: dto.UpdateLink{
				Shortcode: &newShortcode,
			},
			mockService: &mockLinkService{
				UpdateLinkFunc: func(ctx context.Context, userID string, id uuid.UUID, shortcode *string, isActive *bool, expiresAt *time.Time) (db.UpdateLinkRow, error) {
					return db.UpdateLinkRow{}, apperrors.LinkNotFound
				},
			},
			expectedStatus: http.StatusNotFound,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response dto.ErrorResponse
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if response.Error.Code != apperrors.CodeLinkNotFound {
					t.Errorf("Response Error.Code = %s, want %s", response.Error.Code, apperrors.CodeLinkNotFound)
				}
			},
		},
		{
			name:   "shortcode already taken",
			linkID: linkID.String(),
			userID: userID,
			requestBody: dto.UpdateLink{
				Shortcode: &newShortcode,
			},
			mockService: &mockLinkService{
				UpdateLinkFunc: func(ctx context.Context, userID string, id uuid.UUID, shortcode *string, isActive *bool, expiresAt *time.Time) (db.UpdateLinkRow, error) {
					return db.UpdateLinkRow{}, apperrors.LinkShortcodeTaken
				},
			},
			expectedStatus: http.StatusConflict,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response dto.ErrorResponse
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if response.Error.Code != apperrors.CodeCodeTaken {
					t.Errorf("Response Error.Code = %s, want %s", response.Error.Code, apperrors.CodeCodeTaken)
				}
			},
		},
		{
			name:   "service error",
			linkID: linkID.String(),
			userID: userID,
			requestBody: dto.UpdateLink{
				Shortcode: &newShortcode,
			},
			mockService: &mockLinkService{
				UpdateLinkFunc: func(ctx context.Context, userID string, id uuid.UUID, shortcode *string, isActive *bool, expiresAt *time.Time) (db.UpdateLinkRow, error) {
					return db.UpdateLinkRow{}, errors.New("database error")
				},
			},
			expectedStatus: http.StatusInternalServerError,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response dto.ErrorResponse
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if response.Error.Code != apperrors.CodeInternalError {
					t.Errorf("Response Error.Code = %s, want %s", response.Error.Code, apperrors.CodeInternalError)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := &LinkHandler{
				LinkService: tt.mockService,
				logger:      createTestLogger(),
			}

			bodyBytes, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPatch, "/api/v1/links/"+tt.linkID, bytes.NewBuffer(bodyBytes))
			ctx := middleware.WithUserID(req.Context(), tt.userID)
			// Set request body in context (same way RequestValidator middleware does)
			ctx = context.WithValue(ctx, middleware.ReqBodyKey(), tt.requestBody)
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			
			// Create a chi router for testing to handle URL params
			r := chi.NewRouter()
			r.Patch("/api/v1/links/{id}", handler.UpdateLink)
			r.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("UpdateLink() status = %d, want %d", w.Code, tt.expectedStatus)
			}

			if tt.validateResponse != nil {
				tt.validateResponse(t, w)
			}
		})
	}
}

// Note: Error mapping is now tested in pkg/errors/errors_test.go via TestMapError
// The error handling middleware is tested through integration tests
