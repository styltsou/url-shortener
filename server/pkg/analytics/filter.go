package analytics

import (
	"fmt"
	"time"
)

type Range string
type Interval string

const (
	Range7Days    Range = "7d"
	Range30Days   Range = "30d"
	Range90Days   Range = "90d"
	Range12Months Range = "12m"

	IntervalDay   Interval = "day"
	IntervalWeek  Interval = "week"
	IntervalMonth Interval = "month"
)

type Filter struct {
	Range    Range
	Interval Interval
}

func ParseFilter(rangeValue, intervalValue string) (Filter, error) {
	if rangeValue == "" {
		rangeValue = string(Range30Days)
	}
	filter := Filter{Range: Range(rangeValue), Interval: Interval(intervalValue)}
	switch filter.Range {
	case Range7Days, Range30Days:
		if filter.Interval == "" {
			filter.Interval = IntervalDay
		}
	case Range90Days:
		if filter.Interval == "" {
			filter.Interval = IntervalWeek
		}
	case Range12Months:
		if filter.Interval == "" {
			filter.Interval = IntervalMonth
		}
	default:
		return Filter{}, fmt.Errorf("unsupported analytics range %q", rangeValue)
	}
	if filter.Interval != IntervalDay && filter.Interval != IntervalWeek && filter.Interval != IntervalMonth {
		return Filter{}, fmt.Errorf("unsupported analytics interval %q", intervalValue)
	}
	return filter, nil
}

func (f Filter) Bounds(now time.Time) (time.Time, time.Time) {
	until := now.UTC()
	today := time.Date(until.Year(), until.Month(), until.Day(), 0, 0, 0, 0, time.UTC)
	switch f.Range {
	case Range7Days:
		return today.AddDate(0, 0, -6), until
	case Range90Days:
		return today.AddDate(0, 0, -89), until
	case Range12Months:
		month := time.Date(until.Year(), until.Month(), 1, 0, 0, 0, 0, time.UTC)
		return month.AddDate(0, -11, 0), until
	default:
		return today.AddDate(0, 0, -29), until
	}
}

// BucketClicks groups UTC calendar days and includes empty buckets in the window.
func (f Filter) BucketClicks(days []ClicksOverTime, since, until time.Time) []ClicksOverTime {
	counts := make(map[time.Time]int, len(days))
	for _, day := range days {
		counts[bucketStart(day.Date.UTC(), f.Interval)] += day.Clicks
	}
	result := make([]ClicksOverTime, 0, len(days))
	for bucket := bucketStart(since.UTC(), f.Interval); !bucket.After(until.UTC()); bucket = nextBucket(bucket, f.Interval) {
		result = append(result, ClicksOverTime{Date: bucket, Clicks: counts[bucket]})
	}
	return result
}

func bucketStart(date time.Time, interval Interval) time.Time {
	date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	switch interval {
	case IntervalWeek:
		return date.AddDate(0, 0, -((int(date.Weekday()) + 6) % 7))
	case IntervalMonth:
		return time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, time.UTC)
	default:
		return date
	}
}

func nextBucket(date time.Time, interval Interval) time.Time {
	switch interval {
	case IntervalWeek:
		return date.AddDate(0, 0, 7)
	case IntervalMonth:
		return date.AddDate(0, 1, 0)
	default:
		return date.AddDate(0, 0, 1)
	}
}
