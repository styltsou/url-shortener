package service

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"math/big"
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
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

type LinkQueries interface {
	TryCreateLink(ctx context.Context, arg db.TryCreateLinkParams) (db.TryCreateLinkRow, error)
	GetLinkForRedirect(ctx context.Context, shortcode string) (db.GetLinkForRedirectRow, error)
	ListUserLinks(ctx context.Context, userID string) ([]db.ListUserLinksRow, error)
	GetLinkByIdAndUser(ctx context.Context, arg db.GetLinkByIdAndUserParams) (db.GetLinkByIdAndUserRow, error)
	GetLinkByShortcodeAndUser(ctx context.Context, arg db.GetLinkByShortcodeAndUserParams) (db.GetLinkByShortcodeAndUserRow, error)
	UpdateLink(ctx context.Context, arg db.UpdateLinkParams) (db.UpdateLinkRow, error)
	DeleteLink(ctx context.Context, arg db.DeleteLinkParams) (db.DeleteLinkRow, error)
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

func (s *LinkService) CreateShortLink(ctx context.Context, userID string, originalURL string) (db.TryCreateLinkRow, error) {
	// Validate URL - return sentinel error that handlers will map
	if err := validateURL(originalURL); err != nil {
		return db.TryCreateLinkRow{}, err
	}

	// NOTE: When custom shortcode support is added (via DTO or separate method):
	// - Custom shortcode conflicts: Return error to user (don't retry)
	// - Auto-generated conflicts: Retry internally (current behavior - correct)
	// This will require checking if shortcode was user-provided vs auto-generated

	const (
		codeLen     = 9
		maxAttempts = 3 // 62^7 = 3.5T combinations; collisions are extremely rare
	)

	for range maxAttempts {
		code, err := generateRandomCode(codeLen)
		if err != nil {
			return db.TryCreateLinkRow{},
				fmt.Errorf("failed to generate short code: %w", err)
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
		return db.TryCreateLinkRow{},
			fmt.Errorf("failed to create link: %w", err)
	}

	return db.TryCreateLinkRow{},
		fmt.Errorf("failed to create link after %d attempts: %w", maxAttempts, fmt.Errorf("code collision retry limit exceeded"))
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

func (s *LinkService) ListAllLinks(ctx context.Context, userID string) ([]db.ListUserLinksRow, error) {
	s.logger.Debug("Querying database for user links",
		zap.String("user_id", userID),
	)

	links, err := s.queries.ListUserLinks(ctx, userID)
	if err != nil {
		s.logger.Error("Database query failed for ListUserLinks",
			zap.Error(err),
			zap.String("user_id", userID),
		)
		return nil, fmt.Errorf("failed to get link: %w", err)
	}

	s.logger.Debug("Database query completed for ListUserLinks",
		zap.String("user_id", userID),
		zap.Int("links_found", len(links)),
	)

	return links, nil
}

func (s *LinkService) GetLinkByShortcode(ctx context.Context, userID string, shortcode string) (db.GetLinkByShortcodeAndUserRow, error) {
	link, err := s.queries.GetLinkByShortcodeAndUser(ctx, db.GetLinkByShortcodeAndUserParams{
		Shortcode: shortcode,
		UserID:    userID,
	})

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return db.GetLinkByShortcodeAndUserRow{},
				fmt.Errorf("%w: %v", apperrors.LinkNotFound, err)
		}
		return db.GetLinkByShortcodeAndUserRow{},
			fmt.Errorf("failed to get link: %w", err)
	}

	return link, nil
}

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

func (s *LinkService) UpdateLink(
	ctx context.Context,
	userID string,
	id uuid.UUID,
	shortcode *string,
	isActive *bool,
	expiresAt *time.Time,
) (db.UpdateLinkRow, error) {

	var expiresAtTimestamp pgtype.Timestamp
	if expiresAt != nil {
		expiresAtTimestamp = pgtype.Timestamp{
			Time:  *expiresAt,
			Valid: true,
		}
	} else {
		expiresAtTimestamp = pgtype.Timestamp{Valid: false}
	}

	updatedLink, err := s.queries.UpdateLink(ctx, db.UpdateLinkParams{
		UserID:    userID,
		ID:        id,
		Shortcode: shortcode,
		IsActive:  isActive,
		ExpiresAt: expiresAtTimestamp,
	})

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return db.UpdateLinkRow{},
				fmt.Errorf("%w: %v", apperrors.LinkNotFound, err)
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			// NOTE: When is_alias field is added, differentiate between:
			// - Custom alias conflicts: Return error to user (current behavior - correct)
			// - Auto-generated conflicts: Should retry internally (not applicable for updates)
			// For now, all conflicts are treated as user-provided custom aliases

			shortcodeStr := "n/a"
			if shortcode != nil {
				shortcodeStr = *shortcode
			}

			return db.UpdateLinkRow{},
				fmt.Errorf("%w: %s", apperrors.LinkShortcodeTaken, shortcodeStr)
		}

		return db.UpdateLinkRow{},
			fmt.Errorf("failed to update link: %w", err)
	}

	return updatedLink, nil
}

func (s *LinkService) DeleteLink(ctx context.Context, userID string, id uuid.UUID) (db.DeleteLinkRow, error) {
	deletedLink, err := s.queries.DeleteLink(ctx, db.DeleteLinkParams{
		ID:     id,
		UserID: userID,
	})

	if err != nil {
		if errors.Is(sql.ErrNoRows, err) {
			return db.DeleteLinkRow{}, fmt.Errorf("%w: %v", apperrors.LinkNotFound, err)
		}

		return db.DeleteLinkRow{}, fmt.Errorf("failed to delete link: %w", err)
	}

	return deletedLink, nil
}
