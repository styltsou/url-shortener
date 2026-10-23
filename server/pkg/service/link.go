package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redis/go-redis/v9"
	"github.com/styltsou/url-shortener/server/pkg/analytics"
	"github.com/styltsou/url-shortener/server/pkg/db"
	"github.com/styltsou/url-shortener/server/pkg/dto"
	apperrors "github.com/styltsou/url-shortener/server/pkg/errors"
	"github.com/styltsou/url-shortener/server/pkg/logger"
	"go.uber.org/zap"
)

const (
	// Cache key prefix for link lookups
	cacheKeyPrefix = "link:redirect:v2:"
	// Bound stale redirects if Redis invalidation fails after a database change.
	cacheTTL = time.Minute

	// Length of auto-generated shortcodes
	defaultCodeLen  = 7
	maxCodeAttempts = 5
)

// base62Alphabet contains the characters used for shortcode generation.
// Using lowercase + uppercase + digits gives 62 possible characters per position.
const base62Alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// generateCode creates a deterministic shortcode from a URL and user ID.
// Uses SHA-256 as a content hash, then encodes the first 8 bytes as base62.
// Same inputs produce the same first candidate; occupied codes get random retries.
func generateCode(originalURL, userID string) string {
	h := sha256.Sum256([]byte(originalURL + "|" + userID))
	// Use first 8 bytes of the hash (64 bits) as a large number
	num := binary.BigEndian.Uint64(h[:8])
	// Base62 encode the number
	code := base62EncodeUint64(num)
	// Pad to ensure consistent length
	for len(code) < defaultCodeLen {
		code = string(base62Alphabet[0]) + code
	}
	return code[:defaultCodeLen]
}

// A retry uses a random code so unrelated links with the same hash prefix can coexist.
func randomCode() (string, error) {
	code := make([]byte, defaultCodeLen)
	for i := range code {
		for {
			var b [1]byte
			if _, err := rand.Read(b[:]); err != nil {
				return "", err
			}
			if b[0] < 248 { // Avoid modulo bias: 248 is divisible by 62.
				code[i] = base62Alphabet[int(b[0])%len(base62Alphabet)]
				break
			}
		}
	}
	return string(code), nil
}

func validShortcode(code string) bool {
	if len(code) == 0 || len(code) > 20 || strings.EqualFold(code, "api") {
		return false
	}
	for i := range len(code) {
		c := code[i]
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_') {
			return false
		}
	}
	return true
}

func shortcodeConflict(err error) bool {
	if errors.Is(err, sql.ErrNoRows) {
		return true
	}
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == apperrors.PgUniqueViolation && pgErr.ConstraintName == "idx_links_shortcode"
}

// base62EncodeUint64 converts a uint64 to a base62 string.
func base62EncodeUint64(num uint64) string {
	if num == 0 {
		return string(base62Alphabet[0])
	}
	var chars []byte
	for num > 0 {
		chars = append(chars, base62Alphabet[num%62])
		num /= 62
	}
	// Reverse the chars
	for i, j := 0, len(chars)-1; i < j; i, j = i+1, j-1 {
		chars[i], chars[j] = chars[j], chars[i]
	}
	return string(chars)
}

type LinkQueries interface {
	TryCreateLink(ctx context.Context, arg db.TryCreateLinkParams) (db.TryCreateLinkRow, error)
	ShortcodeExists(ctx context.Context, shortcode string) (bool, error)
	GetLinkForRedirect(ctx context.Context, shortcode string) (db.GetLinkForRedirectRow, error)
	ListUserLinks(ctx context.Context, arg db.ListUserLinksParams) ([]db.ListUserLinksRow, error)
	CountUserLinks(ctx context.Context, arg db.CountUserLinksParams) (int64, error)
	GetLinkByIdAndUser(ctx context.Context, arg db.GetLinkByIdAndUserParams) (db.GetLinkByIdAndUserRow, error)
	GetLinkByShortcodeAndUser(ctx context.Context, arg db.GetLinkByShortcodeAndUserParams) (db.GetLinkByShortcodeAndUserRow, error)
	UpdateLink(ctx context.Context, arg db.UpdateLinkParams) (db.UpdateLinkRow, error)
	DeleteLink(ctx context.Context, arg db.DeleteLinkParams) (db.DeleteLinkRow, error)
	AddTagsToLink(ctx context.Context, arg db.AddTagsToLinkParams) error
	RemoveTagsFromLink(ctx context.Context, arg db.RemoveTagsFromLinkParams) error
	CountOwnedTags(ctx context.Context, arg db.CountOwnedTagsParams) (int64, error)
	GetLinkByIdAndUserWithTags(ctx context.Context, arg db.GetLinkByIdAndUserWithTagsParams) (db.GetLinkByIdAndUserWithTagsRow, error)
	LockLinkURL(ctx context.Context, arg db.LockLinkURLParams) error
	GetExistingLinkIDByURLAndUser(ctx context.Context, arg db.GetExistingLinkIDByURLAndUserParams) (uuid.UUID, error)
	GetRecentLinks(ctx context.Context, arg db.GetRecentLinksParams) ([]db.GetRecentLinksRow, error)
	CountActiveLinks(ctx context.Context, userID string) (int64, error)
	ListUserLinkIDs(ctx context.Context, userID string) ([]uuid.UUID, error)
}

// ShortcodeAvailable checks the same global namespace used by link creation.
// The result is advisory; the unique index decides the outcome of an insert.
func (s *LinkService) ShortcodeAvailable(ctx context.Context, shortcode string) (bool, error) {
	if !validShortcode(shortcode) {
		return false, apperrors.InvalidShortcode
	}
	exists, err := s.queries.ShortcodeExists(ctx, shortcode)
	if err != nil {
		return false, fmt.Errorf("failed to check shortcode availability: %w", err)
	}
	return !exists, nil
}

type TxBeginner interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}

type redirectCache interface {
	Get(ctx context.Context, key string) *redis.StringCmd
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
}

type cachedRedirect struct {
	ID        uuid.UUID  `json:"id"`
	URL       string     `json:"url"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

type LinkService struct {
	txBeginner        TxBeginner
	queries           LinkQueries
	cache             redirectCache
	analyticsClient   analytics.ClientInterface
	logger            logger.Logger
	runInTx           func(ctx context.Context, fn func(LinkQueries) error) error
	generateShortcode func(originalURL, userID string, attempt int) (string, error)
}

func NewLinkService(txBeginner TxBeginner, queries LinkQueries, cache redirectCache, analyticsClient analytics.ClientInterface, logger logger.Logger) *LinkService {
	s := &LinkService{
		txBeginner:      txBeginner,
		queries:         queries,
		cache:           cache,
		analyticsClient: analyticsClient,
		logger:          logger,
	}
	s.runInTx = func(ctx context.Context, fn func(LinkQueries) error) error {
		tx, err := s.txBeginner.Begin(ctx)
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}
		defer func() {
			_ = tx.Rollback(ctx)
		}()

		qtx := db.New(tx)
		if err := fn(qtx); err != nil {
			return err
		}
		return tx.Commit(ctx)
	}
	return s
}

func (s *LinkService) CreateShortLink(
	ctx context.Context,
	userID string,
	originalURL string,
	customShortcode *string,
	expiresAt *time.Time,
) (db.TryCreateLinkRow, error) {
	return s.createShortLink(ctx, s.queries, userID, originalURL, customShortcode, expiresAt)
}

func (s *LinkService) CreateShortLinkWithTags(
	ctx context.Context,
	userID string,
	originalURL string,
	customShortcode *string,
	expiresAt *time.Time,
	tagIDs []uuid.UUID,
) (db.GetLinkByIdAndUserWithTagsRow, error) {
	link, _, err := s.CreateOrFindShortLinkWithTags(ctx, userID, originalURL, customShortcode, expiresAt, tagIDs, true)
	return link, err
}

// CreateOrFindShortLinkWithTags offers an active link for the same exact URL to its owner.
// Explicit custom codes and forceNew requests always create a separate link.
func (s *LinkService) CreateOrFindShortLinkWithTags(
	ctx context.Context,
	userID string,
	originalURL string,
	customShortcode *string,
	expiresAt *time.Time,
	tagIDs []uuid.UUID,
	forceNew bool,
) (db.GetLinkByIdAndUserWithTagsRow, bool, error) {
	var created db.GetLinkByIdAndUserWithTagsRow
	var existing bool
	if err := s.runInTx(ctx, func(qtx LinkQueries) error {
		if !forceNew && customShortcode == nil {
			if err := validateURL(originalURL); err != nil {
				return err
			}
			key := db.LockLinkURLParams{UserID: userID, OriginalUrl: originalURL}
			if err := qtx.LockLinkURL(ctx, key); err != nil {
				return fmt.Errorf("failed to lock link URL: %w", err)
			}
			id, err := qtx.GetExistingLinkIDByURLAndUser(ctx, db.GetExistingLinkIDByURLAndUserParams(key))
			if err != nil && !errors.Is(err, pgx.ErrNoRows) {
				return fmt.Errorf("failed to find existing link: %w", err)
			}
			if err == nil {
				created, err = qtx.GetLinkByIdAndUserWithTags(ctx, db.GetLinkByIdAndUserWithTagsParams{ID: id, UserID: userID})
				if err != nil {
					return fmt.Errorf("failed to get existing link: %w", err)
				}
				existing = true
				return nil
			}
		}
		link, err := s.createShortLink(ctx, qtx, userID, originalURL, customShortcode, expiresAt)
		if err != nil {
			return err
		}

		uniqueTagIDs := uniqueUUIDs(tagIDs)
		if len(uniqueTagIDs) > 0 {
			if err := s.validateOwnedTags(ctx, qtx, userID, uniqueTagIDs); err != nil {
				return err
			}
			if err := qtx.AddTagsToLink(ctx, db.AddTagsToLinkParams{
				LinkID: link.ID,
				UserID: userID,
				TagIDs: uniqueTagIDs,
			}); err != nil {
				return fmt.Errorf("failed to add tags to link: %w", err)
			}
		}

		created, err = qtx.GetLinkByIdAndUserWithTags(ctx, db.GetLinkByIdAndUserWithTagsParams{
			ID:     link.ID,
			UserID: userID,
		})
		if err != nil {
			return fmt.Errorf("failed to get created link: %w", err)
		}
		return nil
	}); err != nil {
		return db.GetLinkByIdAndUserWithTagsRow{}, false, err
	}

	return created, existing, nil
}

func (s *LinkService) createShortLink(
	ctx context.Context,
	queries LinkQueries,
	userID string,
	originalURL string,
	customShortcode *string,
	expiresAt *time.Time,
) (db.TryCreateLinkRow, error) {
	if err := validateURL(originalURL); err != nil {
		return db.TryCreateLinkRow{}, err
	}

	var expiresAtTimestamp pgtype.Timestamp
	if expiresAt != nil {
		expiresAtTimestamp = pgtype.Timestamp{
			Time:  *expiresAt,
			Valid: true,
		}
	}

	if customShortcode != nil {
		if !validShortcode(*customShortcode) {
			return db.TryCreateLinkRow{}, apperrors.InvalidShortcode
		}
		link, err := queries.TryCreateLink(ctx, db.TryCreateLinkParams{
			Shortcode:   *customShortcode,
			OriginalUrl: originalURL,
			UserID:      userID,
			ExpiresAt:   expiresAtTimestamp,
		})
		if err == nil {
			return link, nil
		}
		if shortcodeConflict(err) {
			return db.TryCreateLinkRow{},
				fmt.Errorf("%w: %s", apperrors.LinkShortcodeTaken, *customShortcode)
		}
		return db.TryCreateLinkRow{}, fmt.Errorf("failed to create link: %w", err)
	}

	for attempt := range maxCodeAttempts {
		var code string
		var err error
		if s.generateShortcode != nil {
			code, err = s.generateShortcode(originalURL, userID, attempt)
		} else if attempt == 0 {
			code = generateCode(originalURL, userID)
		} else {
			code, err = randomCode()
		}
		if err != nil {
			return db.TryCreateLinkRow{}, fmt.Errorf("failed to generate shortcode: %w", err)
		}
		link, err := queries.TryCreateLink(ctx, db.TryCreateLinkParams{
			Shortcode: code, OriginalUrl: originalURL, UserID: userID, ExpiresAt: expiresAtTimestamp,
		})
		if err == nil {
			return link, nil
		}
		if !shortcodeConflict(err) {
			return db.TryCreateLinkRow{}, fmt.Errorf("failed to create link: %w", err)
		}
	}
	return db.TryCreateLinkRow{}, fmt.Errorf("failed to create link: shortcode collisions after %d attempts", maxCodeAttempts)
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

	if parsedURL.Hostname() == "" || parsedURL.User != nil {
		return fmt.Errorf("%w: URL must have a valid host", apperrors.InvalidURL)
	}
	if port := parsedURL.Port(); port != "" {
		number, err := strconv.Atoi(port)
		if err != nil || number < 1 || number > 65535 {
			return fmt.Errorf("%w: URL port must be between 1 and 65535", apperrors.InvalidURL)
		}
	}

	if len(rawURL) > 2048 {
		return fmt.Errorf("%w: URL is too long (max 2048 characters)", apperrors.InvalidURL)
	}

	return nil
}

type ListLinksResult struct {
	Links      []db.ListUserLinksRow
	Total      int64
	Page       int
	Limit      int
	TotalPages int
}

func (s *LinkService) ListAllLinks(ctx context.Context, userID string, isActive *bool, tagIDs []uuid.UUID, page, limit int) (*ListLinksResult, error) {
	s.logger.Debug("Querying database for user links",
		zap.String("user_id", userID),
		zap.Any("is_active", isActive),
		zap.Any("tag_ids", tagIDs),
		zap.Int("page", page),
		zap.Int("limit", limit),
	)

	// Validate and set defaults
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 5
	}
	if limit > 100 {
		limit = 100 // Max limit
	}
	if page > math.MaxInt32/limit {
		return nil, fmt.Errorf("page is too large for the database offset")
	}

	offset := (page - 1) * limit

	// Get total count
	countParams := db.CountUserLinksParams{
		UserID:   userID,
		IsActive: isActive,
		TagIds:   tagIDs,
	}
	total, err := s.queries.CountUserLinks(ctx, countParams)
	if err != nil {
		s.logger.Error("Database query failed for CountUserLinks",
			zap.Error(err),
			zap.String("user_id", userID),
		)
		return nil, fmt.Errorf("failed to count links: %w", err)
	}

	// Get paginated links
	params := db.ListUserLinksParams{
		UserID:   userID,
		IsActive: isActive,
		TagIds:   tagIDs,
		Offset:   int32(offset),
		Limit:    int32(limit),
	}

	links, err := s.queries.ListUserLinks(ctx, params)
	if err != nil {
		s.logger.Error("Database query failed for ListUserLinks",
			zap.Error(err),
			zap.String("user_id", userID),
		)
		return nil, fmt.Errorf("failed to get links: %w", err)
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit)) //nolint:gosec // small int, max 100

	s.logger.Debug("Database query completed for ListUserLinks",
		zap.String("user_id", userID),
		zap.Int("links_found", len(links)),
		zap.Int64("total", total),
		zap.Int("total_pages", totalPages),
	)

	return &ListLinksResult{
		Links:      links,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
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
	// Cache-aside pattern: Check cache first
	cacheKey := cacheKeyPrefix + code

	// Try to get from cache if Redis is available
	if s.cache != nil {
		payload, err := s.cache.Get(ctx, cacheKey).Result()
		if err == nil {
			var cached cachedRedirect
			if json.Unmarshal([]byte(payload), &cached) == nil && cached.ID != uuid.Nil && validateURL(cached.URL) == nil {
				if cached.ExpiresAt == nil || cached.ExpiresAt.After(time.Now()) {
					row := db.GetLinkForRedirectRow{ID: cached.ID, OriginalUrl: cached.URL}
					if cached.ExpiresAt != nil {
						row.ExpiresAt = pgtype.Timestamp{Time: *cached.ExpiresAt, Valid: true}
					}
					s.logger.Debug("Cache hit for link redirect", zap.String("shortcode", code))
					return row, nil
				}
			}
			// Invalid or expired entries are never used as redirects. In particular,
			// URL-only values from the old cache format cannot be trusted.
			s.logger.Warn("Ignoring invalid redirect cache entry", zap.String("shortcode", code))
		}
		// Cache miss or Redis error - continue to database lookup
		// (We don't log cache misses as errors, they're expected)
		if err != nil && !errors.Is(err, redis.Nil) {
			// Redis error (not a cache miss) - log but continue
			s.logger.Warn("Redis cache error, falling back to database",
				zap.String("shortcode", code),
				zap.Error(err),
			)
		}
	}

	// Cache miss or Redis unavailable - query database
	link, err := s.queries.GetLinkForRedirect(ctx, code)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return db.GetLinkForRedirectRow{}, fmt.Errorf("%w: code %s", apperrors.LinkNotFound, code)
		}
		return db.GetLinkForRedirectRow{}, fmt.Errorf("failed to get link: %w", err)
	}

	// Check if the link has expired (in Go, so we can return a specific error)
	now := time.Now()
	if link.ExpiresAt.Valid && !link.ExpiresAt.Time.After(now) {
		return db.GetLinkForRedirectRow{}, fmt.Errorf("%w: code %s", apperrors.LinkExpired, code)
	}

	// Populate cache for next time (non-blocking - don't fail if cache write fails)
	if s.cache != nil {
		ttl := cacheTTL
		cached := cachedRedirect{ID: link.ID, URL: link.OriginalUrl}
		if link.ExpiresAt.Valid {
			expiresAt := link.ExpiresAt.Time
			cached.ExpiresAt = &expiresAt
			if remaining := expiresAt.Sub(now); remaining < ttl {
				ttl = remaining
			}
		}
		payload, marshalErr := json.Marshal(cached)
		if marshalErr != nil {
			return link, nil
		}
		if err := s.cache.Set(ctx, cacheKey, payload, ttl).Err(); err != nil {
			// Log but don't fail - cache write errors shouldn't break the request
			s.logger.Warn("Failed to populate cache",
				zap.String("shortcode", code),
				zap.Error(err),
			)
		} else {
			s.logger.Debug("Cache populated for link redirect",
				zap.String("shortcode", code),
			)
		}
	}

	return link, nil
}

// RecordClick records a click event for analytics
func (s *LinkService) RecordClick(ctx context.Context, linkID uuid.UUID, ip, userAgent, referrer string) {
	s.analyticsClient.RecordClick(ctx, analytics.ClickEvent{
		LinkID:    linkID,
		Timestamp: time.Now(),
		IP:        ip,
		UserAgent: userAgent,
		Referrer:  referrer,
	})
}

// GetLinkAnalytics returns analytics data for a link
func (s *LinkService) GetLinkAnalytics(ctx context.Context, userID string, shortcode string, filter analytics.Filter) (*analytics.LinkAnalytics, error) {
	link, err := s.queries.GetLinkByShortcodeAndUser(ctx, db.GetLinkByShortcodeAndUserParams{
		Shortcode: shortcode,
		UserID:    userID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: %v", apperrors.LinkNotFound, err)
		}
		return nil, fmt.Errorf("failed to get link: %w", err)
	}

	since, until := filter.Bounds(time.Now())

	if s.analyticsClient == nil {
		return nil, apperrors.AnalyticsUnavailable
	}
	result, err := s.analyticsClient.GetLinkAnalytics(ctx, link.ID, since, until)
	if err != nil {
		s.logger.Warn("Failed to get link analytics", zap.String("shortcode", shortcode), zap.Error(err))
		return nil, apperrors.AnalyticsUnavailable
	}

	result.ClicksOverTime = filter.BucketClicks(result.ClicksOverTime, since, until)
	return result, nil
}

func (s *LinkService) UpdateLink(
	ctx context.Context,
	userID string,
	id uuid.UUID,
	shortcode *string,
	isActive *bool,
	expiresAtSet bool,
	expiresAt *time.Time,
) (db.UpdateLinkRow, error) {
	if shortcode != nil && !validShortcode(*shortcode) {
		return db.UpdateLinkRow{}, apperrors.InvalidShortcode
	}

	// Fetch the existing link to get the old shortcode for cache invalidation
	existingLink, err := s.queries.GetLinkByIdAndUser(ctx, db.GetLinkByIdAndUserParams{
		ID:     id,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return db.UpdateLinkRow{},
				fmt.Errorf("%w: %v", apperrors.LinkNotFound, err)
		}
		return db.UpdateLinkRow{},
			fmt.Errorf("failed to get link for update: %w", err)
	}
	oldShortcode := existingLink.Shortcode

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
		UserID:       userID,
		ID:           id,
		Shortcode:    shortcode,
		IsActive:     isActive,
		ExpiresAtSet: expiresAtSet,
		ExpiresAt:    expiresAtTimestamp,
	})

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return db.UpdateLinkRow{},
				fmt.Errorf("%w: %v", apperrors.LinkNotFound, err)
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == apperrors.PgUniqueViolation {
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

	// Invalidate cache for both old and new shortcodes
	s.invalidateCache(oldShortcode)
	s.invalidateCache(updatedLink.Shortcode)

	return updatedLink, nil
}

func (s *LinkService) DeleteLink(ctx context.Context, userID string, id uuid.UUID) (db.DeleteLinkRow, error) {
	deletedLink, err := s.queries.DeleteLink(ctx, db.DeleteLinkParams{
		ID:     id,
		UserID: userID,
	})

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return db.DeleteLinkRow{}, fmt.Errorf("%w: %v", apperrors.LinkNotFound, err)
		}

		return db.DeleteLinkRow{}, fmt.Errorf("failed to delete link: %w", err)
	}

	// Invalidate cache after successful deletion
	s.invalidateCache(deletedLink.Shortcode)

	return deletedLink, nil
}

// AddTagsToLink adds multiple tags to a link, ensuring both link and tags belong to the user
// Returns the updated link with all tags
func (s *LinkService) AddTagsToLink(ctx context.Context, userID string, linkID uuid.UUID, tagIDs []uuid.UUID) (db.GetLinkByIdAndUserWithTagsRow, error) {
	uniqueTagIDs := uniqueUUIDs(tagIDs)
	if len(uniqueTagIDs) == 0 {
		link, err := s.queries.GetLinkByIdAndUserWithTags(ctx, db.GetLinkByIdAndUserWithTagsParams{
			ID:     linkID,
			UserID: userID,
		})
		if err != nil {
			return db.GetLinkByIdAndUserWithTagsRow{}, fmt.Errorf("failed to get link: %w", err)
		}
		return link, nil
	}

	var link db.GetLinkByIdAndUserWithTagsRow
	if err := s.runInTx(ctx, func(qtx LinkQueries) error {
		if err := s.validateOwnedLink(ctx, qtx, userID, linkID); err != nil {
			return err
		}
		if err := s.validateOwnedTags(ctx, qtx, userID, uniqueTagIDs); err != nil {
			return err
		}
		if err := qtx.AddTagsToLink(ctx, db.AddTagsToLinkParams{
			LinkID: linkID,
			UserID: userID,
			TagIDs: uniqueTagIDs,
		}); err != nil {
			return fmt.Errorf("failed to add tags to link: %w", err)
		}

		var err error
		link, err = qtx.GetLinkByIdAndUserWithTags(ctx, db.GetLinkByIdAndUserWithTagsParams{
			ID:     linkID,
			UserID: userID,
		})
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("%w: %v", apperrors.LinkNotFound, err)
			}
			return fmt.Errorf("failed to get link after adding tags: %w", err)
		}
		return nil
	}); err != nil {
		return db.GetLinkByIdAndUserWithTagsRow{}, err
	}

	return link, nil
}

// RemoveTagsFromLink removes multiple tags from a link, ensuring both link and tags belong to the user
// Returns the updated link with all tags
func (s *LinkService) RemoveTagsFromLink(ctx context.Context, userID string, linkID uuid.UUID, tagIDs []uuid.UUID) (db.GetLinkByIdAndUserWithTagsRow, error) {
	uniqueTagIDs := uniqueUUIDs(tagIDs)
	if len(uniqueTagIDs) == 0 {
		link, err := s.queries.GetLinkByIdAndUserWithTags(ctx, db.GetLinkByIdAndUserWithTagsParams{
			ID:     linkID,
			UserID: userID,
		})
		if err != nil {
			return db.GetLinkByIdAndUserWithTagsRow{}, fmt.Errorf("failed to get link: %w", err)
		}
		return link, nil
	}

	var link db.GetLinkByIdAndUserWithTagsRow
	if err := s.runInTx(ctx, func(qtx LinkQueries) error {
		if err := s.validateOwnedLink(ctx, qtx, userID, linkID); err != nil {
			return err
		}
		if err := s.validateOwnedTags(ctx, qtx, userID, uniqueTagIDs); err != nil {
			return err
		}
		if err := qtx.RemoveTagsFromLink(ctx, db.RemoveTagsFromLinkParams{
			LinkID: linkID,
			UserID: userID,
			TagIDs: uniqueTagIDs,
		}); err != nil {
			return fmt.Errorf("failed to remove tags from link: %w", err)
		}

		var err error
		link, err = qtx.GetLinkByIdAndUserWithTags(ctx, db.GetLinkByIdAndUserWithTagsParams{
			ID:     linkID,
			UserID: userID,
		})
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("%w: %v", apperrors.LinkNotFound, err)
			}
			return fmt.Errorf("failed to get link after removing tags: %w", err)
		}
		return nil
	}); err != nil {
		return db.GetLinkByIdAndUserWithTagsRow{}, err
	}

	return link, nil
}

func (s *LinkService) validateOwnedLink(ctx context.Context, queries LinkQueries, userID string, linkID uuid.UUID) error {
	_, err := queries.GetLinkByIdAndUser(ctx, db.GetLinkByIdAndUserParams{
		ID:     linkID,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("%w: %v", apperrors.LinkNotFound, err)
		}
		return fmt.Errorf("failed to get link: %w", err)
	}
	return nil
}

func (s *LinkService) validateOwnedTags(ctx context.Context, queries LinkQueries, userID string, tagIDs []uuid.UUID) error {
	ownedCount, err := queries.CountOwnedTags(ctx, db.CountOwnedTagsParams{
		UserID: userID,
		TagIDs: tagIDs,
	})
	if err != nil {
		return fmt.Errorf("failed to validate tags: %w", err)
	}
	if ownedCount != int64(len(tagIDs)) {
		return fmt.Errorf("%w: one or more tags were not found", apperrors.TagNotFound)
	}
	return nil
}

func uniqueUUIDs(ids []uuid.UUID) []uuid.UUID {
	if len(ids) == 0 {
		return nil
	}
	seen := make(map[uuid.UUID]struct{}, len(ids))
	unique := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		unique = append(unique, id)
	}
	return unique
}

type DashboardStats struct {
	TotalLinks         int64                      `json:"total_links"`
	ActiveLinks        int64                      `json:"active_links"`
	TotalClicks        int                        `json:"total_clicks"`
	AnalyticsAvailable bool                       `json:"analytics_available"`
	ClicksOverTime     []analytics.ClicksOverTime `json:"clicks_over_time,omitempty"`
	RecentLinks        []dto.LinkResponse         `json:"recent_links"`
}

func (s *LinkService) GetDashboardStats(ctx context.Context, userID string, filter analytics.Filter) (*DashboardStats, error) {
	totalLinks, err := s.queries.CountUserLinks(ctx, db.CountUserLinksParams{
		UserID:   userID,
		IsActive: nil,
		TagIds:   nil,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to count total links: %w", err)
	}

	activeLinks, err := s.queries.CountActiveLinks(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to count active links: %w", err)
	}

	recentLinks, err := s.queries.GetRecentLinks(ctx, db.GetRecentLinksParams{
		UserID: userID,
		Limit:  5,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get recent links: %w", err)
	}

	if recentLinks == nil {
		recentLinks = []db.GetRecentLinksRow{}
	}

	dashboard := &DashboardStats{
		TotalLinks:  totalLinks,
		ActiveLinks: activeLinks,
		RecentLinks: dto.LinksFromRecentRows(recentLinks),
	}

	if s.analyticsClient != nil {
		linkIDs, err := s.queries.ListUserLinkIDs(ctx, userID)
		if err != nil {
			s.logger.Warn("Failed to get user link IDs for analytics",
				zap.String("user_id", userID),
				zap.Error(err),
			)
		} else if len(linkIDs) > 0 {
			since, until := filter.Bounds(time.Now())

			ua, err := s.analyticsClient.GetUserAnalytics(ctx, linkIDs, since, until)
			if err != nil {
				s.logger.Warn("Failed to get user analytics",
					zap.String("user_id", userID),
					zap.Error(err),
				)
			} else {
				dashboard.TotalClicks = ua.TotalClicks
				dashboard.ClicksOverTime = filter.BucketClicks(ua.ClicksOverTime, since, until)
				dashboard.AnalyticsAvailable = true
			}
		}
	}

	return dashboard, nil
}

// invalidateCache removes a link from the cache
// This is called after updates and deletes to ensure cache consistency
func (s *LinkService) invalidateCache(shortcode string) {
	if s.cache == nil {
		return
	}

	// A completed database mutation must not skip invalidation just because its
	// incoming HTTP request was canceled while the response was being written.
	invalidateCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	cacheKey := cacheKeyPrefix + shortcode
	if err := s.cache.Del(invalidateCtx, cacheKey).Err(); err != nil {
		// Log but don't fail - cache invalidation errors shouldn't break the request
		s.logger.Warn("Failed to invalidate cache",
			zap.String("shortcode", shortcode),
			zap.Error(err),
		)
	} else {
		s.logger.Debug("Cache invalidated",
			zap.String("shortcode", shortcode),
		)
	}
}
