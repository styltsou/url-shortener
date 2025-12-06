import type { Url, AnalyticsData, Tag } from "@/types/url";

// Common referrer sources for mock data
const REFERRER_SOURCES = [
	"Direct/None",
	"Google Search",
	"Twitter.com",
	"LinkedIn.com",
	"Facebook.com",
	"Reddit.com",
	"Newsletter",
	"Email",
	"Other",
];

/**
 * Generate mock clicks data for a time range
 */
function generateClicksData(
	days: number,
	totalClicks: number,
	linkAge: number // days since creation
): Array<{ name: string; clicks: number }> {
	const data: Array<{ name: string; clicks: number }> = [];
	const actualDays = Math.min(days, linkAge || days);
	const avgClicksPerDay = totalClicks / Math.max(actualDays, 1);

	// Generate day labels
	const dayLabels =
		days === 7
			? ["Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"]
			: Array.from({ length: days }, (_, i) => {
					const date = new Date();
					date.setDate(date.getDate() - (days - i - 1));
					return date.toLocaleDateString("en-US", {
						month: "short",
						day: "numeric",
					});
			  });

	for (let i = 0; i < days; i++) {
		// Add some randomness and variation
		const baseClicks = avgClicksPerDay * (0.7 + Math.random() * 0.6); // 70-130% of average
		// Weekend effect (lower traffic on weekends)
		const isWeekend = i % 7 >= 5;
		const weekendFactor = isWeekend ? 0.6 : 1.0;
		// Trend: slightly increasing over time
		const trendFactor = 1 + (i / days) * 0.2;

		// Ensure clicks is always an integer (can't have fractional clicks)
		const clicks = Math.max(
			0,
			Math.floor(baseClicks * weekendFactor * trendFactor)
		);
		data.push({
			name: dayLabels[i] || `Day ${i + 1}`,
			clicks,
		});
	}

	return data;
}

/**
 * Generate mock referrers data based on total clicks
 */
function generateReferrersData(
	totalClicks: number
): Array<{ referrer: string; clicks: number }> {
	// Distribution percentages (should sum to ~100%)
	const distribution = [
		{ referrer: "Direct/None", weight: 0.4 },
		{ referrer: "x.com", weight: 0.2 },
		{ referrer: "instagram.com", weight: 0.15 },
		{ referrer: "facebook.com", weight: 0.12 },
		{ referrer: "linkedin.com", weight: 0.08 },
		{ referrer: "reddit.com", weight: 0.05 },
	];

	const data = distribution
		.map(({ referrer, weight }) => ({
			referrer,
			clicks: Math.floor(totalClicks * weight * (0.8 + Math.random() * 0.4)), // Add some variance, ensure integer
		}))
		.filter((item) => item.clicks > 0) // Remove zero-click referrers
		.sort((a, b) => b.clicks - a.clicks); // Sort by clicks descending

	return data;
}

/**
 * Generate mock analytics data for a link
 */
export function generateMockAnalytics(
	link: Pick<Url, "clicks" | "createdAt">,
	timeRange: "7days" | "30days" = "7days"
): AnalyticsData {
	const days = timeRange === "7days" ? 7 : 30;
	const linkAge = Math.floor(
		(Date.now() - link.createdAt.getTime()) / (1000 * 60 * 60 * 24)
	);

	// Use actual clicks if available, otherwise generate based on link age
	const totalClicks =
		link.clicks > 0
			? link.clicks
			: Math.max(10, Math.floor(linkAge * (5 + Math.random() * 10)));

	return {
		clicks_data: generateClicksData(days, totalClicks, linkAge),
		referrers_data: generateReferrersData(totalClicks),
	};
}

// Legacy export for backwards compatibility
export const MOCK_ANALYTICS: AnalyticsData = {
	clicks_data: [
		{ name: "Mon", clicks: 120 },
		{ name: "Tue", clicks: 230 },
		{ name: "Wed", clicks: 180 },
		{ name: "Thu", clicks: 340 },
		{ name: "Fri", clicks: 290 },
		{ name: "Sat", clicks: 150 },
		{ name: "Sun", clicks: 190 },
	],
	referrers_data: [
		{ referrer: "Direct/None", clicks: 450 },
		{ referrer: "Google Search", clicks: 320 },
		{ referrer: "Twitter.com", clicks: 180 },
		{ referrer: "Linkedin.com", clicks: 90 },
		{ referrer: "Newsletter", clicks: 45 },
	],
};

// Mock tags database - must be defined before INITIAL_MOCK_URLS
export const MOCK_TAGS: Tag[] = [
	{ id: "1", name: "marketing" },
	{ id: "2", name: "product" },
	{ id: "3", name: "documentation" },
	{ id: "4", name: "social" },
	{ id: "5", name: "internal" },
	{ id: "6", name: "external" },
	{ id: "7", name: "promo" },
	{ id: "8", name: "blog" },
	{ id: "9", name: "urgent" },
	{ id: "10", name: "featured" },
	{ id: "11", name: "tutorial" },
	{ id: "12", name: "api" },
];

export const INITIAL_MOCK_URLS: Url[] = [
	{
		id: "1",
		originalUrl: "https://www.example.com/very/long/path/to/product/launch-v2",
		shortCode: "launch24",
		createdAt: new Date(Date.now() - 1000 * 60 * 60 * 24 * 2),
		expiresAt: null,
		clicks: 1245,
		tags: [
			MOCK_TAGS[0],
			MOCK_TAGS[1],
			MOCK_TAGS[6],
			MOCK_TAGS[9],
			MOCK_TAGS[10],
		], // marketing, product, promo, urgent, featured
		analytics: generateMockAnalytics(
			{
				clicks: 1245,
				createdAt: new Date(Date.now() - 1000 * 60 * 60 * 24 * 2),
			},
			"7days"
		),
	},
	{
		id: "2",
		originalUrl: "https://golang.org/doc/tutorial/getting-started",
		shortCode: "goLang",
		createdAt: new Date(Date.now() - 1000 * 60 * 60 * 5),
		expiresAt: new Date(Date.now() + 1000 * 60 * 60 * 24 * 5),
		clicks: 42,
		tags: [MOCK_TAGS[2], MOCK_TAGS[4], MOCK_TAGS[11]], // documentation, internal, tutorial
		analytics: generateMockAnalytics(
			{
				clicks: 42,
				createdAt: new Date(Date.now() - 1000 * 60 * 60 * 5),
			},
			"7days"
		),
	},
	{
		id: "3",
		originalUrl: "https://react.dev/learn",
		shortCode: "react",
		createdAt: new Date(Date.now() - 1000 * 60 * 60 * 24 * 10),
		expiresAt: new Date(Date.now() - 1000 * 60 * 60 * 24),
		clicks: 890,
		tags: [
			MOCK_TAGS[2],
			MOCK_TAGS[5],
			MOCK_TAGS[7],
			MOCK_TAGS[11],
			MOCK_TAGS[12],
		], // documentation, external, blog, tutorial, api
		analytics: generateMockAnalytics(
			{
				clicks: 890,
				createdAt: new Date(Date.now() - 1000 * 60 * 60 * 24 * 10),
			},
			"7days"
		),
	},
	{
		id: "4",
		originalUrl: "https://example.com/old-campaign",
		shortCode: "oldcamp",
		createdAt: new Date(Date.now() - 1000 * 60 * 60 * 24 * 30),
		expiresAt: null,
		clicks: 523,
		isActive: false, // This link is inactive
		tags: [MOCK_TAGS[0], MOCK_TAGS[6]], // marketing, promo
		analytics: generateMockAnalytics(
			{
				clicks: 523,
				createdAt: new Date(Date.now() - 1000 * 60 * 60 * 24 * 30),
			},
			"7days"
		),
	},
];

export const generateShortCode = () =>
	Math.random().toString(36).substring(2, 8);

export const formatDate = (date: Date | null) => {
	if (!date) return null;
	return new Date(date).toLocaleDateString("en-US", {
		month: "short",
		day: "numeric",
		year: "numeric",
	});
};
