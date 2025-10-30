package service

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"math/big"

	"github.com/google/uuid"
	"github.com/styltsou/url-shortener/server/pkg/db"
)

func generateRandomCode(n int) (string, error) {
	codeAlphabet := []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	b := make([]byte, n)

	for i := range n {
		// crypto/rand for unpredictability; map to alphabet via modulo bias-free method
		// Use rand.Int with max = len(alphabet)
		idxBig, err := rand.Int(rand.Reader, big.NewInt(int64(len(codeAlphabet))))
		if err != nil {
			return "", err
		}

		b[i] = codeAlphabet[idxBig.Int64()]
	}

	return string(b), nil
}

type LinkService struct {
	queries *db.Queries
}

func NewLinkService(queries *db.Queries) *LinkService {
	return &LinkService{queries: queries}
}

func (s *LinkService) CreateShortLink(ctx context.Context, userID uuid.UUID, originalURL string) (db.Link, error) {
	const (
		codeLen     = 9
		maxAttempts = 3 // 62^7 = 3.5T combinations; collisions are extremely rare
	)

	for range maxAttempts {
		code, err := generateRandomCode(codeLen)
		if err != nil {
			return db.Link{}, fmt.Errorf("failed to generate short code: %w", err)
		}

		link, err := s.queries.TryCreateLink(ctx, db.TryCreateLinkParams{
			Code:        code,
			OriginalUrl: originalURL,
			UserID:      userID,
		})

		if err == nil {
			return link, nil
		}

		// Collision: ON CONFLICT DO NOTHING returned no rows
		// Note: This is the ONLY way sql.ErrNoRows can occur here,
		// since successful inserts always return a row
		if errors.Is(err, sql.ErrNoRows) {
			continue // Generate new code and retry
		}

		// Other database error
		return db.Link{}, fmt.Errorf("failed to create link: %w", err)
	}

	return db.Link{}, fmt.Errorf("failed to create link after %d attempts", maxAttempts)
}

func (s *LinkService) ListAllLinks(ctx context.Context, userID uuid.UUID) ([]db.Link, error) {
	return s.queries.ListUserLinks(ctx, userID)
}

func (s *LinkService) GetLinkByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (db.Link, error) {
	return s.queries.GetLinkByIdAndUser(ctx, db.GetLinkByIdAndUserParams{
		ID:     id,
		UserID: userID,
	})
}

func (s *LinkService) GetOriginalURL(ctx context.Context, code string) (db.GetLinkForRedirectRow, error) {
	return s.queries.GetLinkForRedirect(ctx, code)
}

// TODO Implement the following
// This will stay empty untill i actuall see my use case
func (s *LinkService) UpdateLink(ctx context.Context) {}

func (s *LinkService) DeleteLink(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	return s.queries.DeleteLink(ctx, db.DeleteLinkParams{
		ID:     id,
		UserID: userID,
	})
}
