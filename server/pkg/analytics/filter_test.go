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
		{"", "", Filter{Range30Days, IntervalDay}},
		{"7d", "", Filter{Range7Days, IntervalDay}},
		{"90d", "", Filter{Range90Days, IntervalWeek}},
		{"12m", "", Filter{Range12Months, IntervalMonth}},
		{"90d", "month", Filter{Range90Days, IntervalMonth}},
	} {
		got, err := ParseFilter(tc.rangeValue, tc.interval)
		if err != nil || got != tc.want {
			t.Errorf("ParseFilter(%q, %q) = %+v, %v; want %+v", tc.rangeValue, tc.interval, got, err, tc.want)
		}
	}
	for _, tc := range []struct{ rangeValue, interval string }{
		{"all", ""}, {"30d", "hour"}, {"30d", "WEEK"},
	} {
		if _, err := ParseFilter(tc.rangeValue, tc.interval); err == nil {
			t.Errorf("ParseFilter(%q, %q) should reject invalid input", tc.rangeValue, tc.interval)
		}
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
