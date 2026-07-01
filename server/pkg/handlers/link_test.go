package handlers

//nolint:goconst // test fixtures, not worth extracting

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
	"github.com/styltsou/url-shortener/server/pkg/analytics"
	"github.com/styltsou/url-shortener/server/pkg/db"
	"github.com/styltsou/url-shortener/server/pkg/dto"
	apperrors "github.com/styltsou/url-shortener/server/pkg/errors"
	"github.com/styltsou/url-shortener/server/pkg/logger"
	"github.com/styltsou/url-shortener/server/pkg/middleware"
	"github.com/styltsou/url-shortener/server/pkg/service"
)

// mockLinkService is a mock implementation of LinkServiceInterface
type mockLinkService struct {
	CreateShortLinkFunc         func(ctx context.Context, userID string, originalURL string, customShortcode *string, expiresAt *time.Time) (db.TryCreateLinkRow, error)
	CreateShortLinkWithTagsFunc func(ctx context.Context, userID string, originalURL string, customShortcode *string, expiresAt *time.Time, tagIDs []uuid.UUID) (db.GetLinkByIdAndUserWithTagsRow, error)
	ListAllLinksFunc            func(ctx context.Context, userID string, isActive *bool, tagIDs []uuid.UUID, page, limit int) (*service.ListLinksResult, error)
	GetLinkByShortcodeFunc      func(ctx context.Context, userID string, shortcode string) (db.GetLinkByShortcodeAndUserRow, error)
	GetOriginalURLFunc          func(ctx context.Context, code string) (db.GetLinkForRedirectRow, error)
	UpdateLinkFunc              func(ctx context.Context, userID string, id uuid.UUID, shortcode *string, isActive *bool, expiresAtSet bool, expiresAt *time.Time) (db.UpdateLinkRow, error)
	DeleteLinkFunc              func(ctx context.Context, userID string, id uuid.UUID) (db.DeleteLinkRow, error)
	AddTagsToLinkFunc           func(ctx context.Context, userID string, linkID uuid.UUID, tagIDs []uuid.UUID) (db.GetLinkByIdAndUserWithTagsRow, error)
	RemoveTagsFromLinkFunc      func(ctx context.Context, userID string, linkID uuid.UUID, tagIDs []uuid.UUID) (db.GetLinkByIdAndUserWithTagsRow, error)
	RecordClickFunc             func(ctx context.Context, linkID uuid.UUID, ip, userAgent, referrer string)
	GetLinkAnalyticsFunc        func(ctx context.Context, userID string, shortcode string) (*analytics.LinkAnalytics, error)
	GetDashboardStatsFunc       func(ctx context.Context, userID string) (*service.DashboardStats, error)
}

func (m *mockLinkService) CreateShortLink(ctx context.Context, userID string, originalURL string, customShortcode *string, expiresAt *time.Time) (db.TryCreateLinkRow, error) {
	if m.CreateShortLinkFunc != nil {
		return m.CreateShortLinkFunc(ctx, userID, originalURL, customShortcode, expiresAt)
	}
	return db.TryCreateLinkRow{}, errors.New("not implemented")
}

func (m *mockLinkService) CreateShortLinkWithTags(ctx context.Context, userID string, originalURL string, customShortcode *string, expiresAt *time.Time, tagIDs []uuid.UUID) (db.GetLinkByIdAndUserWithTagsRow, error) {
	if m.CreateShortLinkWithTagsFunc != nil {
		return m.CreateShortLinkWithTagsFunc(ctx, userID, originalURL, customShortcode, expiresAt, tagIDs)
	}
	link, err := m.CreateShortLink(ctx, userID, originalURL, customShortcode, expiresAt)
	if err != nil {
		return db.GetLinkByIdAndUserWithTagsRow{}, err
	}
	return db.GetLinkByIdAndUserWithTagsRow{
		ID:          link.ID,
		Shortcode:   link.Shortcode,
		OriginalUrl: link.OriginalUrl,
		ExpiresAt:   link.ExpiresAt,
		IsActive:    link.IsActive,
		CreatedAt:   link.CreatedAt,
		UpdatedAt:   link.UpdatedAt,
		Tags:        []interface{}{},
	}, nil
}

func (m *mockLinkService) ListAllLinks(ctx context.Context, userID string, isActive *bool, tagIDs []uuid.UUID, page, limit int) (*service.ListLinksResult, error) {
	if m.ListAllLinksFunc != nil {
		return m.ListAllLinksFunc(ctx, userID, isActive, tagIDs, page, limit)
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

func (m *mockLinkService) UpdateLink(ctx context.Context, userID string, id uuid.UUID, shortcode *string, isActive *bool, expiresAtSet bool, expiresAt *time.Time) (db.UpdateLinkRow, error) {
	if m.UpdateLinkFunc != nil {
		return m.UpdateLinkFunc(ctx, userID, id, shortcode, isActive, expiresAtSet, expiresAt)
	}
	return db.UpdateLinkRow{}, errors.New("not implemented")
}

func (m *mockLinkService) DeleteLink(ctx context.Context, userID string, id uuid.UUID) (db.DeleteLinkRow, error) {
	if m.DeleteLinkFunc != nil {
		return m.DeleteLinkFunc(ctx, userID, id)
	}
	return db.DeleteLinkRow{}, errors.New("not implemented")
}

func (m *mockLinkService) AddTagsToLink(ctx context.Context, userID string, linkID uuid.UUID, tagIDs []uuid.UUID) (db.GetLinkByIdAndUserWithTagsRow, error) {
	if m.AddTagsToLinkFunc != nil {
		return m.AddTagsToLinkFunc(ctx, userID, linkID, tagIDs)
	}
	return db.GetLinkByIdAndUserWithTagsRow{}, errors.New("not implemented")
}

func (m *mockLinkService) RemoveTagsFromLink(ctx context.Context, userID string, linkID uuid.UUID, tagIDs []uuid.UUID) (db.GetLinkByIdAndUserWithTagsRow, error) {
	if m.RemoveTagsFromLinkFunc != nil {
		return m.RemoveTagsFromLinkFunc(ctx, userID, linkID, tagIDs)
	}
	return db.GetLinkByIdAndUserWithTagsRow{}, errors.New("not implemented")
}

func (m *mockLinkService) RecordClick(ctx context.Context, linkID uuid.UUID, ip, userAgent, referrer string) {
	if m.RecordClickFunc != nil {
		m.RecordClickFunc(ctx, linkID, ip, userAgent, referrer)
	}
}

func (m *mockLinkService) GetLinkAnalytics(ctx context.Context, userID string, shortcode string) (*analytics.LinkAnalytics, error) {
	if m.GetLinkAnalyticsFunc != nil {
		return m.GetLinkAnalyticsFunc(ctx, userID, shortcode)
	}
	return &analytics.LinkAnalytics{}, nil
}

func (m *mockLinkService) GetDashboardStats(ctx context.Context, userID string) (*service.DashboardStats, error) {
	if m.GetDashboardStatsFunc != nil {
		return m.GetDashboardStatsFunc(ctx, userID)
	}
	return &service.DashboardStats{}, errors.New("not implemented")
}

func createTestLogger() logger.Logger {
	log, err := logger.New("test")
	if err != nil {
		panic("failed to create test logger: " + err.Error())
	}
	return log
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
				CreateShortLinkFunc: func(ctx context.Context, userID string, originalURL string, customShortcode *string, expiresAt *time.Time) (db.TryCreateLinkRow, error) {
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
				CreateShortLinkFunc: func(ctx context.Context, userID string, originalURL string, customShortcode *string, expiresAt *time.Time) (db.TryCreateLinkRow, error) {
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
				CreateShortLinkFunc: func(ctx context.Context, userID string, originalURL string, customShortcode *string, expiresAt *time.Time) (db.TryCreateLinkRow, error) {
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
				ctx = middleware.CtxWithRequestBody(ctx, tt.requestBody)
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
				ListAllLinksFunc: func(ctx context.Context, userID string, isActive *bool, tagIDs []uuid.UUID, page, limit int) (*service.ListLinksResult, error) {
					if userID != "user_123" {
						t.Errorf("ListAllLinks called with wrong userID: got %s, want user_123", userID)
					}
					return &service.ListLinksResult{
						Links: []db.ListUserLinksRow{
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
						},
						Total:      2,
						Page:       1,
						Limit:      5,
						TotalPages: 1,
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
				if response.Pagination == nil {
					t.Errorf("Response Pagination should not be nil")
				} else if response.Pagination.Total != 2 {
					t.Errorf("Response Pagination.Total = %d, want 2", response.Pagination.Total)
				}
			},
		},
		{
			name:   "successful list with no links",
			userID: "user_123",
			mockService: &mockLinkService{
				ListAllLinksFunc: func(ctx context.Context, userID string, isActive *bool, tagIDs []uuid.UUID, page, limit int) (*service.ListLinksResult, error) {
					return &service.ListLinksResult{
						Links:      []db.ListUserLinksRow{},
						Total:      0,
						Page:       1,
						Limit:      5,
						TotalPages: 0,
					}, nil
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
				if response.Pagination == nil {
					t.Errorf("Response Pagination should not be nil")
				} else if response.Pagination.Total != 0 {
					t.Errorf("Response Pagination.Total = %d, want 0", response.Pagination.Total)
				}
			},
		},
		{
			name:   "service error",
			userID: "user_123",
			mockService: &mockLinkService{
				ListAllLinksFunc: func(ctx context.Context, userID string, isActive *bool, tagIDs []uuid.UUID, page, limit int) (*service.ListLinksResult, error) {
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
		name             string
		linkID           string
		userID           string
		requestBody      dto.UpdateLink
		mockService      *mockLinkService
		expectedStatus   int
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
				UpdateLinkFunc: func(ctx context.Context, userIDParam string, id uuid.UUID, shortcode *string, isActive *bool, expiresAtSet bool, expiresAt *time.Time) (db.UpdateLinkRow, error) {
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
				UpdateLinkFunc: func(ctx context.Context, userID string, id uuid.UUID, shortcode *string, isActive *bool, expiresAtSet bool, expiresAt *time.Time) (db.UpdateLinkRow, error) {
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
				UpdateLinkFunc: func(ctx context.Context, userID string, id uuid.UUID, shortcode *string, isActive *bool, expiresAtSet bool, expiresAt *time.Time) (db.UpdateLinkRow, error) {
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
				UpdateLinkFunc: func(ctx context.Context, userID string, id uuid.UUID, shortcode *string, isActive *bool, expiresAtSet bool, expiresAt *time.Time) (db.UpdateLinkRow, error) {
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
				UpdateLinkFunc: func(ctx context.Context, userID string, id uuid.UUID, shortcode *string, isActive *bool, expiresAtSet bool, expiresAt *time.Time) (db.UpdateLinkRow, error) {
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
			ctx = middleware.CtxWithRequestBody(ctx, tt.requestBody)
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

func TestLinkHandler_GetLink(t *testing.T) {
	userID := "user_123"
	shortcode := "abc123" //nolint:goconst
	linkID := uuid.New()

	tests := []struct {
		name             string
		shortcode        string
		userID           string
		mockService      *mockLinkService
		expectedStatus   int
		validateResponse func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name:      "successful get",
			shortcode: shortcode,
			userID:    userID,
			mockService: &mockLinkService{
				GetLinkByShortcodeFunc: func(ctx context.Context, uid string, code string) (db.GetLinkByShortcodeAndUserRow, error) {
					if uid != userID {
						t.Errorf("GetLinkByShortcode called with wrong userID: got %s, want %s", uid, userID)
					}
					if code != shortcode {
						t.Errorf("GetLinkByShortcode called with wrong shortcode: got %s, want %s", code, shortcode)
					}
					return db.GetLinkByShortcodeAndUserRow{
						ID:          linkID,
						Shortcode:   shortcode,
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
				var response dto.SuccessResponse[db.GetLinkByShortcodeAndUserRow]
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if response.Data.Shortcode != shortcode {
					t.Errorf("Response Data.Shortcode = %s, want %s", response.Data.Shortcode, shortcode)
				}
			},
		},
		{
			name:      "link not found",
			shortcode: shortcode,
			userID:    userID,
			mockService: &mockLinkService{
				GetLinkByShortcodeFunc: func(ctx context.Context, uid string, code string) (db.GetLinkByShortcodeAndUserRow, error) {
					return db.GetLinkByShortcodeAndUserRow{}, apperrors.LinkNotFound
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
			name:      "service error",
			shortcode: shortcode,
			userID:    userID,
			mockService: &mockLinkService{
				GetLinkByShortcodeFunc: func(ctx context.Context, uid string, code string) (db.GetLinkByShortcodeAndUserRow, error) {
					return db.GetLinkByShortcodeAndUserRow{}, errors.New("database error")
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

			req := httptest.NewRequest(http.MethodGet, "/api/v1/links/"+tt.shortcode, nil)
			ctx := middleware.WithUserID(req.Context(), tt.userID)
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()

			// Create a chi router for testing to handle URL params
			r := chi.NewRouter()
			r.Get("/api/v1/links/{shortcode}", handler.GetLink)
			r.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("GetLink() status = %d, want %d", w.Code, tt.expectedStatus)
			}

			if tt.validateResponse != nil {
				tt.validateResponse(t, w)
			}
		})
	}
}

func TestLinkHandler_DeleteLink(t *testing.T) {
	linkID := uuid.New()
	userID := "user_123"

	tests := []struct {
		name             string
		linkID           string
		userID           string
		mockService      *mockLinkService
		expectedStatus   int
		validateResponse func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name:   "successful delete",
			linkID: linkID.String(),
			userID: userID,
			mockService: &mockLinkService{
				DeleteLinkFunc: func(ctx context.Context, uid string, id uuid.UUID) (db.DeleteLinkRow, error) {
					if id != linkID {
						t.Errorf("DeleteLink called with wrong ID")
					}
					if uid != userID {
						t.Errorf("DeleteLink called with wrong userID: got %s, want %s", uid, userID)
					}
					return db.DeleteLinkRow{
						ID:          linkID,
						Shortcode:   "abc123",
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
				var response dto.SuccessResponse[db.DeleteLinkRow]
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if response.Data.ID != linkID {
					t.Errorf("Response Data.ID = %s, want %s", response.Data.ID, linkID)
				}
			},
		},
		{
			name:           "invalid UUID format",
			linkID:         "invalid-uuid",
			userID:         userID,
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
			mockService: &mockLinkService{
				DeleteLinkFunc: func(ctx context.Context, uid string, id uuid.UUID) (db.DeleteLinkRow, error) {
					return db.DeleteLinkRow{}, apperrors.LinkNotFound
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
			name:   "service error",
			linkID: linkID.String(),
			userID: userID,
			mockService: &mockLinkService{
				DeleteLinkFunc: func(ctx context.Context, uid string, id uuid.UUID) (db.DeleteLinkRow, error) {
					return db.DeleteLinkRow{}, errors.New("database error")
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

			req := httptest.NewRequest(http.MethodDelete, "/api/v1/links/"+tt.linkID, nil)
			ctx := middleware.WithUserID(req.Context(), tt.userID)
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()

			// Create a chi router for testing to handle URL params
			r := chi.NewRouter()
			r.Delete("/api/v1/links/{id}", handler.DeleteLink)
			r.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("DeleteLink() status = %d, want %d", w.Code, tt.expectedStatus)
			}

			if tt.validateResponse != nil {
				tt.validateResponse(t, w)
			}
		})
	}
}

func TestLinkHandler_AddTagsToLink(t *testing.T) {
	linkID := uuid.New()
	userID := "user_123"
	tagID1 := uuid.New()
	tagID2 := uuid.New()

	tests := []struct {
		name             string
		linkID           string
		userID           string
		requestBody      dto.AddTagsToLink
		mockService      *mockLinkService
		expectedStatus   int
		validateResponse func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name:   "successful add tags",
			linkID: linkID.String(),
			userID: userID,
			requestBody: dto.AddTagsToLink{
				TagIDs: []uuid.UUID{tagID1, tagID2},
			},
			mockService: &mockLinkService{
				AddTagsToLinkFunc: func(ctx context.Context, uid string, id uuid.UUID, tagIDs []uuid.UUID) (db.GetLinkByIdAndUserWithTagsRow, error) {
					if id != linkID {
						t.Errorf("AddTagsToLink called with wrong linkID")
					}
					if uid != userID {
						t.Errorf("AddTagsToLink called with wrong userID: got %s, want %s", uid, userID)
					}
					if len(tagIDs) != 2 || tagIDs[0] != tagID1 || tagIDs[1] != tagID2 {
						t.Errorf("AddTagsToLink called with wrong tagIDs")
					}
					return db.GetLinkByIdAndUserWithTagsRow{
						ID:          linkID,
						Shortcode:   "abc123",
						OriginalUrl: "https://example.com",
						IsActive:    true,
						ExpiresAt:   pgtype.Timestamp{Valid: false},
						CreatedAt:   pgtype.Timestamp{Valid: false},
						UpdatedAt:   pgtype.Timestamp{Valid: false},
						Tags:        []interface{}{},
					}, nil
				},
			},
			expectedStatus: http.StatusOK,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response dto.SuccessResponse[db.GetLinkByIdAndUserWithTagsRow]
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if response.Data.ID != linkID {
					t.Errorf("Response Data.ID = %s, want %s", response.Data.ID, linkID)
				}
			},
		},
		{
			name:   "invalid UUID format",
			linkID: "invalid-uuid",
			userID: userID,
			requestBody: dto.AddTagsToLink{
				TagIDs: []uuid.UUID{tagID1},
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
			name:   "service error",
			linkID: linkID.String(),
			userID: userID,
			requestBody: dto.AddTagsToLink{
				TagIDs: []uuid.UUID{tagID1},
			},
			mockService: &mockLinkService{
				AddTagsToLinkFunc: func(ctx context.Context, uid string, id uuid.UUID, tagIDs []uuid.UUID) (db.GetLinkByIdAndUserWithTagsRow, error) {
					return db.GetLinkByIdAndUserWithTagsRow{}, errors.New("database error")
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
			req := httptest.NewRequest(http.MethodPost, "/api/v1/links/"+tt.linkID+"/tags", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			ctx := middleware.WithUserID(req.Context(), tt.userID)
			ctx = middleware.CtxWithRequestBody(ctx, tt.requestBody)
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()

			// Create a chi router for testing to handle URL params
			r := chi.NewRouter()
			r.Post("/api/v1/links/{id}/tags", handler.AddTagsToLink)
			r.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("AddTagsToLink() status = %d, want %d", w.Code, tt.expectedStatus)
			}

			if tt.validateResponse != nil {
				tt.validateResponse(t, w)
			}
		})
	}
}

func TestLinkHandler_RemoveTagsFromLink(t *testing.T) {
	linkID := uuid.New()
	userID := "user_123"
	tagID1 := uuid.New()
	tagID2 := uuid.New()

	tests := []struct {
		name             string
		linkID           string
		userID           string
		requestBody      dto.RemoveTagsFromLink
		mockService      *mockLinkService
		expectedStatus   int
		validateResponse func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name:   "successful remove tags",
			linkID: linkID.String(),
			userID: userID,
			requestBody: dto.RemoveTagsFromLink{
				TagIDs: []uuid.UUID{tagID1, tagID2},
			},
			mockService: &mockLinkService{
				RemoveTagsFromLinkFunc: func(ctx context.Context, uid string, id uuid.UUID, tagIDs []uuid.UUID) (db.GetLinkByIdAndUserWithTagsRow, error) {
					if id != linkID {
						t.Errorf("RemoveTagsFromLink called with wrong linkID")
					}
					if uid != userID {
						t.Errorf("RemoveTagsFromLink called with wrong userID: got %s, want %s", uid, userID)
					}
					if len(tagIDs) != 2 || tagIDs[0] != tagID1 || tagIDs[1] != tagID2 {
						t.Errorf("RemoveTagsFromLink called with wrong tagIDs")
					}
					return db.GetLinkByIdAndUserWithTagsRow{
						ID:          linkID,
						Shortcode:   "abc123",
						OriginalUrl: "https://example.com",
						IsActive:    true,
						ExpiresAt:   pgtype.Timestamp{Valid: false},
						CreatedAt:   pgtype.Timestamp{Valid: false},
						UpdatedAt:   pgtype.Timestamp{Valid: false},
						Tags:        []interface{}{},
					}, nil
				},
			},
			expectedStatus: http.StatusOK,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response dto.SuccessResponse[db.GetLinkByIdAndUserWithTagsRow]
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if response.Data.ID != linkID {
					t.Errorf("Response Data.ID = %s, want %s", response.Data.ID, linkID)
				}
			},
		},
		{
			name:   "invalid UUID format",
			linkID: "invalid-uuid",
			userID: userID,
			requestBody: dto.RemoveTagsFromLink{
				TagIDs: []uuid.UUID{tagID1},
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
			name:   "service error",
			linkID: linkID.String(),
			userID: userID,
			requestBody: dto.RemoveTagsFromLink{
				TagIDs: []uuid.UUID{tagID1},
			},
			mockService: &mockLinkService{
				RemoveTagsFromLinkFunc: func(ctx context.Context, uid string, id uuid.UUID, tagIDs []uuid.UUID) (db.GetLinkByIdAndUserWithTagsRow, error) {
					return db.GetLinkByIdAndUserWithTagsRow{}, errors.New("database error")
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
			req := httptest.NewRequest(http.MethodDelete, "/api/v1/links/"+tt.linkID+"/tags", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			ctx := middleware.WithUserID(req.Context(), tt.userID)
			ctx = middleware.CtxWithRequestBody(ctx, tt.requestBody)
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()

			// Create a chi router for testing to handle URL params
			r := chi.NewRouter()
			r.Delete("/api/v1/links/{id}/tags", handler.RemoveTagsFromLink)
			r.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("RemoveTagsFromLink() status = %d, want %d", w.Code, tt.expectedStatus)
			}

			if tt.validateResponse != nil {
				tt.validateResponse(t, w)
			}
		})
	}
}

func TestLinkHandler_Redirect(t *testing.T) {
	shortcode := "abc123"
	originalURL := "https://example.com"

	tests := []struct {
		name           string
		shortcode      string
		mockService    *mockLinkService
		expectedStatus int
		expectedURL    string
	}{
		{
			name:      "successful redirect",
			shortcode: shortcode,
			mockService: &mockLinkService{
				GetOriginalURLFunc: func(ctx context.Context, code string) (db.GetLinkForRedirectRow, error) {
					if code != shortcode {
						t.Errorf("GetOriginalURL called with wrong shortcode: got %s, want %s", code, shortcode)
					}
					return db.GetLinkForRedirectRow{
						ID:          uuid.New(),
						OriginalUrl: originalURL,
					}, nil
				},
			},
			expectedStatus: http.StatusFound,
			expectedURL:    originalURL,
		},
		{
			name:      "link not found",
			shortcode: shortcode,
			mockService: &mockLinkService{
				GetOriginalURLFunc: func(ctx context.Context, code string) (db.GetLinkForRedirectRow, error) {
					return db.GetLinkForRedirectRow{}, apperrors.LinkNotFound
				},
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := &LinkHandler{
				LinkService: tt.mockService,
				logger:      createTestLogger(),
			}

			req := httptest.NewRequest(http.MethodGet, "/"+tt.shortcode, nil)
			w := httptest.NewRecorder()

			// Create a chi router for testing to handle URL params
			r := chi.NewRouter()
			r.Get("/{shortcode}", handler.Redirect)
			r.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Redirect() status = %d, want %d", w.Code, tt.expectedStatus)
			}

			if tt.expectedStatus == http.StatusFound {
				location := w.Header().Get("Location")
				if location != tt.expectedURL {
					t.Errorf("Redirect() Location = %s, want %s", location, tt.expectedURL)
				}
			}
		})
	}
}

func TestLinkHandler_GetQRCode(t *testing.T) {
	shortcode := "abc123"

	t.Run("successful QR code generation", func(t *testing.T) {
		handler := &LinkHandler{
			LinkService: &mockLinkService{},
			baseURL:     "http://localhost:8080",
			logger:      createTestLogger(),
		}

		req := httptest.NewRequest(http.MethodGet, "/api/v1/links/"+shortcode+"/qrcode", nil)
		w := httptest.NewRecorder()

		r := chi.NewRouter()
		r.Get("/api/v1/links/{shortcode}/qrcode", handler.GetQRCode)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("GetQRCode() status = %d, want %d", w.Code, http.StatusOK)
		}
		contentType := w.Header().Get("Content-Type")
		if contentType != "image/png" {
			t.Errorf("GetQRCode() Content-Type = %s, want image/png", contentType)
		}
		if w.Body.Len() == 0 {
			t.Errorf("GetQRCode() returned empty body")
		}
	})

	t.Run("uses custom size parameter", func(t *testing.T) {
		handler := &LinkHandler{
			LinkService: &mockLinkService{},
			baseURL:     "http://localhost:8080",
			logger:      createTestLogger(),
		}

		req := httptest.NewRequest(http.MethodGet, "/api/v1/links/"+shortcode+"/qrcode?size=512", nil)
		w := httptest.NewRecorder()

		r := chi.NewRouter()
		r.Get("/api/v1/links/{shortcode}/qrcode", handler.GetQRCode)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("GetQRCode() with size=512 status = %d, want %d", w.Code, http.StatusOK)
		}
	})

	t.Run("clamps size to valid range (below 64)", func(t *testing.T) {
		handler := &LinkHandler{
			LinkService: &mockLinkService{},
			baseURL:     "http://localhost:8080",
			logger:      createTestLogger(),
		}

		req := httptest.NewRequest(http.MethodGet, "/api/v1/links/"+shortcode+"/qrcode?size=32", nil)
		w := httptest.NewRecorder()

		r := chi.NewRouter()
		r.Get("/api/v1/links/{shortcode}/qrcode", handler.GetQRCode)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("GetQRCode() with invalid size status = %d, want %d", w.Code, http.StatusOK)
		}
	})

	t.Run("clamps size to valid range (above 1024)", func(t *testing.T) {
		handler := &LinkHandler{
			LinkService: &mockLinkService{},
			baseURL:     "http://localhost:8080",
			logger:      createTestLogger(),
		}

		req := httptest.NewRequest(http.MethodGet, "/api/v1/links/"+shortcode+"/qrcode?size=2048", nil)
		w := httptest.NewRecorder()

		r := chi.NewRouter()
		r.Get("/api/v1/links/{shortcode}/qrcode", handler.GetQRCode)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("GetQRCode() with oversized size status = %d, want %d", w.Code, http.StatusOK)
		}
	})
}

func TestLinkHandler_GetDashboard(t *testing.T) {
	userID := "user_123"

	t.Run("successful dashboard", func(t *testing.T) {
		handler := &LinkHandler{
			LinkService: &mockLinkService{
				GetDashboardStatsFunc: func(ctx context.Context, uid string) (*service.DashboardStats, error) {
					if uid != userID {
						t.Errorf("GetDashboardStats called with wrong userID: got %s, want %s", uid, userID)
					}
					return &service.DashboardStats{
						TotalLinks:  10,
						ActiveLinks: 5,
						TotalClicks: 100,
						RecentLinks: []dto.LinkResponse{},
					}, nil
				},
			},
			logger: createTestLogger(),
		}

		req := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard", nil)
		ctx := middleware.WithUserID(req.Context(), userID)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.GetDashboard(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("GetDashboard() status = %d, want %d", w.Code, http.StatusOK)
		}

		var response dto.SuccessResponse[*service.DashboardStats]
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}
		if response.Data.TotalLinks != 10 {
			t.Errorf("Response Data.TotalLinks = %d, want 10", response.Data.TotalLinks)
		}
		if response.Data.ActiveLinks != 5 {
			t.Errorf("Response Data.ActiveLinks = %d, want 5", response.Data.ActiveLinks)
		}
		if response.Data.TotalClicks != 100 {
			t.Errorf("Response Data.TotalClicks = %d, want 100", response.Data.TotalClicks)
		}
	})

	t.Run("service error", func(t *testing.T) {
		handler := &LinkHandler{
			LinkService: &mockLinkService{
				GetDashboardStatsFunc: func(ctx context.Context, uid string) (*service.DashboardStats, error) {
					return nil, errors.New("database error")
				},
			},
			logger: createTestLogger(),
		}

		req := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard", nil)
		ctx := middleware.WithUserID(req.Context(), userID)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.GetDashboard(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("GetDashboard() status = %d, want %d", w.Code, http.StatusInternalServerError)
		}
	})
}

func TestLinkHandler_GetLinkAnalytics(t *testing.T) {
	userID := "user_123"
	shortcode := "abc123"

	t.Run("successful analytics", func(t *testing.T) {
		handler := &LinkHandler{
			LinkService: &mockLinkService{
				GetLinkAnalyticsFunc: func(ctx context.Context, uid string, code string) (*analytics.LinkAnalytics, error) {
					if uid != userID {
						t.Errorf("GetLinkAnalytics called with wrong userID: got %s, want %s", uid, userID)
					}
					if code != shortcode {
						t.Errorf("GetLinkAnalytics called with wrong shortcode: got %s, want %s", code, shortcode)
					}
					return &analytics.LinkAnalytics{
						TotalClicks:    42,
						ClicksOverTime: []analytics.ClicksOverTime{},
						TopReferrers:   []analytics.ReferrerStat{},
						TopUserAgents:  []analytics.UserAgentStat{},
					}, nil
				},
			},
			logger: createTestLogger(),
		}

		req := httptest.NewRequest(http.MethodGet, "/api/v1/links/"+shortcode+"/analytics", nil)
		ctx := middleware.WithUserID(req.Context(), userID)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		r := chi.NewRouter()
		r.Get("/api/v1/links/{shortcode}/analytics", handler.GetLinkAnalytics)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("GetLinkAnalytics() status = %d, want %d", w.Code, http.StatusOK)
		}

		var response dto.SuccessResponse[*analytics.LinkAnalytics]
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}
		if response.Data.TotalClicks != 42 {
			t.Errorf("Response Data.TotalClicks = %d, want 42", response.Data.TotalClicks)
		}
	})

	t.Run("link not found", func(t *testing.T) {
		handler := &LinkHandler{
			LinkService: &mockLinkService{
				GetLinkAnalyticsFunc: func(ctx context.Context, uid string, code string) (*analytics.LinkAnalytics, error) {
					return nil, apperrors.LinkNotFound
				},
			},
			logger: createTestLogger(),
		}

		req := httptest.NewRequest(http.MethodGet, "/api/v1/links/"+shortcode+"/analytics", nil)
		ctx := middleware.WithUserID(req.Context(), userID)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		r := chi.NewRouter()
		r.Get("/api/v1/links/{shortcode}/analytics", handler.GetLinkAnalytics)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("GetLinkAnalytics() status = %d, want %d", w.Code, http.StatusNotFound)
		}
	})

	t.Run("service error", func(t *testing.T) {
		handler := &LinkHandler{
			LinkService: &mockLinkService{
				GetLinkAnalyticsFunc: func(ctx context.Context, uid string, code string) (*analytics.LinkAnalytics, error) {
					return nil, errors.New("analytics service error")
				},
			},
			logger: createTestLogger(),
		}

		req := httptest.NewRequest(http.MethodGet, "/api/v1/links/"+shortcode+"/analytics", nil)
		ctx := middleware.WithUserID(req.Context(), userID)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		r := chi.NewRouter()
		r.Get("/api/v1/links/{shortcode}/analytics", handler.GetLinkAnalytics)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("GetLinkAnalytics() status = %d, want %d", w.Code, http.StatusInternalServerError)
		}
	})
}

// Note: Error mapping is now tested in pkg/errors/errors_test.go via TestMapError
// The error handling middleware is tested through integration tests
