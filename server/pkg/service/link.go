package service

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/binary"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redis/go-redis/v9"
	"github.com/styltsou/url-shortener/server/pkg/analytics"
	"github.com/styltsou/url-shortener/server/pkg/db"
	apperrors "github.com/styltsou/url-shortener/server/pkg/errors"
	"github.com/styltsou/url-shortener/server/pkg/logger"
	"go.uber.org/zap"
)

const (
	// Cache key prefix for link lookups
	cacheKeyPrefix = "link:"
	// Cache TTL: 24 hours
	cacheTTL = 24 * time.Hour

	// Length of auto-generated shortcodes
	defaultCodeLen = 7
)

// base62Alphabet contains the characters used for shortcode generation.
// Using lowercase + uppercase + digits gives 62 possible characters per position.
const base62Alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// generateCode creates a deterministic shortcode from a URL and user ID.
// Uses SHA-256 as a content hash, then encodes the first 8 bytes as base62.
// Same inputs always produce the same shortcode (deterministic deduplication).
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
	GetLinkForRedirect(ctx context.Context, shortcode string) (db.GetLinkForRedirectRow, error)
	ListUserLinks(ctx context.Context, arg db.ListUserLinksParams) ([]db.ListUserLinksRow, error)
	CountUserLinks(ctx context.Context, arg db.CountUserLinksParams) (int64, error)
	GetLinkByIdAndUser(ctx context.Context, arg db.GetLinkByIdAndUserParams) (db.GetLinkByIdAndUserRow, error)
	GetLinkByShortcodeAndUser(ctx context.Context, arg db.GetLinkByShortcodeAndUserParams) (db.GetLinkByShortcodeAndUserRow, error)
	UpdateLink(ctx context.Context, arg db.UpdateLinkParams) (db.UpdateLinkRow, error)
	DeleteLink(ctx context.Context, arg db.DeleteLinkParams) (db.DeleteLinkRow, error)
	AddTagsToLink(ctx context.Context, arg db.AddTagsToLinkParams) error
	RemoveTagsFromLink(ctx context.Context, arg db.RemoveTagsFromLinkParams) error
	GetLinkByIdAndUserWithTags(ctx context.Context, arg db.GetLinkByIdAndUserWithTagsParams) (db.GetLinkByIdAndUserWithTagsRow, error)
	GetRecentLinks(ctx context.Context, arg db.GetRecentLinksParams) ([]db.GetRecentLinksRow, error)
	CountActiveLinks(ctx context.Context, userID string) (int64, error)
	ListUserLinkIDs(ctx context.Context, userID string) ([]uuid.UUID, error)
}

type LinkService struct {
	queries          LinkQueries
	cache            *redis.Client
	analyticsClient  *analytics.Client
	logger           logger.Logger
}

func NewLinkService(queries LinkQueries, cache *redis.Client, analyticsClient *analytics.Client, logger logger.Logger) *LinkService {
	return &LinkService{
		queries:         queries,
		cache:           cache,
		analyticsClient: analyticsClient,
		logger:          logger,
	}
}

func (s *LinkService) CreateShortLink(
	ctx context.Context,
	userID string,
	originalURL string,
	customShortcode *string,
	expiresAt *time.Time,
) (db.TryCreateLinkRow, error) {
	// Validate URL - return sentinel error that handlers will map
	if err := validateURL(originalURL); err != nil {
		return db.TryCreateLinkRow{}, err
	}

	// Validate expiration date if provided
	if expiresAt != nil && expiresAt.Before(time.Now()) {
		return db.TryCreateLinkRow{},
			fmt.Errorf("%w: expires_at must be set to a future time", apperrors.InvalidURL)
	}

	// Prepare expires_at for database
	// When expiresAt is nil, pgtype.Timestamp{Valid: false} will be converted to NULL in PostgreSQL
	var expiresAtTimestamp pgtype.Timestamp
	if expiresAt != nil {
		expiresAtTimestamp = pgtype.Timestamp{
			Time:  *expiresAt,
			Valid: true,
		}
	} else {
		expiresAtTimestamp = pgtype.Timestamp{Valid: false} // NULL expiration date
	}

	// If custom shortcode is provided, try once and return error on conflict
	if customShortcode != nil {
		link, err := s.queries.TryCreateLink(ctx, db.TryCreateLinkParams{
			Shortcode:   *customShortcode,
			OriginalUrl: originalURL,
			UserID:      userID,
			ExpiresAt:   expiresAtTimestamp,
		})

		if err == nil {
			return link, nil
		}

		// Collision: ON CONFLICT DO NOTHING returned no rows
		if errors.Is(err, sql.ErrNoRows) {
			return db.TryCreateLinkRow{},
				fmt.Errorf("%w: %s", apperrors.LinkShortcodeTaken, *customShortcode)
		}

		// Other database error - wrap with context
		return db.TryCreateLinkRow{},
			fmt.Errorf("failed to create link: %w", err)
	}

	// Auto-generate deterministic shortcode based on URL + userID
	code := generateCode(originalURL, userID)
	link, err := s.queries.TryCreateLink(ctx, db.TryCreateLinkParams{
		Shortcode:   code,
		OriginalUrl: originalURL,
		UserID:      userID,
		ExpiresAt:   expiresAtTimestamp,
	})

	if err != nil {
		// Collision: ON CONFLICT DO NOTHING returned no rows.
		// This is astronomically unlikely with SHA-256 + base62(7) → 3.5T combinations.
		if errors.Is(err, sql.ErrNoRows) {
			return db.TryCreateLinkRow{},
				fmt.Errorf("collision after inserting code %s: %w", code, err)
		}

		return db.TryCreateLinkRow{},
			fmt.Errorf("failed to create link: %w", err)
	}

	return link, nil
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

	totalPages := int((total + int64(limit) - 1) / int64(limit)) // Ceiling division

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
		cachedURL, err := s.cache.Get(ctx, cacheKey).Result()
		if err == nil {
			// Cache hit - return immediately
			s.logger.Debug("Cache hit for link redirect",
				zap.String("shortcode", code),
			)
			return db.GetLinkForRedirectRow{
				OriginalUrl: cachedURL,
			}, nil
		}
		// Cache miss or Redis error - continue to database lookup
		// (We don't log cache misses as errors, they're expected)
		if !errors.Is(err, redis.Nil) {
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

	// Populate cache for next time (non-blocking - don't fail if cache write fails)
	if s.cache != nil {
		if err := s.cache.Set(ctx, cacheKey, link.OriginalUrl, cacheTTL).Err(); err != nil {
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
	if s.analyticsClient == nil {
		return
	}

	s.analyticsClient.RecordClick(ctx, analytics.ClickEvent{
		LinkID:    linkID,
		Timestamp: time.Now(),
		IP:        ip,
		UserAgent: userAgent,
		Referrer:  referrer,
	})
}

// GetLinkAnalytics returns analytics data for a link
func (s *LinkService) GetLinkAnalytics(ctx context.Context, userID string, shortcode string) (*analytics.LinkAnalytics, error) {
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

	if s.analyticsClient == nil {
		return &analytics.LinkAnalytics{}, nil
	}

	until := time.Now()
	since := until.AddDate(0, 0, -30)

	result, err := s.analyticsClient.GetLinkAnalytics(ctx, link.ID, since, until)
	if err != nil {
		return nil, fmt.Errorf("failed to get analytics: %w", err)
	}

	return result, nil
}

func (s *LinkService) UpdateLink(
	ctx context.Context,
	userID string,
	id uuid.UUID,
	shortcode *string,
	isActive *bool,
	expiresAt *time.Time,
) (db.UpdateLinkRow, error) {

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
	s.invalidateCache(ctx, oldShortcode)
	s.invalidateCache(ctx, updatedLink.Shortcode)

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
	s.invalidateCache(ctx, deletedLink.Shortcode)

	return deletedLink, nil
}

// AddTagsToLink adds multiple tags to a link, ensuring both link and tags belong to the user
// Returns the updated link with all tags
func (s *LinkService) AddTagsToLink(ctx context.Context, userID string, linkID uuid.UUID, tagIDs []uuid.UUID) (db.GetLinkByIdAndUserWithTagsRow, error) {
	if len(tagIDs) == 0 {
		// No-op if empty, but still return the link
		link, err := s.queries.GetLinkByIdAndUserWithTags(ctx, db.GetLinkByIdAndUserWithTagsParams{
			ID:     linkID,
			UserID: userID,
		})
		if err != nil {
			return db.GetLinkByIdAndUserWithTagsRow{}, fmt.Errorf("failed to get link: %w", err)
		}
		return link, nil
	}

	err := s.queries.AddTagsToLink(ctx, db.AddTagsToLinkParams{
		LinkID: linkID,
		UserID: userID,
		TagIDs: tagIDs,
	})

	if err != nil {
		return db.GetLinkByIdAndUserWithTagsRow{}, fmt.Errorf("failed to add tags to link: %w", err)
	}

	// Fetch and return the updated link with tags
	link, err := s.queries.GetLinkByIdAndUserWithTags(ctx, db.GetLinkByIdAndUserWithTagsParams{
		ID:     linkID,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return db.GetLinkByIdAndUserWithTagsRow{}, fmt.Errorf("%w: %v", apperrors.LinkNotFound, err)
		}
		return db.GetLinkByIdAndUserWithTagsRow{}, fmt.Errorf("failed to get link after adding tags: %w", err)
	}

	return link, nil
}

// RemoveTagsFromLink removes multiple tags from a link, ensuring both link and tags belong to the user
// Returns the updated link with all tags
func (s *LinkService) RemoveTagsFromLink(ctx context.Context, userID string, linkID uuid.UUID, tagIDs []uuid.UUID) (db.GetLinkByIdAndUserWithTagsRow, error) {
	if len(tagIDs) == 0 {
		// No-op if empty, but still return the link
		link, err := s.queries.GetLinkByIdAndUserWithTags(ctx, db.GetLinkByIdAndUserWithTagsParams{
			ID:     linkID,
			UserID: userID,
		})
		if err != nil {
			return db.GetLinkByIdAndUserWithTagsRow{}, fmt.Errorf("failed to get link: %w", err)
		}
		return link, nil
	}

	err := s.queries.RemoveTagsFromLink(ctx, db.RemoveTagsFromLinkParams{
		LinkID: linkID,
		UserID: userID,
		TagIDs: tagIDs,
	})

	if err != nil {
		return db.GetLinkByIdAndUserWithTagsRow{}, fmt.Errorf("failed to remove tags from link: %w", err)
	}

	// Fetch and return the updated link with tags
	link, err := s.queries.GetLinkByIdAndUserWithTags(ctx, db.GetLinkByIdAndUserWithTagsParams{
		ID:     linkID,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return db.GetLinkByIdAndUserWithTagsRow{}, fmt.Errorf("%w: %v", apperrors.LinkNotFound, err)
		}
		return db.GetLinkByIdAndUserWithTagsRow{}, fmt.Errorf("failed to get link after removing tags: %w", err)
	}

	return link, nil
}

type DashboardStats struct {
	TotalLinks    int64                    `json:"total_links"`
	ActiveLinks   int64                    `json:"active_links"`
	TotalClicks   int                      `json:"total_clicks"`
	ClicksOverTime []analytics.ClicksOverTime `json:"clicks_over_time,omitempty"`
	RecentLinks   []db.GetRecentLinksRow   `json:"recent_links"`
}

func (s *LinkService) GetDashboardStats(ctx context.Context, userID string) (*DashboardStats, error) {
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
		RecentLinks: recentLinks,
	}

	if s.analyticsClient != nil {
		linkIDs, err := s.queries.ListUserLinkIDs(ctx, userID)
		if err != nil {
			s.logger.Warn("Failed to get user link IDs for analytics",
				zap.String("user_id", userID),
				zap.Error(err),
			)
		} else if len(linkIDs) > 0 {
			until := time.Now()
			since := until.AddDate(0, 0, -30)

			ua, err := s.analyticsClient.GetUserAnalytics(ctx, linkIDs, since, until)
			if err != nil {
				s.logger.Warn("Failed to get user analytics",
					zap.String("user_id", userID),
					zap.Error(err),
				)
			} else {
				dashboard.TotalClicks = ua.TotalClicks
				dashboard.ClicksOverTime = ua.ClicksOverTime
			}
		}
	}

	return dashboard, nil
}

// invalidateCache removes a link from the cache
// This is called after updates and deletes to ensure cache consistency
func (s *LinkService) invalidateCache(ctx context.Context, shortcode string) {
	if s.cache == nil {
		return
	}

	cacheKey := cacheKeyPrefix + shortcode
	if err := s.cache.Del(ctx, cacheKey).Err(); err != nil {
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
