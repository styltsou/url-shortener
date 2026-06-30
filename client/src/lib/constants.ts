/**
 * Application-wide constants
 */

// Pagination Defaults
export const DEFAULT_PAGE_SIZE = 5;
export const MIN_PAGE_SIZE = 5;
export const MAX_PAGE_SIZE = 50;

// Debounce Delays
export const DEBOUNCE_DELAY = 300;
export const SEARCH_DEBOUNCE_DELAY = 300;

// Date/Time Formats
export const DATE_FORMAT_OPTIONS: Intl.DateTimeFormatOptions = {
	month: "short",
	day: "numeric",
	year: "numeric",
};

export const DATETIME_FORMAT_OPTIONS: Intl.DateTimeFormatOptions = {
	month: "short",
	day: "numeric",
	year: "numeric",
	hour: "numeric",
	minute: "2-digit",
	hour12: true,
};

// UI Constants
export const MAX_VISIBLE_TAGS = 3;
export const EXPIRATION_WARNING_HOURS = 48; // 2 days
