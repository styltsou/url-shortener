package service

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/styltsou/url-shortener/server/pkg/db"
	apperrors "github.com/styltsou/url-shortener/server/pkg/errors"
)

func TestCustomShortcodeValidationAndConflict(t *testing.T) {
	for _, code := range []string{"", "api", "API", "a/b", "a b", "é", "123456789012345678901"} {
		t.Run(code, func(t *testing.T) {
			s := &LinkService{queries: &mockQueries{}}
			_, err := s.CreateShortLink(context.Background(), "user", "https://example.com", &code, nil)
			if !errors.Is(err, apperrors.InvalidShortcode) {
				t.Fatalf("code %q: got %v", code, err)
			}
		})
	}
	for _, code := range []string{"a", "a-b_C9", "12345678901234567890"} {
		if !validShortcode(code) {
			t.Fatalf("valid shortcode %q rejected", code)
		}
	}
	for _, conflict := range []error{sql.ErrNoRows, &pgconn.PgError{Code: apperrors.PgUniqueViolation, ConstraintName: "idx_links_shortcode"}} {
		code := "taken"
		s := &LinkService{queries: &mockQueries{TryCreateLinkFunc: func(context.Context, db.TryCreateLinkParams) (db.TryCreateLinkRow, error) {
			return db.TryCreateLinkRow{}, conflict
		}}}
		_, err := s.CreateShortLink(context.Background(), "user", "https://example.com", &code, nil)
		if !errors.Is(err, apperrors.LinkShortcodeTaken) {
			t.Fatalf("conflict %v mapped to %v", conflict, err)
		}
	}
	code := "fine"
	s := &LinkService{queries: &mockQueries{TryCreateLinkFunc: func(context.Context, db.TryCreateLinkParams) (db.TryCreateLinkRow, error) {
		return db.TryCreateLinkRow{}, &pgconn.PgError{Code: apperrors.PgUniqueViolation, ConstraintName: "other_unique_index"}
	}}}
	_, err := s.CreateShortLink(context.Background(), "user", "https://example.com", &code, nil)
	if errors.Is(err, apperrors.LinkShortcodeTaken) {
		t.Fatal("unrelated unique index was treated as shortcode conflict")
	}
}

func TestGeneratedCodeCollisionRetries(t *testing.T) {
	var attempts []string
	s := &LinkService{
		queries: &mockQueries{TryCreateLinkFunc: func(_ context.Context, arg db.TryCreateLinkParams) (db.TryCreateLinkRow, error) {
			attempts = append(attempts, arg.Shortcode)
			if len(attempts) == 1 {
				return db.TryCreateLinkRow{}, sql.ErrNoRows
			}
			return db.TryCreateLinkRow{ID: uuid.New(), Shortcode: arg.Shortcode}, nil
		}},
		generateShortcode: func(_, _ string, attempt int) (string, error) {
			if attempt == 0 {
				return "first", nil
			}
			return "second", nil
		},
	}
	link, err := s.CreateShortLink(context.Background(), "user", "https://example.com", nil, nil)
	if err != nil || link.Shortcode != "second" || len(attempts) != 2 {
		t.Fatalf("retry failed: %+v %v attempts=%v", link, err, attempts)
	}
}
