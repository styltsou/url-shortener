export type AnalyticsRange = "7d" | "30d" | "90d" | "12m" | "custom";
export type AnalyticsInterval = "day" | "week" | "month";

export interface AnalyticsFilter {
	range: AnalyticsRange;
	interval: AnalyticsInterval;
	from?: string;
	to?: string;
}

export const DEFAULT_ANALYTICS_FILTER: AnalyticsFilter = { range: "30d", interval: "day" };

const rangeLabels: Record<Exclude<AnalyticsRange, "custom">, string> = {
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
	custom: "day",
};

export function analyticsQueryString(filter: AnalyticsFilter): string {
	const params = new URLSearchParams({ range: filter.range, interval: filter.interval });
	if (filter.range === "custom" && filter.from && filter.to) {
		params.set("from", filter.from);
		params.set("to", filter.to);
	}
	return params.toString();
}

export function analyticsPeriodLabel(filter: AnalyticsFilter): string {
	if (filter.range !== "custom") return `last ${rangeLabels[filter.range]}`;
	const formatDate = (date: string) =>
		new Date(`${date}T00:00:00Z`).toLocaleDateString("en-US", {
			month: "short",
			day: "numeric",
			year: "numeric",
			timeZone: "UTC",
		});
	if (!filter.from || !filter.to) return "custom period";
	if (filter.from === filter.to) return formatDate(filter.from);
	return `${formatDate(filter.from)} – ${formatDate(filter.to)}`;
}
