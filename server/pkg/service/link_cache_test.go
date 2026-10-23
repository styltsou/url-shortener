package service

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redis/go-redis/v9"
	"github.com/styltsou/url-shortener/server/pkg/analytics"
	"github.com/styltsou/url-shortener/server/pkg/db"
	apperrors "github.com/styltsou/url-shortener/server/pkg/errors"
)

type fakeRedirectCache struct {
	values map[string]string
	ttl    time.Duration
	getErr error
	setErr error
	delErr error
}

func (c *fakeRedirectCache) Get(_ context.Context, key string) *redis.StringCmd {
	if c.getErr != nil {
		return redis.NewStringResult("", c.getErr)
	}
	value, ok := c.values[key]
	if !ok {
		return redis.NewStringResult("", redis.Nil)
	}
	return redis.NewStringResult(value, nil)
}

func (c *fakeRedirectCache) Set(_ context.Context, key string, value interface{}, ttl time.Duration) *redis.StatusCmd {
	if c.setErr != nil {
		return redis.NewStatusResult("", c.setErr)
	}
	if c.values == nil {
		c.values = make(map[string]string)
	}
	data, ok := value.([]byte)
	if !ok {
		return redis.NewStatusResult("", errors.New("expected serialized bytes"))
	}
	c.values[key] = string(data)
	c.ttl = ttl
	return redis.NewStatusResult("OK", nil)
}

func (c *fakeRedirectCache) Del(_ context.Context, keys ...string) *redis.IntCmd {
	if c.delErr != nil {
		return redis.NewIntResult(0, c.delErr)
	}
	for _, key := range keys {
		delete(c.values, key)
	}
	return redis.NewIntResult(int64(len(keys)), nil)
}

type recordingAnalytics struct{ linkID uuid.UUID }

func (a *recordingAnalytics) RecordClick(_ context.Context, event analytics.ClickEvent) {
	a.linkID = event.LinkID
}
func (a *recordingAnalytics) GetLinkAnalytics(context.Context, uuid.UUID, time.Time, time.Time) (*analytics.LinkAnalytics, error) {
	return nil, nil
}
func (a *recordingAnalytics) GetUserAnalytics(context.Context, []uuid.UUID, time.Time, time.Time) (*analytics.UserAnalytics, error) {
	return nil, nil
}
func (a *recordingAnalytics) Close() error { return nil }
func (a *recordingAnalytics) DeliveryStats() analytics.DeliveryStats {
	return analytics.DeliveryStats{Status: analytics.DeliveryOK}
}

func TestRedirectCachePreservesLinkIdentityAndClickAttribution(t *testing.T) {
	ctx := context.Background()
	linkID := uuid.New()
	expires := time.Now().Add(10 * time.Minute)
	cache := &fakeRedirectCache{}
	lookups := 0
	queries := &mockQueries{GetLinkForRedirectFunc: func(context.Context, string) (db.GetLinkForRedirectRow, error) {
		lookups++
		return db.GetLinkForRedirectRow{ID: linkID, OriginalUrl: "https://example.com/path", ExpiresAt: pgtype.Timestamp{Time: expires, Valid: true}}, nil
	}}
	clicks := &recordingAnalytics{}
	svc := &LinkService{queries: queries, cache: cache, analyticsClient: clicks, logger: createTestLogger()}
	first, err := svc.GetOriginalURL(ctx, "code123")
	if err != nil {
		t.Fatal(err)
	}
	second, err := svc.GetOriginalURL(ctx, "code123")
	if err != nil {
		t.Fatal(err)
	}
	if lookups != 1 || first.ID != second.ID || first.OriginalUrl != second.OriginalUrl || !second.ExpiresAt.Valid {
		t.Fatalf("cache hit differs from miss: lookups=%d first=%+v second=%+v", lookups, first, second)
	}
	if cache.ttl != cacheTTL || cache.ttl > 10*time.Minute {
		t.Fatalf("cache TTL %v does not respect expiration", cache.ttl)
	}
	svc.RecordClick(ctx, second.ID, "", "", "")
	if clicks.linkID != linkID {
		t.Fatalf("cached click attributed to %s, want %s", clicks.linkID, linkID)
	}
}

func TestRedirectCacheRejectsExpiredAndLegacyEntries(t *testing.T) {
	ctx := context.Background()
	linkID := uuid.New()
	cache := &fakeRedirectCache{values: map[string]string{cacheKeyPrefix + "old": "https://old.example"}}
	lookups := 0
	queries := &mockQueries{GetLinkForRedirectFunc: func(_ context.Context, code string) (db.GetLinkForRedirectRow, error) {
		lookups++
		if code == "expired" {
			return db.GetLinkForRedirectRow{ID: linkID, OriginalUrl: "https://example.com", ExpiresAt: pgtype.Timestamp{Time: time.Now().Add(-time.Second), Valid: true}}, nil
		}
		return db.GetLinkForRedirectRow{ID: linkID, OriginalUrl: "https://example.com"}, nil
	}}
	svc := &LinkService{queries: queries, cache: cache, logger: createTestLogger()}
	legacy, err := svc.GetOriginalURL(ctx, "old")
	if err != nil || legacy.ID != linkID || legacy.OriginalUrl != "https://example.com" {
		t.Fatalf("legacy cache value was trusted: link=%+v err=%v", legacy, err)
	}
	if lookups != 1 {
		t.Fatalf("legacy cache should cause a DB lookup, got %d", lookups)
	}
	stale, err := json.Marshal(cachedRedirect{ID: linkID, URL: "https://stale.example", ExpiresAt: &time.Time{}})
	if err != nil {
		t.Fatal(err)
	}
	cache.values[cacheKeyPrefix+"expired"] = string(stale)
	_, err = svc.GetOriginalURL(ctx, "expired")
	if !errors.Is(err, apperrors.LinkExpired) || lookups != 2 {
		t.Fatalf("expired cached value bypassed database expiration: lookups=%d err=%v", lookups, err)
	}
}

func TestRedirectCacheTTLStopsAtLinkExpiration(t *testing.T) {
	expires := time.Now().Add(10 * time.Second)
	cache := &fakeRedirectCache{}
	svc := &LinkService{
		queries: &mockQueries{GetLinkForRedirectFunc: func(context.Context, string) (db.GetLinkForRedirectRow, error) {
			return db.GetLinkForRedirectRow{ID: uuid.New(), OriginalUrl: "https://example.com", ExpiresAt: pgtype.Timestamp{Time: expires, Valid: true}}, nil
		}},
		cache: cache, logger: createTestLogger(),
	}
	if _, err := svc.GetOriginalURL(context.Background(), "soon"); err != nil {
		t.Fatal(err)
	}
	if cache.ttl <= 0 || cache.ttl > 10*time.Second || cache.ttl >= cacheTTL {
		t.Fatalf("cache TTL %v exceeds link expiration", cache.ttl)
	}
}

func TestRedirectFallsBackWhenRedisFails(t *testing.T) {
	cache := &fakeRedirectCache{getErr: errors.New("redis unavailable"), setErr: errors.New("redis unavailable")}
	linkID := uuid.New()
	svc := &LinkService{
		queries: &mockQueries{GetLinkForRedirectFunc: func(context.Context, string) (db.GetLinkForRedirectRow, error) {
			return db.GetLinkForRedirectRow{ID: linkID, OriginalUrl: "https://example.com"}, nil
		}},
		cache:  cache,
		logger: createTestLogger(),
	}
	link, err := svc.GetOriginalURL(context.Background(), "code123")
	if err != nil || link.ID != linkID {
		t.Fatalf("Redis failure broke redirect: link=%+v err=%v", link, err)
	}
}

func TestCacheInvalidationFailureDoesNotUndoDatabaseMutation(t *testing.T) {
	ctx := context.Background()
	id := uuid.New()
	cache := &fakeRedirectCache{values: map[string]string{cacheKeyPrefix + "old": "cached"}, delErr: errors.New("redis unavailable")}
	queries := &mockQueries{
		GetLinkByIdAndUserFunc: func(context.Context, db.GetLinkByIdAndUserParams) (db.GetLinkByIdAndUserRow, error) {
			return db.GetLinkByIdAndUserRow{ID: id, Shortcode: "old"}, nil
		},
		UpdateLinkFunc: func(context.Context, db.UpdateLinkParams) (db.UpdateLinkRow, error) {
			return db.UpdateLinkRow{ID: id, Shortcode: "new"}, nil
		},
		DeleteLinkFunc: func(context.Context, db.DeleteLinkParams) (db.DeleteLinkRow, error) {
			return db.DeleteLinkRow{ID: id, Shortcode: "old"}, nil
		},
	}
	svc := &LinkService{queries: queries, cache: cache, logger: createTestLogger()}
	if _, err := svc.UpdateLink(ctx, "user", id, nil, nil, false, nil); err != nil {
		t.Fatalf("database update failed because Redis was unavailable: %v", err)
	}
	if _, err := svc.DeleteLink(ctx, "user", id); err != nil {
		t.Fatalf("database delete failed because Redis was unavailable: %v", err)
	}
	if _, ok := cache.values[cacheKeyPrefix+"old"]; !ok {
		t.Fatal("fake Redis unexpectedly deleted the entry despite returning an error")
	}
}
