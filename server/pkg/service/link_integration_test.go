package service

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/styltsou/url-shortener/server/pkg/db"
	apperrors "github.com/styltsou/url-shortener/server/pkg/errors"
)

// The opt-in test uses its own schema and drops it after the assertions.
func integrationPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("set TEST_DATABASE_URL for PostgreSQL integration test")
	}
	ctx := context.Background()
	admin, err := pgx.Connect(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	schema := "test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if _, err := admin.Exec(ctx, "CREATE SCHEMA "+schema); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = admin.Exec(context.Background(), "DROP SCHEMA "+schema+" CASCADE")
		_ = admin.Close(context.Background())
	})
	config, err := pgxpool.ParseConfig(url)
	if err != nil {
		t.Fatal(err)
	}
	config.ConnConfig.RuntimeParams["search_path"] = schema
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	files, err := filepath.Glob("../../migrations/*.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		sql, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		conn, err := pool.Acquire(ctx)
		if err != nil {
			t.Fatal(err)
		}
		_, err = conn.Conn().Exec(ctx, string(sql), pgx.QueryExecModeSimpleProtocol)
		conn.Release()
		if err != nil {
			t.Fatalf("migration %s: %v", file, err)
		}
	}
	return pool
}

func TestPostgresConcurrentShortcodesAndRedisLifecycle(t *testing.T) {
	pool := integrationPool(t)
	ctx := context.Background()
	queries := db.New(pool)
	var cache *redis.Client
	if address := os.Getenv("TEST_REDIS_ADDR"); address != "" {
		cache = redis.NewClient(&redis.Options{Addr: address})
		if err := cache.Ping(ctx).Err(); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = cache.Close() })
	}
	svc := NewLinkService(pool, queries, cache, nil, createTestLogger())
	code := "samecode"
	const concurrent = 16
	var wg sync.WaitGroup
	results := make(chan error, concurrent)
	for range concurrent {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := svc.CreateShortLink(ctx, "user-a", "https://example.com", &code, nil)
			results <- err
		}()
	}
	wg.Wait()
	close(results)
	successes := 0
	for err := range results {
		if err == nil {
			successes++
			continue
		}
		if !errors.Is(err, apperrors.LinkShortcodeTaken) {
			t.Fatalf("unexpected conflict: %v", err)
		}
	}
	if successes != 1 {
		t.Fatalf("got %d successful inserts", successes)
	}
	owner, err := queries.GetLinkByShortcodeAndUser(ctx, db.GetLinkByShortcodeAndUserParams{Shortcode: code, UserID: "user-a"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = svc.GetLinkByShortcode(ctx, "user-b", code)
	if !errors.Is(err, apperrors.LinkNotFound) {
		t.Fatalf("other user fetched link: %v", err)
	}
	first, err := svc.GetOriginalURL(ctx, code)
	if err != nil {
		t.Fatal(err)
	}
	second, err := svc.GetOriginalURL(ctx, code)
	if err != nil || second.ID != first.ID || second.OriginalUrl != first.OriginalUrl {
		t.Fatalf("cached redirect mismatch: %+v %+v %v", first, second, err)
	}
	active := false
	if _, err := svc.UpdateLink(ctx, "user-a", owner.ID, nil, &active, false, nil); err != nil {
		t.Fatal(err)
	}
	_, err = svc.GetOriginalURL(ctx, code)
	if !errors.Is(err, apperrors.LinkNotFound) {
		t.Fatalf("disabled link redirects: %v", err)
	}
	if _, err := svc.DeleteLink(ctx, "user-a", owner.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.CreateShortLink(ctx, "user-a", "https://another.example", &code, nil); err != nil {
		t.Fatalf("soft-deleted code cannot be reused: %v", err)
	}
	generated1, err := svc.CreateShortLink(ctx, "user-a", "https://duplicate.example", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	generated2, err := svc.CreateShortLink(ctx, "user-a", "https://duplicate.example", nil, nil)
	if err != nil || generated1.Shortcode == generated2.Shortcode {
		t.Fatalf("duplicate URL retry failed: %s %s %v", generated1.Shortcode, generated2.Shortcode, err)
	}
	expires := time.Now().Add(100 * time.Millisecond)
	expiring, err := svc.CreateShortLink(ctx, "user-a", "https://expiring.example", nil, &expires)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.GetOriginalURL(ctx, expiring.Shortcode); err != nil {
		t.Fatal(err)
	}
	time.Sleep(150 * time.Millisecond)
	_, err = svc.GetOriginalURL(ctx, expiring.Shortcode)
	if !errors.Is(err, apperrors.LinkExpired) {
		t.Fatalf("expired link redirected: %v", err)
	}
}

func TestPostgresOwnershipTagsAndLinkUpdates(t *testing.T) {
	pool := integrationPool(t)
	ctx := context.Background()
	queries := db.New(pool)
	links := NewLinkService(pool, queries, nil, nil, createTestLogger())
	tags := NewTagService(queries, createTestLogger())
	tagA, err := tags.CreateTag(ctx, "user-a", "private-a")
	if err != nil {
		t.Fatal(err)
	}
	tagB, err := tags.CreateTag(ctx, "user-b", "private-b")
	if err != nil {
		t.Fatal(err)
	}
	linkA, err := links.CreateShortLink(ctx, "user-a", "https://example.com/a", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	linkB, err := links.CreateShortLink(ctx, "user-b", "https://example.com/b", nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := links.AddTagsToLink(ctx, "user-b", linkA.ID, []uuid.UUID{tagB.ID}); !errors.Is(err, apperrors.LinkNotFound) {
		t.Fatalf("other user assigned a tag to link: %v", err)
	}
	if _, err := links.AddTagsToLink(ctx, "user-a", linkA.ID, []uuid.UUID{tagB.ID}); !errors.Is(err, apperrors.TagNotFound) {
		t.Fatalf("other user's tag was assigned: %v", err)
	}
	if _, err := links.CreateShortLinkWithTags(ctx, "user-a", "https://example.com/rollback", nil, nil, []uuid.UUID{tagB.ID}); !errors.Is(err, apperrors.TagNotFound) {
		t.Fatalf("foreign tag did not roll back link creation: %v", err)
	}
	linksA, err := links.ListAllLinks(ctx, "user-a", nil, nil, 1, 5)
	if err != nil {
		t.Fatal(err)
	}
	if linksA.Total != 1 {
		t.Fatalf("failed create transaction left %d links", linksA.Total)
	}
	if _, err := links.AddTagsToLink(ctx, "user-a", linkA.ID, []uuid.UUID{tagA.ID, tagA.ID}); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := pool.QueryRow(ctx, "SELECT count(*) FROM link_tags WHERE link_id=$1", linkA.ID).Scan(&count); err != nil || count != 1 {
		t.Fatalf("duplicate tag assignment count=%d err=%v", count, err)
	}
	if _, err := links.RemoveTagsFromLink(ctx, "user-b", linkA.ID, []uuid.UUID{tagA.ID}); !errors.Is(err, apperrors.LinkNotFound) {
		t.Fatalf("other user removed a tag: %v", err)
	}
	if _, err := tags.UpdateTag(ctx, "user-b", tagA.ID, "stolen"); !errors.Is(err, apperrors.TagNotFound) {
		t.Fatalf("other user renamed a tag: %v", err)
	}
	if _, err := tags.DeleteTag(ctx, "user-b", tagA.ID); !errors.Is(err, apperrors.TagNotFound) {
		t.Fatalf("other user deleted a tag: %v", err)
	}
	if _, err := links.UpdateLink(ctx, "user-b", linkA.ID, nil, nil, false, nil); !errors.Is(err, apperrors.LinkNotFound) {
		t.Fatalf("other user updated a link: %v", err)
	}
	if _, err := links.DeleteLink(ctx, "user-b", linkA.ID); !errors.Is(err, apperrors.LinkNotFound) {
		t.Fatalf("other user deleted a link: %v", err)
	}
	conflict := linkB.Shortcode
	if _, err := links.UpdateLink(ctx, "user-a", linkA.ID, &conflict, nil, false, nil); !errors.Is(err, apperrors.LinkShortcodeTaken) {
		t.Fatalf("shortcode update conflict: %v", err)
	}
	expires := time.Now().Add(time.Hour)
	if _, err := links.UpdateLink(ctx, "user-a", linkA.ID, nil, nil, true, &expires); err != nil {
		t.Fatal(err)
	}
	cleared, err := links.UpdateLink(ctx, "user-a", linkA.ID, nil, nil, true, nil)
	if err != nil || cleared.ExpiresAt.Valid {
		t.Fatalf("expiration was not cleared: link=%+v err=%v", cleared, err)
	}
}
