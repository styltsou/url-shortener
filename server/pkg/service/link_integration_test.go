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
	"github.com/styltsou/url-shortener/server/pkg/dto"
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

func TestPostgresCreateOrFindShortLink(t *testing.T) {
	pool := integrationPool(t)
	ctx := context.Background()
	queries := db.New(pool)
	links := NewLinkService(pool, queries, nil, nil, createTestLogger())
	tags := NewTagService(queries, createTestLogger())
	if _, _, err := links.CreateOrFindShortLinkWithTags(ctx, "user-a", "not-a-url", nil, nil, nil, false); !errors.Is(err, apperrors.InvalidURL) {
		t.Fatalf("invalid URL: %v", err)
	}
	tag, err := tags.CreateTag(ctx, "user-a", "campaign")
	if err != nil {
		t.Fatal(err)
	}
	url := "https://example.com/repeated"
	first, existing, err := links.CreateOrFindShortLinkWithTags(ctx, "user-a", url, nil, nil, nil, false)
	if err != nil || existing {
		t.Fatalf("first create: existing=%v err=%v", existing, err)
	}
	expires := time.Now().Add(time.Hour)
	again, existing, err := links.CreateOrFindShortLinkWithTags(ctx, "user-a", url, nil, &expires, []uuid.UUID{tag.ID}, false)
	if err != nil {
		t.Fatal(err)
	}
	againResponse, err := dto.LinkFromWithTagsRow(again)
	if err != nil {
		t.Fatal(err)
	}
	if !existing || again.ID != first.ID || again.ExpiresAt.Valid || len(againResponse.Tags) != 0 {
		t.Fatalf("repeat changed existing link: first=%+v again=%+v existing=%v err=%v", first, again, existing, err)
	}
	forced, existing, err := links.CreateOrFindShortLinkWithTags(ctx, "user-a", url, nil, &expires, []uuid.UUID{tag.ID}, true)
	if err != nil {
		t.Fatal(err)
	}
	forcedResponse, err := dto.LinkFromWithTagsRow(forced)
	if err != nil {
		t.Fatal(err)
	}
	if existing || forced.ID == first.ID || !forced.ExpiresAt.Valid || len(forcedResponse.Tags) != 1 || forcedResponse.Tags[0].ID != tag.ID {
		t.Fatalf("force new: link=%+v existing=%v err=%v", forced, existing, err)
	}
	other, existing, err := links.CreateOrFindShortLinkWithTags(ctx, "user-b", url, nil, nil, nil, false)
	if err != nil || existing || other.ID == first.ID {
		t.Fatalf("other user: link=%+v existing=%v err=%v", other, existing, err)
	}
	alias := "my-campaign"
	custom, existing, err := links.CreateOrFindShortLinkWithTags(ctx, "user-a", url, &alias, nil, nil, false)
	if err != nil || existing || custom.Shortcode != alias {
		t.Fatalf("custom alias: link=%+v existing=%v err=%v", custom, existing, err)
	}

	for _, tc := range []struct {
		name   string
		url    string
		change func(uuid.UUID) error
	}{
		{"inactive", "https://example.com/inactive", func(id uuid.UUID) error {
			active := false
			_, err := links.UpdateLink(ctx, "user-a", id, nil, &active, false, nil)
			return err
		}},
		{"deleted", "https://example.com/deleted", func(id uuid.UUID) error {
			_, err := links.DeleteLink(ctx, "user-a", id)
			return err
		}},
		{"expired", "https://example.com/expired", func(id uuid.UUID) error {
			_, err := pool.Exec(ctx, "UPDATE links SET expires_at=NOW() - INTERVAL '1 second' WHERE id=$1", id)
			return err
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			initial, _, err := links.CreateOrFindShortLinkWithTags(ctx, "user-a", tc.url, nil, nil, nil, false)
			if err != nil {
				t.Fatal(err)
			}
			if err := tc.change(initial.ID); err != nil {
				t.Fatal(err)
			}
			next, existing, err := links.CreateOrFindShortLinkWithTags(ctx, "user-a", tc.url, nil, nil, nil, false)
			if err != nil || existing || next.ID == initial.ID {
				t.Fatalf("should create after %s: next=%+v existing=%v err=%v", tc.name, next, existing, err)
			}
		})
	}

	const concurrent = 8
	results := make(chan struct {
		id       uuid.UUID
		existing bool
		err      error
	}, concurrent)
	var wg sync.WaitGroup
	for range concurrent {
		wg.Add(1)
		go func() {
			defer wg.Done()
			link, existing, err := links.CreateOrFindShortLinkWithTags(ctx, "user-a", "https://example.com/concurrent", nil, nil, nil, false)
			results <- struct {
				id       uuid.UUID
				existing bool
				err      error
			}{link.ID, existing, err}
		}()
	}
	wg.Wait()
	close(results)
	var firstID uuid.UUID
	created := 0
	for result := range results {
		if result.err != nil {
			t.Fatal(result.err)
		}
		if !result.existing {
			created++
		}
		if firstID == uuid.Nil {
			firstID = result.id
		} else if result.id != firstID {
			t.Fatalf("concurrent requests returned different links: %s, %s", firstID, result.id)
		}
	}
	if created != 1 {
		t.Fatalf("created %d links for concurrent requests", created)
	}
}

func TestPostgresShortcodeAvailability(t *testing.T) {
	pool := integrationPool(t)
	ctx := context.Background()
	links := NewLinkService(pool, db.New(pool), nil, nil, createTestLogger())
	for _, code := range []string{"", "api", "bad alias", "123456789012345678901"} {
		if _, err := links.ShortcodeAvailable(ctx, code); !errors.Is(err, apperrors.InvalidShortcode) {
			t.Fatalf("invalid alias %q: %v", code, err)
		}
	}
	code := "campaign"
	if available, err := links.ShortcodeAvailable(ctx, code); err != nil || !available {
		t.Fatalf("new alias: available=%v err=%v", available, err)
	}
	link, err := links.CreateShortLink(ctx, "user-a", "https://example.com", &code, nil)
	if err != nil {
		t.Fatal(err)
	}
	if available, err := links.ShortcodeAvailable(ctx, code); err != nil || available {
		t.Fatalf("occupied alias: available=%v err=%v", available, err)
	}
	active := false
	if _, err := links.UpdateLink(ctx, "user-a", link.ID, nil, &active, false, nil); err != nil {
		t.Fatal(err)
	}
	if available, err := links.ShortcodeAvailable(ctx, code); err != nil || available {
		t.Fatalf("inactive link released alias: available=%v err=%v", available, err)
	}
	if _, err := pool.Exec(ctx, "UPDATE links SET expires_at=NOW() - INTERVAL '1 second' WHERE id=$1", link.ID); err != nil {
		t.Fatal(err)
	}
	if available, err := links.ShortcodeAvailable(ctx, code); err != nil || available {
		t.Fatalf("expired link released alias: available=%v err=%v", available, err)
	}
	if _, err := links.DeleteLink(ctx, "user-a", link.ID); err != nil {
		t.Fatal(err)
	}
	if available, err := links.ShortcodeAvailable(ctx, code); err != nil || !available {
		t.Fatalf("deleted link kept alias: available=%v err=%v", available, err)
	}
}
