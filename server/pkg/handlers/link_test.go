package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

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
	CreateShortLinkFunc func(ctx context.Context, userID string, originalURL string) (db.Link, error)
	ListAllLinksFunc    func(ctx context.Context, userID string) ([]db.Link, error)
	GetLinkByIDFunc     func(ctx context.Context, id uuid.UUID, userID string) (db.Link, error)
	GetOriginalURLFunc  func(ctx context.Context, code string) (db.GetLinkForRedirectRow, error)
	DeleteLinkFunc      func(ctx context.Context, id uuid.UUID, userID string) error
}

func (m *mockLinkService) CreateShortLink(ctx context.Context, userID string, originalURL string) (db.Link, error) {
	if m.CreateShortLinkFunc != nil {
		return m.CreateShortLinkFunc(ctx, userID, originalURL)
	}
	return db.Link{}, errors.New("not implemented")
}

func (m *mockLinkService) ListAllLinks(ctx context.Context, userID string) ([]db.Link, error) {
	if m.ListAllLinksFunc != nil {
		return m.ListAllLinksFunc(ctx, userID)
	}
	return nil, errors.New("not implemented")
}

func (m *mockLinkService) GetLinkByID(ctx context.Context, id uuid.UUID, userID string) (db.Link, error) {
	if m.GetLinkByIDFunc != nil {
		return m.GetLinkByIDFunc(ctx, id, userID)
	}
	return db.Link{}, errors.New("not implemented")
}

func (m *mockLinkService) GetOriginalURL(ctx context.Context, code string) (db.GetLinkForRedirectRow, error) {
	if m.GetOriginalURLFunc != nil {
		return m.GetOriginalURLFunc(ctx, code)
	}
	return db.GetLinkForRedirectRow{}, errors.New("not implemented")
}

func (m *mockLinkService) DeleteLink(ctx context.Context, id uuid.UUID, userID string) error {
	if m.DeleteLinkFunc != nil {
		return m.DeleteLinkFunc(ctx, id, userID)
	}
	return errors.New("not implemented")
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
		Clicks:      nil,
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
				CreateShortLinkFunc: func(ctx context.Context, userID string, originalURL string) (db.Link, error) {
					if userID != "user_123" {
						t.Errorf("CreateShortLink called with wrong userID: got %s, want user_123", userID)
					}
					if originalURL != "https://example.com" {
						t.Errorf("CreateShortLink called with wrong URL: got %s, want https://example.com", originalURL)
					}
					return createTestLink(uuid.New(), "abc123", originalURL, userID), nil
				},
			},
			expectedStatus: http.StatusCreated,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response dto.SuccessResponse[db.Link]
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if response.Data.OriginalUrl != "https://example.com" {
					t.Errorf("Response OriginalUrl = %s, want https://example.com", response.Data.OriginalUrl)
				}
				if response.Message != "Short Link created successfully" {
					t.Errorf("Response Message = %s, want 'Short Link created successfully'", response.Message)
				}
			},
		},
		{
			name: "invalid JSON body",
			requestBody: dto.CreateLink{
				URL: "https://example.com",
			},
			userID:         "user_123",
			mockService:    &mockLinkService{},
			expectedStatus: http.StatusBadRequest,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response dto.ErrorResponse
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				// Generic HTTP error - no code field (omitted)
				if response.Error.Code != "" {
					t.Errorf("Response Error.Code = %s, want empty (generic HTTP error)", response.Error.Code)
				}
			},
		},
		{
			name: "invalid URL",
			requestBody: dto.CreateLink{
				URL: "invalid-url",
			},
			userID: "user_123",
			mockService: &mockLinkService{
				CreateShortLinkFunc: func(ctx context.Context, userID string, originalURL string) (db.Link, error) {
					return db.Link{}, apperrors.InvalidURL
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
				CreateShortLinkFunc: func(ctx context.Context, userID string, originalURL string) (db.Link, error) {
					return db.Link{}, errors.New("database error")
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
				ListAllLinksFunc: func(ctx context.Context, userID string) ([]db.Link, error) {
					if userID != "user_123" {
						t.Errorf("ListAllLinks called with wrong userID: got %s, want user_123", userID)
					}
					return []db.Link{
						createTestLink(uuid.New(), "abc123", "https://example.com/1", userID),
						createTestLink(uuid.New(), "def456", "https://example.com/2", userID),
					}, nil
				},
			},
			expectedStatus: http.StatusOK,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response dto.SuccessResponse[[]db.Link]
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
				ListAllLinksFunc: func(ctx context.Context, userID string) ([]db.Link, error) {
					return []db.Link{}, nil
				},
			},
			expectedStatus: http.StatusOK,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response dto.SuccessResponse[[]db.Link]
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
				ListAllLinksFunc: func(ctx context.Context, userID string) ([]db.Link, error) {
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

// Note: Error mapping is now tested in pkg/errors/errors_test.go via TestMapError
// The error handling middleware is tested through integration tests
