package service

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/styltsou/url-shortener/server/pkg/db"
	apperrors "github.com/styltsou/url-shortener/server/pkg/errors"
)

type mockTagQueries struct {
	ListUserTagsFunc func(ctx context.Context, userID string) ([]db.ListUserTagsRow, error)
	CreateTagFunc    func(ctx context.Context, arg db.CreateTagParams) (db.CreateTagRow, error)
	UpdateTagFunc    func(ctx context.Context, arg db.UpdateTagParams) (db.UpdateTagRow, error)
	DeleteTagFunc    func(ctx context.Context, arg db.DeleteTagParams) (db.DeleteTagRow, error)
	DeleteTagsFunc   func(ctx context.Context, arg db.DeleteTagsParams) ([]db.DeleteTagsRow, error)
}

func (m *mockTagQueries) ListUserTags(ctx context.Context, userID string) ([]db.ListUserTagsRow, error) {
	if m.ListUserTagsFunc != nil {
		return m.ListUserTagsFunc(ctx, userID)
	}
	return nil, errors.New("not implemented")
}

func (m *mockTagQueries) CreateTag(ctx context.Context, arg db.CreateTagParams) (db.CreateTagRow, error) {
	if m.CreateTagFunc != nil {
		return m.CreateTagFunc(ctx, arg)
	}
	return db.CreateTagRow{}, errors.New("not implemented")
}

func (m *mockTagQueries) UpdateTag(ctx context.Context, arg db.UpdateTagParams) (db.UpdateTagRow, error) {
	if m.UpdateTagFunc != nil {
		return m.UpdateTagFunc(ctx, arg)
	}
	return db.UpdateTagRow{}, errors.New("not implemented")
}

func (m *mockTagQueries) DeleteTag(ctx context.Context, arg db.DeleteTagParams) (db.DeleteTagRow, error) {
	if m.DeleteTagFunc != nil {
		return m.DeleteTagFunc(ctx, arg)
	}
	return db.DeleteTagRow{}, errors.New("not implemented")
}

func (m *mockTagQueries) DeleteTags(ctx context.Context, arg db.DeleteTagsParams) ([]db.DeleteTagsRow, error) {
	if m.DeleteTagsFunc != nil {
		return m.DeleteTagsFunc(ctx, arg)
	}
	return nil, errors.New("not implemented")
}

func createTestTagRow(id uuid.UUID, name string) db.ListUserTagsRow {
	return db.ListUserTagsRow{
		ID:        id,
		Name:      name,
		CreatedAt: pgtype.Timestamp{Valid: false},
		UpdatedAt: pgtype.Timestamp{Valid: false},
	}
}

func createTestCreateTagRow(id uuid.UUID, name string) db.CreateTagRow {
	return db.CreateTagRow{
		ID:        id,
		Name:      name,
		CreatedAt: pgtype.Timestamp{Valid: false},
		UpdatedAt: pgtype.Timestamp{Valid: false},
	}
}

func createTestUpdateTagRow(id uuid.UUID, name string) db.UpdateTagRow {
	return db.UpdateTagRow{
		ID:        id,
		Name:      name,
		CreatedAt: pgtype.Timestamp{Valid: false},
		UpdatedAt: pgtype.Timestamp{Valid: false},
	}
}

func TestTagService_ListAllTags(t *testing.T) {
	ctx := context.Background()
	userID := "user_123"

	t.Run("successful list with tags", func(t *testing.T) {
		expectedTags := []db.ListUserTagsRow{
			createTestTagRow(uuid.New(), "tag1"),
			createTestTagRow(uuid.New(), "tag2"),
		}

		mockQueries := &mockTagQueries{
			ListUserTagsFunc: func(ctx context.Context, gotUserID string) ([]db.ListUserTagsRow, error) {
				if gotUserID != userID {
					t.Errorf("ListUserTags called with wrong userID: got %s, want %s", gotUserID, userID)
				}
				return expectedTags, nil
			},
		}

		service := NewTagService(mockQueries, createTestLogger())
		tags, err := service.ListAllTags(ctx, userID)

		if err != nil {
			t.Errorf("ListAllTags() error = %v, want nil", err)
		}
		if len(tags) != 2 {
			t.Errorf("ListAllTags() length = %d, want 2", len(tags))
		}
	})

	t.Run("successful list with no tags", func(t *testing.T) {
		mockQueries := &mockTagQueries{
			ListUserTagsFunc: func(ctx context.Context, gotUserID string) ([]db.ListUserTagsRow, error) {
				return []db.ListUserTagsRow{}, nil
			},
		}

		service := NewTagService(mockQueries, createTestLogger())
		tags, err := service.ListAllTags(ctx, userID)

		if err != nil {
			t.Errorf("ListAllTags() error = %v, want nil", err)
		}
		if len(tags) != 0 {
			t.Errorf("ListAllTags() length = %d, want 0", len(tags))
		}
	})

	t.Run("handles database errors", func(t *testing.T) {
		dbError := errors.New("database query failed")
		mockQueries := &mockTagQueries{
			ListUserTagsFunc: func(ctx context.Context, gotUserID string) ([]db.ListUserTagsRow, error) {
				return nil, dbError
			},
		}

		service := NewTagService(mockQueries, createTestLogger())
		_, err := service.ListAllTags(ctx, userID)

		if err == nil {
			t.Errorf("ListAllTags() expected error for database failure")
		}
	})
}

func TestTagService_CreateTag(t *testing.T) {
	ctx := context.Background()
	userID := "user_123"
	tagName := "my-tag"

	t.Run("successful creation", func(t *testing.T) {
		mockQueries := &mockTagQueries{
			CreateTagFunc: func(ctx context.Context, arg db.CreateTagParams) (db.CreateTagRow, error) {
				if arg.Name != tagName {
					t.Errorf("CreateTag called with wrong name: got %s, want %s", arg.Name, tagName)
				}
				if arg.UserID != userID {
					t.Errorf("CreateTag called with wrong userID: got %s, want %s", arg.UserID, userID)
				}
				return createTestCreateTagRow(uuid.New(), arg.Name), nil
			},
		}

		service := NewTagService(mockQueries, createTestLogger())
		tag, err := service.CreateTag(ctx, userID, tagName)

		if err != nil {
			t.Errorf("CreateTag() error = %v, want nil", err)
		}
		if tag.Name != tagName {
			t.Errorf("CreateTag() Name = %s, want %s", tag.Name, tagName)
		}
	})

	t.Run("duplicate name returns TagNameTaken", func(t *testing.T) {
		pgErr := &pgconn.PgError{
			Code: apperrors.PgUniqueViolation,
		}

		mockQueries := &mockTagQueries{
			CreateTagFunc: func(ctx context.Context, arg db.CreateTagParams) (db.CreateTagRow, error) {
				return db.CreateTagRow{}, pgErr
			},
		}

		service := NewTagService(mockQueries, createTestLogger())
		_, err := service.CreateTag(ctx, userID, tagName)

		if err == nil {
			t.Errorf("CreateTag() expected error for duplicate name")
		}
		if !errors.Is(err, apperrors.TagNameTaken) {
			t.Errorf("CreateTag() error = %v, want TagNameTaken", err)
		}
	})

	t.Run("handles database errors", func(t *testing.T) {
		dbError := errors.New("database query failed")
		mockQueries := &mockTagQueries{
			CreateTagFunc: func(ctx context.Context, arg db.CreateTagParams) (db.CreateTagRow, error) {
				return db.CreateTagRow{}, dbError
			},
		}

		service := NewTagService(mockQueries, createTestLogger())
		_, err := service.CreateTag(ctx, userID, tagName)

		if err == nil {
			t.Errorf("CreateTag() expected error for database failure")
		}
	})
}

func TestTagService_UpdateTag(t *testing.T) {
	ctx := context.Background()
	userID := "user_123"
	tagID := uuid.New()
	newName := "updated-tag"

	t.Run("successful update", func(t *testing.T) {
		mockQueries := &mockTagQueries{
			UpdateTagFunc: func(ctx context.Context, arg db.UpdateTagParams) (db.UpdateTagRow, error) {
				if arg.ID != tagID {
					t.Errorf("UpdateTag called with wrong ID: got %s, want %s", arg.ID, tagID)
				}
				if arg.UserID != userID {
					t.Errorf("UpdateTag called with wrong userID: got %s, want %s", arg.UserID, userID)
				}
				if arg.Name != newName {
					t.Errorf("UpdateTag called with wrong name: got %s, want %s", arg.Name, newName)
				}
				return createTestUpdateTagRow(tagID, newName), nil
			},
		}

		service := NewTagService(mockQueries, createTestLogger())
		tag, err := service.UpdateTag(ctx, userID, tagID, newName)

		if err != nil {
			t.Errorf("UpdateTag() error = %v, want nil", err)
		}
		if tag.Name != newName {
			t.Errorf("UpdateTag() Name = %s, want %s", tag.Name, newName)
		}
	})

	t.Run("tag not found", func(t *testing.T) {
		mockQueries := &mockTagQueries{
			UpdateTagFunc: func(ctx context.Context, arg db.UpdateTagParams) (db.UpdateTagRow, error) {
				return db.UpdateTagRow{}, sql.ErrNoRows
			},
		}

		service := NewTagService(mockQueries, createTestLogger())
		_, err := service.UpdateTag(ctx, userID, tagID, newName)

		if err == nil {
			t.Errorf("UpdateTag() expected error for not found")
		}
		if !errors.Is(err, apperrors.TagNotFound) {
			t.Errorf("UpdateTag() error = %v, want TagNotFound", err)
		}
	})

	t.Run("duplicate name on update returns TagNameTaken", func(t *testing.T) {
		pgErr := &pgconn.PgError{
			Code: apperrors.PgUniqueViolation,
		}

		mockQueries := &mockTagQueries{
			UpdateTagFunc: func(ctx context.Context, arg db.UpdateTagParams) (db.UpdateTagRow, error) {
				return db.UpdateTagRow{}, pgErr
			},
		}

		service := NewTagService(mockQueries, createTestLogger())
		_, err := service.UpdateTag(ctx, userID, tagID, newName)

		if err == nil {
			t.Errorf("UpdateTag() expected error for duplicate name")
		}
		if !errors.Is(err, apperrors.TagNameTaken) {
			t.Errorf("UpdateTag() error = %v, want TagNameTaken", err)
		}
	})

	t.Run("handles database errors", func(t *testing.T) {
		dbError := errors.New("database query failed")
		mockQueries := &mockTagQueries{
			UpdateTagFunc: func(ctx context.Context, arg db.UpdateTagParams) (db.UpdateTagRow, error) {
				return db.UpdateTagRow{}, dbError
			},
		}

		service := NewTagService(mockQueries, createTestLogger())
		_, err := service.UpdateTag(ctx, userID, tagID, newName)

		if err == nil {
			t.Errorf("UpdateTag() expected error for database failure")
		}
	})
}

func TestTagService_DeleteTag(t *testing.T) {
	ctx := context.Background()
	userID := "user_123"
	tagID := uuid.New()

	t.Run("successful deletion", func(t *testing.T) {
		mockQueries := &mockTagQueries{
			DeleteTagFunc: func(ctx context.Context, arg db.DeleteTagParams) (db.DeleteTagRow, error) {
				if arg.ID != tagID {
					t.Errorf("DeleteTag called with wrong ID: got %s, want %s", arg.ID, tagID)
				}
				if arg.UserID != userID {
					t.Errorf("DeleteTag called with wrong userID: got %s, want %s", arg.UserID, userID)
				}
				return db.DeleteTagRow{
					ID:        tagID,
					Name:      "deleted-tag",
					CreatedAt: pgtype.Timestamp{Valid: false},
					UpdatedAt: pgtype.Timestamp{Valid: false},
				}, nil
			},
		}

		service := NewTagService(mockQueries, createTestLogger())
		deleted, err := service.DeleteTag(ctx, userID, tagID)

		if err != nil {
			t.Errorf("DeleteTag() error = %v, want nil", err)
		}
		if deleted.ID != tagID {
			t.Errorf("DeleteTag() returned wrong ID: got %s, want %s", deleted.ID, tagID)
		}
	})

	t.Run("tag not found", func(t *testing.T) {
		mockQueries := &mockTagQueries{
			DeleteTagFunc: func(ctx context.Context, arg db.DeleteTagParams) (db.DeleteTagRow, error) {
				return db.DeleteTagRow{}, sql.ErrNoRows
			},
		}

		service := NewTagService(mockQueries, createTestLogger())
		_, err := service.DeleteTag(ctx, userID, tagID)

		if err == nil {
			t.Errorf("DeleteTag() expected error for not found")
		}
		if !errors.Is(err, apperrors.TagNotFound) {
			t.Errorf("DeleteTag() error = %v, want TagNotFound", err)
		}
	})

	t.Run("handles database errors", func(t *testing.T) {
		dbError := errors.New("database query failed")
		mockQueries := &mockTagQueries{
			DeleteTagFunc: func(ctx context.Context, arg db.DeleteTagParams) (db.DeleteTagRow, error) {
				return db.DeleteTagRow{}, dbError
			},
		}

		service := NewTagService(mockQueries, createTestLogger())
		_, err := service.DeleteTag(ctx, userID, tagID)

		if err == nil {
			t.Errorf("DeleteTag() expected error for database failure")
		}
	})
}

func TestTagService_DeleteTags(t *testing.T) {
	ctx := context.Background()
	userID := "user_123"
	tagID1 := uuid.New()
	tagID2 := uuid.New()

	t.Run("successful bulk delete", func(t *testing.T) {
		mockQueries := &mockTagQueries{
			DeleteTagsFunc: func(ctx context.Context, arg db.DeleteTagsParams) ([]db.DeleteTagsRow, error) {
				if len(arg.TagIDs) != 2 {
					t.Errorf("DeleteTags called with wrong number of tagIDs: got %d, want 2", len(arg.TagIDs))
				}
				if arg.UserID != userID {
					t.Errorf("DeleteTags called with wrong userID: got %s, want %s", arg.UserID, userID)
				}
				return []db.DeleteTagsRow{
					{ID: tagID1, Name: "tag1", CreatedAt: pgtype.Timestamp{Valid: false}, UpdatedAt: pgtype.Timestamp{Valid: false}},
					{ID: tagID2, Name: "tag2", CreatedAt: pgtype.Timestamp{Valid: false}, UpdatedAt: pgtype.Timestamp{Valid: false}},
				}, nil
			},
		}

		service := NewTagService(mockQueries, createTestLogger())
		deleted, err := service.DeleteTags(ctx, userID, []uuid.UUID{tagID1, tagID2})

		if err != nil {
			t.Errorf("DeleteTags() error = %v, want nil", err)
		}
		if len(deleted) != 2 {
			t.Errorf("DeleteTags() returned %d tags, want 2", len(deleted))
		}
	})

	t.Run("empty ID list returns empty without calling DB", func(t *testing.T) {
		called := false
		mockQueries := &mockTagQueries{
			DeleteTagsFunc: func(ctx context.Context, arg db.DeleteTagsParams) ([]db.DeleteTagsRow, error) {
				called = true
				return nil, nil
			},
		}

		service := NewTagService(mockQueries, createTestLogger())
		deleted, err := service.DeleteTags(ctx, userID, []uuid.UUID{})

		if err != nil {
			t.Errorf("DeleteTags() with empty list error = %v, want nil", err)
		}
		if len(deleted) != 0 {
			t.Errorf("DeleteTags() with empty list returned %d tags, want 0", len(deleted))
		}
		if called {
			t.Errorf("DeleteTags should not call database with empty ID list")
		}
	})

	t.Run("handles database errors", func(t *testing.T) {
		dbError := errors.New("database query failed")
		mockQueries := &mockTagQueries{
			DeleteTagsFunc: func(ctx context.Context, arg db.DeleteTagsParams) ([]db.DeleteTagsRow, error) {
				return nil, dbError
			},
		}

		service := NewTagService(mockQueries, createTestLogger())
		_, err := service.DeleteTags(ctx, userID, []uuid.UUID{tagID1})

		if err == nil {
			t.Errorf("DeleteTags() expected error for database failure")
		}
	})
}
