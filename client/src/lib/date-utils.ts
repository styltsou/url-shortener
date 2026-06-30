import { DATE_FORMAT_OPTIONS, DATETIME_FORMAT_OPTIONS } from "./constants";

export const formatDate = (date: Date | null) => {
	if (!date) return null;
	return new Date(date).toLocaleDateString("en-US", DATE_FORMAT_OPTIONS);
};

export const formatDateTime = (date: Date | string | null) => {
	if (!date) return null;
	const d = new Date(date);
	return d.toLocaleString("en-US", DATETIME_FORMAT_OPTIONS);
};

export const getTimePeriod = (date: Date | string, referenceDate: Date = new Date()): string => {
	const targetDate = new Date(date);
	const now = new Date(referenceDate);
	const diffMs = targetDate.getTime() - now.getTime();
	const diffSeconds = Math.floor(Math.abs(diffMs) / 1000);
	const isPast = diffMs < 0;

	if (diffSeconds < 60) {
		return isPast ? "just now" : "in a moment";
	}

	const diffMinutes = Math.floor(diffSeconds / 60);
	if (diffMinutes < 60) {
		const period = diffMinutes === 1 ? "minute" : "minutes";
		return isPast ? `${diffMinutes} ${period} ago` : `in ${diffMinutes} ${period}`;
	}

	const diffHours = Math.floor(diffSeconds / 3600);
	if (diffHours < 24) {
		const period = diffHours === 1 ? "hour" : "hours";
		return isPast ? `${diffHours} ${period} ago` : `in ${diffHours} ${period}`;
	}

	const diffDays = Math.floor(diffSeconds / 86400);
	if (diffDays < 7) {
		const period = diffDays === 1 ? "day" : "days";
		return isPast ? `${diffDays} ${period} ago` : `in ${diffDays} ${period}`;
	}

	const diffWeeks = Math.floor(diffDays / 7);
	if (diffWeeks < 4) {
		const period = diffWeeks === 1 ? "week" : "weeks";
		return isPast ? `${diffWeeks} ${period} ago` : `in ${diffWeeks} ${period}`;
	}

	const diffMonths = Math.floor(diffDays / 30);
	if (diffMonths < 12) {
		const period = diffMonths === 1 ? "month" : "months";
		return isPast ? `${diffMonths} ${period} ago` : `in ${diffMonths} ${period}`;
	}

	const diffYears = Math.floor(diffDays / 365);
	const period = diffYears === 1 ? "year" : "years";
	return isPast ? `${diffYears} ${period} ago` : `in ${diffYears} ${period}`;
};
