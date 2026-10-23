package analytics

import (
	"testing"
	"time"
)

func TestParseFilterDefaultsAndValidation(t *testing.T) {
	for _, tc := range []struct {
		rangeValue string
		interval   string
		want       Filter
	}{
		{"", "", Filter{Range: Range30Days, Interval: IntervalDay}},
		{"7d", "", Filter{Range: Range7Days, Interval: IntervalDay}},
		{"90d", "", Filter{Range: Range90Days, Interval: IntervalWeek}},
		{"12m", "", Filter{Range: Range12Months, Interval: IntervalMonth}},
		{"90d", "month", Filter{Range: Range90Days, Interval: IntervalMonth}},
	} {
		got, err := ParseFilter(tc.rangeValue, tc.interval, "", "")
		if err != nil || got != tc.want {
			t.Errorf("ParseFilter(%q, %q) = %+v, %v; want %+v", tc.rangeValue, tc.interval, got, err, tc.want)
		}
	}
	for _, tc := range []struct{ rangeValue, interval, from, to string }{
		{"all", "", "", ""},
		{"30d", "hour", "", ""},
		{"30d", "WEEK", "", ""},
		{"30d", "day", "2026-10-01", "2026-10-02"},
		{"custom", "", "", "2026-10-02"},
		{"custom", "", "2026-02-30", "2026-03-01"},
		{"custom", "", "2026-10-02", "2026-10-01"},
		{"custom", "", "2025-01-01", "2026-10-01"},
	} {
		if _, err := ParseFilter(tc.rangeValue, tc.interval, tc.from, tc.to); err == nil {
			t.Errorf("ParseFilter(%q, %q, %q, %q) should reject invalid input", tc.rangeValue, tc.interval, tc.from, tc.to)
		}
	}
}

func TestCustomFilterIncludesBothDatesAndSupportsOneDay(t *testing.T) {
	filter, err := ParseFilter("custom", "", "2026-10-20", "2026-10-20")
	if err != nil || filter.Interval != IntervalDay {
		t.Fatalf("ParseFilter custom day = %+v, %v", filter, err)
	}
	since, until := filter.Bounds(time.Time{})
	if since.Format(time.DateOnly) != "2026-10-20" || until.Format(time.RFC3339Nano) != "2026-10-20T23:59:59.999999999Z" {
		t.Fatalf("custom day bounds = %v to %v", since, until)
	}
	rows := []ClicksOverTime{{Date: since, Clicks: 3}}
	if buckets := filter.BucketClicks(rows, since, until); len(buckets) != 1 || buckets[0].Clicks != 3 {
		t.Fatalf("custom day buckets = %+v", buckets)
	}
	filter, err = ParseFilter("custom", "week", "2026-10-20", "2026-11-03")
	if err != nil {
		t.Fatal(err)
	}
	since, until = filter.Bounds(time.Time{})
	if buckets := filter.BucketClicks(nil, since, until); len(buckets) != 3 {
		t.Fatalf("custom weekly buckets = %+v", buckets)
	}
}

func TestFilterBoundsAndCalendarBuckets(t *testing.T) {
	now := time.Date(2026, time.November, 2, 12, 0, 0, 0, time.UTC)
	filter := Filter{Range: Range90Days, Interval: IntervalWeek}
	since, until := filter.Bounds(now)
	firstDay := time.Date(2026, time.November, 2, 0, 0, 0, 0, time.UTC).AddDate(0, 0, -89)
	if !since.Equal(firstDay) || !until.Equal(now) {
		t.Fatalf("Bounds() = %v to %v, want 90 calendar days ending at %v", since, until, now)
	}
	monthFilter := Filter{Range: Range12Months, Interval: IntervalMonth}
	monthStart, _ := monthFilter.Bounds(now)
	if want := time.Date(2025, time.December, 1, 0, 0, 0, 0, time.UTC); !monthStart.Equal(want) {
		t.Errorf("12-month start = %v, want %v", monthStart, want)
	}
	windowStart := time.Date(2026, time.October, 30, 12, 0, 0, 0, time.UTC)
	rows := []ClicksOverTime{
		{Date: time.Date(2026, time.October, 30, 0, 0, 0, 0, time.UTC), Clicks: 2},
		{Date: time.Date(2026, time.November, 2, 0, 0, 0, 0, time.UTC), Clicks: 3},
	}
	weekly := filter.BucketClicks(rows, windowStart, now)
	if len(weekly) != 2 || weekly[0].Clicks != 2 || weekly[1].Clicks != 3 || weekly[0].Date.Weekday() != time.Monday {
		t.Fatalf("weekly buckets = %+v", weekly)
	}
	filter.Interval = IntervalMonth
	monthly := filter.BucketClicks(rows, windowStart, now)
	if len(monthly) != 2 || monthly[0].Clicks != 2 || monthly[1].Clicks != 3 || monthly[0].Date.Day() != 1 {
		t.Fatalf("monthly buckets = %+v", monthly)
	}
	filter.Interval = IntervalDay
	daily := filter.BucketClicks(rows, windowStart, now)
	if len(daily) != 4 || daily[1].Clicks != 0 || daily[2].Clicks != 0 {
		t.Fatalf("daily buckets = %+v", daily)
	}
}
