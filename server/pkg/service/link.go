package service

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"math/big"
	"net/url"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/styltsou/url-shortener/server/pkg/db"
	apperrors "github.com/styltsou/url-shortener/server/pkg/errors"
	"github.com/styltsou/url-shortener/server/pkg/logger"
	"go.uber.org/zap"
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

// TODO: Might need  to define an interface for RedisCache

// LinkQueries defines the database operations needed by LinkService
type LinkQueries interface {
	TryCreateLink(ctx context.Context, arg db.TryCreateLinkParams) (db.Link, error)
	ListUserLinks(ctx context.Context, userID string) ([]db.Link, error)
	GetLinkByIdAndUser(ctx context.Context, arg db.GetLinkByIdAndUserParams) (db.Link, error)
	GetLinkForRedirect(ctx context.Context, shortcode string) (db.GetLinkForRedirectRow, error)
	DeleteLink(ctx context.Context, arg db.DeleteLinkParams) (int64, error)
}

type LinkService struct {
	queries LinkQueries
	cache   *redis.Client
	logger  logger.Logger
}

func NewLinkService(queries LinkQueries, cache *redis.Client, logger logger.Logger) *LinkService {
	return &LinkService{
		queries: queries,
		cache:   cache,
		logger:  logger,
	}
}

func (s *LinkService) CreateShortLink(ctx context.Context, userID string, originalURL string) (db.Link, error) {
	// Validate URL - return sentinel error that handlers will map
	if err := validateURL(originalURL); err != nil {
		return db.Link{}, err
	}

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
			Shortcode:   code,
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

		// Other database error - wrap with context
		return db.Link{}, fmt.Errorf("failed to create link: %w", err)
	}

	return db.Link{}, fmt.Errorf("failed to create link after %d attempts: %w", maxAttempts, fmt.Errorf("code collision retry limit exceeded"))
}

// validateURL validates that the URL is well-formed and uses http/https
// Returns sentinel error ErrInvalidURL that handlers will map to HTTP response
func validateURL(rawURL string) error {
	if rawURL == "" {
		return fmt.Errorf("%w: URL is required", apperrors.InvalidURL)
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("%w: %v", apperrors.InvalidURL, err)
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("%w: URL must use http or https scheme", apperrors.InvalidURL)
	}

	if parsedURL.Host == "" {
		return fmt.Errorf("%w: URL must have a valid host", apperrors.InvalidURL)
	}

	if len(rawURL) > 2048 {
		return fmt.Errorf("%w: URL is too long (max 2048 characters)", apperrors.InvalidURL)
	}

	return nil
}

func (s *LinkService) ListAllLinks(ctx context.Context, userID string) ([]db.Link, error) {
	s.logger.Debug("Querying database for user links",
		zap.String("user_id", userID),
	)

	links, err := s.queries.ListUserLinks(ctx, userID)
	if err != nil {
		s.logger.Error("Database query failed for ListUserLinks",
			zap.Error(err),
			zap.String("user_id", userID),
		)
		return nil, err
	}

	s.logger.Debug("Database query completed for ListUserLinks",
		zap.String("user_id", userID),
		zap.Int("links_found", len(links)),
	)

	return links, nil
}

func (s *LinkService) GetLinkByID(ctx context.Context, id uuid.UUID, userID string) (db.Link, error) {
	link, err := s.queries.GetLinkByIdAndUser(ctx, db.GetLinkByIdAndUserParams{
		ID:     id,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return db.Link{}, fmt.Errorf("%w: %v", apperrors.LinkNotFound, err)
		}
		return db.Link{}, fmt.Errorf("failed to get link: %w", err)
	}
	return link, nil
}

// TODO: here we check the cache first and if we dont find it, we query the db and then save it to cache
func (s *LinkService) GetOriginalURL(ctx context.Context, code string) (db.GetLinkForRedirectRow, error) {
	link, err := s.queries.GetLinkForRedirect(ctx, code)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return db.GetLinkForRedirectRow{}, fmt.Errorf("%w: code %s", apperrors.LinkNotFound, code)
		}
		return db.GetLinkForRedirectRow{}, fmt.Errorf("failed to get link: %w", err)
	}
	return link, nil
}

// TODO Implement the following
// I should invalidate the cache here
// This will stay empty untill i actuall see my use case
func (s *LinkService) UpdateLink(ctx context.Context) {}

// TODO: I should invalidate the cache here too
func (s *LinkService) DeleteLink(ctx context.Context, id uuid.UUID, userID string) error {
	rowsAffected, err := s.queries.DeleteLink(ctx, db.DeleteLinkParams{
		ID:     id,
		UserID: userID,
	})

	if err != nil {
		return fmt.Errorf("failed to delete link: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("%w: link not found or already deleted", apperrors.LinkNotFound)
	}

	return nil
}
