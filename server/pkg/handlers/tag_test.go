package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/styltsou/url-shortener/server/pkg/db"
	"github.com/styltsou/url-shortener/server/pkg/dto"
	apperrors "github.com/styltsou/url-shortener/server/pkg/errors"
	"github.com/styltsou/url-shortener/server/pkg/middleware"
)

type mockTagService struct {
	ListAllTagsFunc func(ctx context.Context, userID string) ([]db.ListUserTagsRow, error)
	CreateTagFunc   func(ctx context.Context, userID string, name string) (db.CreateTagRow, error)
	UpdateTagFunc   func(ctx context.Context, userID string, tagID uuid.UUID, name string) (db.UpdateTagRow, error)
	DeleteTagFunc   func(ctx context.Context, userID string, tagID uuid.UUID) (db.DeleteTagRow, error)
	DeleteTagsFunc  func(ctx context.Context, userID string, tagIDs []uuid.UUID) ([]db.DeleteTagsRow, error)
}

func (m *mockTagService) ListAllTags(ctx context.Context, userID string) ([]db.ListUserTagsRow, error) {
	if m.ListAllTagsFunc != nil {
		return m.ListAllTagsFunc(ctx, userID)
	}
	return nil, errors.New("not implemented")
}

func (m *mockTagService) CreateTag(ctx context.Context, userID string, name string) (db.CreateTagRow, error) {
	if m.CreateTagFunc != nil {
		return m.CreateTagFunc(ctx, userID, name)
	}
	return db.CreateTagRow{}, errors.New("not implemented")
}

func (m *mockTagService) UpdateTag(ctx context.Context, userID string, tagID uuid.UUID, name string) (db.UpdateTagRow, error) {
	if m.UpdateTagFunc != nil {
		return m.UpdateTagFunc(ctx, userID, tagID, name)
	}
	return db.UpdateTagRow{}, errors.New("not implemented")
}

func (m *mockTagService) DeleteTag(ctx context.Context, userID string, tagID uuid.UUID) (db.DeleteTagRow, error) {
	if m.DeleteTagFunc != nil {
		return m.DeleteTagFunc(ctx, userID, tagID)
	}
	return db.DeleteTagRow{}, errors.New("not implemented")
}

func (m *mockTagService) DeleteTags(ctx context.Context, userID string, tagIDs []uuid.UUID) ([]db.DeleteTagsRow, error) {
	if m.DeleteTagsFunc != nil {
		return m.DeleteTagsFunc(ctx, userID, tagIDs)
	}
	return nil, errors.New("not implemented")
}

func TestTagHandler_ListTags(t *testing.T) {
	userID := "user_123"

	t.Run("successful list", func(t *testing.T) {
		handler := &TagHandler{
			TagService: &mockTagService{
				ListAllTagsFunc: func(ctx context.Context, uid string) ([]db.ListUserTagsRow, error) {
					if uid != userID {
						t.Errorf("ListAllTags called with wrong userID: got %s, want %s", uid, userID)
					}
					return []db.ListUserTagsRow{
						{ID: uuid.New(), Name: "tag1", CreatedAt: pgtype.Timestamp{Valid: false}, UpdatedAt: pgtype.Timestamp{Valid: false}},
						{ID: uuid.New(), Name: "tag2", CreatedAt: pgtype.Timestamp{Valid: false}, UpdatedAt: pgtype.Timestamp{Valid: false}},
					}, nil
				},
			},
			logger: createTestLogger(),
		}

		req := httptest.NewRequest(http.MethodGet, "/api/v1/tags", nil)
		ctx := middleware.WithUserID(req.Context(), userID)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.ListTags(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("ListTags() status = %d, want %d", w.Code, http.StatusOK)
		}

		var response dto.SuccessResponse[[]db.ListUserTagsRow]
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}
		if len(response.Data) != 2 {
			t.Errorf("Response Data length = %d, want 2", len(response.Data))
		}
	})

	t.Run("service error", func(t *testing.T) {
		handler := &TagHandler{
			TagService: &mockTagService{
				ListAllTagsFunc: func(ctx context.Context, uid string) ([]db.ListUserTagsRow, error) {
					return nil, errors.New("database error")
				},
			},
			logger: createTestLogger(),
		}

		req := httptest.NewRequest(http.MethodGet, "/api/v1/tags", nil)
		ctx := middleware.WithUserID(req.Context(), userID)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.ListTags(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("ListTags() status = %d, want %d", w.Code, http.StatusInternalServerError)
		}
	})
}

func TestTagHandler_CreateTag(t *testing.T) {
	userID := "user_123"

	t.Run("successful creation", func(t *testing.T) {
		handler := &TagHandler{
			TagService: &mockTagService{
				CreateTagFunc: func(ctx context.Context, uid string, name string) (db.CreateTagRow, error) {
					if uid != userID {
						t.Errorf("CreateTag called with wrong userID: got %s, want %s", uid, userID)
					}
					if name != "my-tag" {
						t.Errorf("CreateTag called with wrong name: got %s, want my-tag", name)
					}
					return db.CreateTagRow{
						ID:        uuid.New(),
						Name:      name,
						CreatedAt: pgtype.Timestamp{Valid: false},
						UpdatedAt: pgtype.Timestamp{Valid: false},
					}, nil
				},
			},
			logger: createTestLogger(),
		}

		body := dto.CreateTag{Name: "my-tag"}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/tags", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		ctx := middleware.WithUserID(req.Context(), userID)
		ctx = middleware.CtxWithRequestBody(ctx, body)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.CreateTag(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("CreateTag() status = %d, want %d", w.Code, http.StatusCreated)
		}

		var response dto.SuccessResponse[db.CreateTagRow]
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}
		if response.Data.Name != "my-tag" {
			t.Errorf("Response Data.Name = %s, want my-tag", response.Data.Name)
		}
	})

	t.Run("name already taken", func(t *testing.T) {
		handler := &TagHandler{
			TagService: &mockTagService{
				CreateTagFunc: func(ctx context.Context, uid string, name string) (db.CreateTagRow, error) {
					return db.CreateTagRow{}, apperrors.TagNameTaken
				},
			},
			logger: createTestLogger(),
		}

		body := dto.CreateTag{Name: "taken-tag"}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/tags", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		ctx := middleware.WithUserID(req.Context(), userID)
		ctx = middleware.CtxWithRequestBody(ctx, body)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.CreateTag(w, req)

		if w.Code != http.StatusConflict {
			t.Errorf("CreateTag() status = %d, want %d", w.Code, http.StatusConflict)
		}

		var response dto.ErrorResponse
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}
		if response.Error.Code != apperrors.CodeTagNameTaken {
			t.Errorf("Response Error.Code = %s, want %s", response.Error.Code, apperrors.CodeTagNameTaken)
		}
	})

	t.Run("service error", func(t *testing.T) {
		handler := &TagHandler{
			TagService: &mockTagService{
				CreateTagFunc: func(ctx context.Context, uid string, name string) (db.CreateTagRow, error) {
					return db.CreateTagRow{}, errors.New("database error")
				},
			},
			logger: createTestLogger(),
		}

		body := dto.CreateTag{Name: "my-tag"}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/tags", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		ctx := middleware.WithUserID(req.Context(), userID)
		ctx = middleware.CtxWithRequestBody(ctx, body)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.CreateTag(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("CreateTag() status = %d, want %d", w.Code, http.StatusInternalServerError)
		}
	})
}

func TestTagHandler_UpdateTag(t *testing.T) {
	userID := "user_123"
	tagID := uuid.New()

	t.Run("successful update", func(t *testing.T) {
		handler := &TagHandler{
			TagService: &mockTagService{
				UpdateTagFunc: func(ctx context.Context, uid string, id uuid.UUID, name string) (db.UpdateTagRow, error) {
					if id != tagID {
						t.Errorf("UpdateTag called with wrong ID: got %s, want %s", id, tagID)
					}
					if uid != userID {
						t.Errorf("UpdateTag called with wrong userID: got %s, want %s", uid, userID)
					}
					if name != "updated-name" {
						t.Errorf("UpdateTag called with wrong name: got %s, want updated-name", name)
					}
					return db.UpdateTagRow{
						ID:        tagID,
						Name:      name,
						CreatedAt: pgtype.Timestamp{Valid: false},
						UpdatedAt: pgtype.Timestamp{Valid: false},
					}, nil
				},
			},
			logger: createTestLogger(),
		}

		body := dto.UpdateTag{Name: "updated-name"}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/tags/"+tagID.String(), bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		ctx := middleware.WithUserID(req.Context(), userID)
		ctx = middleware.CtxWithRequestBody(ctx, body)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		r := chi.NewRouter()
		r.Patch("/api/v1/tags/{id}", handler.UpdateTag)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("UpdateTag() status = %d, want %d", w.Code, http.StatusOK)
		}

		var response dto.SuccessResponse[db.UpdateTagRow]
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}
		if response.Data.Name != "updated-name" {
			t.Errorf("Response Data.Name = %s, want updated-name", response.Data.Name)
		}
	})

	t.Run("invalid UUID format", func(t *testing.T) {
		handler := &TagHandler{
			TagService: &mockTagService{},
			logger:     createTestLogger(),
		}

		body := dto.UpdateTag{Name: "test"}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/tags/invalid-uuid", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		ctx := middleware.WithUserID(req.Context(), userID)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		r := chi.NewRouter()
		r.Patch("/api/v1/tags/{id}", handler.UpdateTag)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("UpdateTag() status = %d, want %d", w.Code, http.StatusBadRequest)
		}

		var response dto.ErrorResponse
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}
		if response.Error.Code != apperrors.CodeInvalidID {
			t.Errorf("Response Error.Code = %s, want %s", response.Error.Code, apperrors.CodeInvalidID)
		}
	})

	t.Run("tag not found", func(t *testing.T) {
		handler := &TagHandler{
			TagService: &mockTagService{
				UpdateTagFunc: func(ctx context.Context, uid string, id uuid.UUID, name string) (db.UpdateTagRow, error) {
					return db.UpdateTagRow{}, apperrors.TagNotFound
				},
			},
			logger: createTestLogger(),
		}

		body := dto.UpdateTag{Name: "updated-name"}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/tags/"+tagID.String(), bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		ctx := middleware.WithUserID(req.Context(), userID)
		ctx = middleware.CtxWithRequestBody(ctx, body)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		r := chi.NewRouter()
		r.Patch("/api/v1/tags/{id}", handler.UpdateTag)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("UpdateTag() status = %d, want %d", w.Code, http.StatusNotFound)
		}
	})

	t.Run("name already taken", func(t *testing.T) {
		handler := &TagHandler{
			TagService: &mockTagService{
				UpdateTagFunc: func(ctx context.Context, uid string, id uuid.UUID, name string) (db.UpdateTagRow, error) {
					return db.UpdateTagRow{}, apperrors.TagNameTaken
				},
			},
			logger: createTestLogger(),
		}

		body := dto.UpdateTag{Name: "taken-name"}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/tags/"+tagID.String(), bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		ctx := middleware.WithUserID(req.Context(), userID)
		ctx = middleware.CtxWithRequestBody(ctx, body)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		r := chi.NewRouter()
		r.Patch("/api/v1/tags/{id}", handler.UpdateTag)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusConflict {
			t.Errorf("UpdateTag() status = %d, want %d", w.Code, http.StatusConflict)
		}

		var response dto.ErrorResponse
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}
		if response.Error.Code != apperrors.CodeTagNameTaken {
			t.Errorf("Response Error.Code = %s, want %s", response.Error.Code, apperrors.CodeTagNameTaken)
		}
	})

	t.Run("service error", func(t *testing.T) {
		handler := &TagHandler{
			TagService: &mockTagService{
				UpdateTagFunc: func(ctx context.Context, uid string, id uuid.UUID, name string) (db.UpdateTagRow, error) {
					return db.UpdateTagRow{}, errors.New("database error")
				},
			},
			logger: createTestLogger(),
		}

		body := dto.UpdateTag{Name: "updated-name"}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/tags/"+tagID.String(), bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		ctx := middleware.WithUserID(req.Context(), userID)
		ctx = middleware.CtxWithRequestBody(ctx, body)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		r := chi.NewRouter()
		r.Patch("/api/v1/tags/{id}", handler.UpdateTag)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("UpdateTag() status = %d, want %d", w.Code, http.StatusInternalServerError)
		}
	})
}

func TestTagHandler_DeleteTag(t *testing.T) {
	userID := "user_123"
	tagID := uuid.New()

	t.Run("successful deletion", func(t *testing.T) {
		handler := &TagHandler{
			TagService: &mockTagService{
				DeleteTagFunc: func(ctx context.Context, uid string, id uuid.UUID) (db.DeleteTagRow, error) {
					if id != tagID {
						t.Errorf("DeleteTag called with wrong ID: got %s, want %s", id, tagID)
					}
					if uid != userID {
						t.Errorf("DeleteTag called with wrong userID: got %s, want %s", uid, userID)
					}
					return db.DeleteTagRow{
						ID:        tagID,
						Name:      "deleted-tag",
						CreatedAt: pgtype.Timestamp{Valid: false},
						UpdatedAt: pgtype.Timestamp{Valid: false},
					}, nil
				},
			},
			logger: createTestLogger(),
		}

		req := httptest.NewRequest(http.MethodDelete, "/api/v1/tags/"+tagID.String(), nil)
		ctx := middleware.WithUserID(req.Context(), userID)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		r := chi.NewRouter()
		r.Delete("/api/v1/tags/{id}", handler.DeleteTag)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("DeleteTag() status = %d, want %d", w.Code, http.StatusOK)
		}

		var response dto.SuccessResponse[db.DeleteTagRow]
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}
		if response.Data.ID != tagID {
			t.Errorf("Response Data.ID = %s, want %s", response.Data.ID, tagID)
		}
	})

	t.Run("invalid UUID format", func(t *testing.T) {
		handler := &TagHandler{
			TagService: &mockTagService{},
			logger:     createTestLogger(),
		}

		req := httptest.NewRequest(http.MethodDelete, "/api/v1/tags/invalid-uuid", nil)
		ctx := middleware.WithUserID(req.Context(), userID)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		r := chi.NewRouter()
		r.Delete("/api/v1/tags/{id}", handler.DeleteTag)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("DeleteTag() status = %d, want %d", w.Code, http.StatusBadRequest)
		}

		var response dto.ErrorResponse
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}
		if response.Error.Code != apperrors.CodeInvalidID {
			t.Errorf("Response Error.Code = %s, want %s", response.Error.Code, apperrors.CodeInvalidID)
		}
	})

	t.Run("tag not found", func(t *testing.T) {
		handler := &TagHandler{
			TagService: &mockTagService{
				DeleteTagFunc: func(ctx context.Context, uid string, id uuid.UUID) (db.DeleteTagRow, error) {
					return db.DeleteTagRow{}, apperrors.TagNotFound
				},
			},
			logger: createTestLogger(),
		}

		req := httptest.NewRequest(http.MethodDelete, "/api/v1/tags/"+tagID.String(), nil)
		ctx := middleware.WithUserID(req.Context(), userID)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		r := chi.NewRouter()
		r.Delete("/api/v1/tags/{id}", handler.DeleteTag)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("DeleteTag() status = %d, want %d", w.Code, http.StatusNotFound)
		}
	})

	t.Run("service error", func(t *testing.T) {
		handler := &TagHandler{
			TagService: &mockTagService{
				DeleteTagFunc: func(ctx context.Context, uid string, id uuid.UUID) (db.DeleteTagRow, error) {
					return db.DeleteTagRow{}, errors.New("database error")
				},
			},
			logger: createTestLogger(),
		}

		req := httptest.NewRequest(http.MethodDelete, "/api/v1/tags/"+tagID.String(), nil)
		ctx := middleware.WithUserID(req.Context(), userID)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		r := chi.NewRouter()
		r.Delete("/api/v1/tags/{id}", handler.DeleteTag)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("DeleteTag() status = %d, want %d", w.Code, http.StatusInternalServerError)
		}
	})
}

func TestTagHandler_DeleteTags(t *testing.T) {
	userID := "user_123"
	tagID1 := uuid.New()
	tagID2 := uuid.New()

	t.Run("successful bulk delete", func(t *testing.T) {
		handler := &TagHandler{
			TagService: &mockTagService{
				DeleteTagsFunc: func(ctx context.Context, uid string, tagIDs []uuid.UUID) ([]db.DeleteTagsRow, error) {
					if uid != userID {
						t.Errorf("DeleteTags called with wrong userID: got %s, want %s", uid, userID)
					}
					if len(tagIDs) != 2 {
						t.Errorf("DeleteTags called with %d tagIDs, want 2", len(tagIDs))
					}
					return []db.DeleteTagsRow{
						{ID: tagID1, Name: "tag1", CreatedAt: pgtype.Timestamp{Valid: false}, UpdatedAt: pgtype.Timestamp{Valid: false}},
						{ID: tagID2, Name: "tag2", CreatedAt: pgtype.Timestamp{Valid: false}, UpdatedAt: pgtype.Timestamp{Valid: false}},
					}, nil
				},
			},
			logger: createTestLogger(),
		}

		body := dto.DeleteTags{TagIDs: []uuid.UUID{tagID1, tagID2}}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/tags/bulk-delete", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		ctx := middleware.WithUserID(req.Context(), userID)
		ctx = middleware.CtxWithRequestBody(ctx, body)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.DeleteTags(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("DeleteTags() status = %d, want %d", w.Code, http.StatusOK)
		}

		var response dto.SuccessResponse[[]db.DeleteTagsRow]
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}
		if len(response.Data) != 2 {
			t.Errorf("Response Data length = %d, want 2", len(response.Data))
		}
	})

	t.Run("service error", func(t *testing.T) {
		handler := &TagHandler{
			TagService: &mockTagService{
				DeleteTagsFunc: func(ctx context.Context, uid string, tagIDs []uuid.UUID) ([]db.DeleteTagsRow, error) {
					return nil, errors.New("database error")
				},
			},
			logger: createTestLogger(),
		}

		body := dto.DeleteTags{TagIDs: []uuid.UUID{tagID1}}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/tags/bulk-delete", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		ctx := middleware.WithUserID(req.Context(), userID)
		ctx = middleware.CtxWithRequestBody(ctx, body)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.DeleteTags(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("DeleteTags() status = %d, want %d", w.Code, http.StatusInternalServerError)
		}
	})
}
