package analytics

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/styltsou/url-shortener/server/pkg/logger"
)

func TestNewReturnsNopClientWhenURLMissing(t *testing.T) {
	log := newTestLogger(t)

	client, err := New(context.Background(), Config{}, log)
	if err != nil {
		t.Fatalf("New() error = %v, want nil", err)
	}
	if _, ok := client.(*nopClient); !ok {
		t.Fatalf("New() client = %T, want *nopClient", client)
	}
	if _, err := client.GetLinkAnalytics(context.Background(), uuid.New(), time.Time{}, time.Now()); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("GetLinkAnalytics() error = %v, want ErrUnavailable", err)
	}
	if _, err := client.GetUserAnalytics(context.Background(), []uuid.UUID{uuid.New()}, time.Time{}, time.Now()); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("GetUserAnalytics() error = %v, want ErrUnavailable", err)
	}
	if err := client.Close(); err != nil {
		t.Fatalf("Close() error = %v, want nil", err)
	}
}

func newTestLogger(t *testing.T) logger.Logger {
	t.Helper()

	log, err := logger.New("test")
	if err != nil {
		t.Fatalf("logger.New() error = %v", err)
	}
	return log
}

func TestFillDailyClicksPreservesCalendarGaps(t *testing.T) {
	since := time.Date(2026, time.October, 30, 12, 0, 0, 0, time.UTC)
	until := time.Date(2026, time.November, 2, 12, 0, 0, 0, time.UTC)
	rows := []ClicksOverTime{
		{Date: time.Date(2026, time.October, 30, 0, 0, 0, 0, time.UTC), Clicks: 2},
		{Date: time.Date(2026, time.November, 1, 0, 0, 0, 0, time.UTC), Clicks: 3},
	}
	got := fillDailyClicks(rows, since, until)
	if len(got) != 4 {
		t.Fatalf("fillDailyClicks() returned %d days, want 4", len(got))
	}
	for i, want := range []int{2, 0, 3, 0} {
		if got[i].Clicks != want {
			t.Errorf("day %d clicks = %d, want %d", i, got[i].Clicks, want)
		}
		if got[i].Date.Location() != time.UTC || got[i].Date.Day() != []int{30, 31, 1, 2}[i] {
			t.Errorf("day %d date = %v, want UTC calendar sequence", i, got[i].Date)
		}
	}
}
