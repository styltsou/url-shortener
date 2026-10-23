export type AnalyticsRange = "7d" | "30d" | "90d" | "12m";
export type AnalyticsInterval = "day" | "week" | "month";

export interface AnalyticsFilter {
	range: AnalyticsRange;
	interval: AnalyticsInterval;
}

export const DEFAULT_ANALYTICS_FILTER: AnalyticsFilter = { range: "30d", interval: "day" };

export const rangeLabels: Record<AnalyticsRange, string> = {
	"7d": "7 days",
	"30d": "30 days",
	"90d": "90 days",
	"12m": "12 months",
};

export const defaultIntervalForRange: Record<AnalyticsRange, AnalyticsInterval> = {
	"7d": "day",
	"30d": "day",
	"90d": "week",
	"12m": "month",
};

export function analyticsQueryString(filter: AnalyticsFilter): string {
	return new URLSearchParams({ range: filter.range, interval: filter.interval }).toString();
}
