import type { Url, Tag } from "@/types/url";
import { DATE_FORMAT_OPTIONS, DATETIME_FORMAT_OPTIONS } from "./constants";

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
		],
	},
	{
		id: "2",
		originalUrl: "https://golang.org/doc/tutorial/getting-started",
		shortCode: "goLang",
		createdAt: new Date(Date.now() - 1000 * 60 * 60 * 5),
		expiresAt: new Date(Date.now() + 1000 * 60 * 60 * 24 * 5),
		clicks: 42,
		tags: [MOCK_TAGS[2], MOCK_TAGS[4], MOCK_TAGS[11]],
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
		],
	},
	{
		id: "4",
		originalUrl: "https://example.com/old-campaign",
		shortCode: "oldcamp",
		createdAt: new Date(Date.now() - 1000 * 60 * 60 * 24 * 30),
		expiresAt: null,
		clicks: 523,
		isActive: false,
		tags: [MOCK_TAGS[0], MOCK_TAGS[6]],
	},
];

export const generateShortCode = () =>
	Math.random().toString(36).substring(2, 8);

export const formatDate = (date: Date | null) => {
	if (!date) return null;
	return new Date(date).toLocaleDateString("en-US", DATE_FORMAT_OPTIONS);
};

export const formatDateTime = (date: Date | string | null) => {
	if (!date) return null;
	const d = new Date(date);
	return d.toLocaleString("en-US", DATETIME_FORMAT_OPTIONS);
};

/**
 * Get a human-readable relative time period (e.g., "2 hours", "3 days")
 * @param date - The date to compare
 * @param referenceDate - Optional reference date (defaults to now)
 * @returns A string like "2 hours", "3 days", "in 2 hours", or "2 hours ago"
 */
export const getTimePeriod = (
	date: Date | string,
	referenceDate: Date = new Date()
): string => {
	const targetDate = new Date(date);
	const now = new Date(referenceDate);
	const diffMs = targetDate.getTime() - now.getTime();
	const diffSeconds = Math.floor(Math.abs(diffMs) / 1000);
	const isPast = diffMs < 0;

	// Less than a minute
	if (diffSeconds < 60) {
		return isPast ? "just now" : "in a moment";
	}

	// Minutes
	const diffMinutes = Math.floor(diffSeconds / 60);
	if (diffMinutes < 60) {
		const period = diffMinutes === 1 ? "minute" : "minutes";
		return isPast
			? `${diffMinutes} ${period} ago`
			: `in ${diffMinutes} ${period}`;
	}

	// Hours
	const diffHours = Math.floor(diffSeconds / 3600);
	if (diffHours < 24) {
		const period = diffHours === 1 ? "hour" : "hours";
		return isPast ? `${diffHours} ${period} ago` : `in ${diffHours} ${period}`;
	}

	// Days
	const diffDays = Math.floor(diffSeconds / 86400);
	if (diffDays < 7) {
		const period = diffDays === 1 ? "day" : "days";
		return isPast ? `${diffDays} ${period} ago` : `in ${diffDays} ${period}`;
	}

	// Weeks
	const diffWeeks = Math.floor(diffDays / 7);
	if (diffWeeks < 4) {
		const period = diffWeeks === 1 ? "week" : "weeks";
		return isPast ? `${diffWeeks} ${period} ago` : `in ${diffWeeks} ${period}`;
	}

	// Months
	const diffMonths = Math.floor(diffDays / 30);
	if (diffMonths < 12) {
		const period = diffMonths === 1 ? "month" : "months";
		return isPast
			? `${diffMonths} ${period} ago`
			: `in ${diffMonths} ${period}`;
	}

	// Years
	const diffYears = Math.floor(diffDays / 365);
	const period = diffYears === 1 ? "year" : "years";
	return isPast ? `${diffYears} ${period} ago` : `in ${diffYears} ${period}`;
};
