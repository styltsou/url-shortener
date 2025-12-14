// API types matching Go backend DTOs

export interface Tag {
	id: string;
	name: string;
	created_at: string; // ISO date string
	updated_at?: string | null; // ISO date string
}

export interface Link {
	id: string;
	shortcode: string;
	original_url: string;
	user_id?: string;
	clicks?: number | null;
	expires_at: string | null; // ISO date string
	created_at: string; // ISO date string
	updated_at: string | null; // ISO date string
	is_active: boolean;
	tags?: Tag[]; // Optional - create response doesn't include tags
}

export interface CreateLinkRequest {
	url: string;
	shortcode?: string;
	expires_at?: string | null; // ISO 8601 datetime string
}

export interface UpdateLinkRequest {
	shortcode?: string;
	is_active?: boolean;
	expires_at?: string | null; // ISO date string or null
}

export interface PaginationMeta {
	page: number;
	limit: number;
	total: number;
	total_pages: number;
}

// SuccessResponse matches backend - data is required, pagination is optional
export interface SuccessResponse<T> {
	data: T;
	pagination?: PaginationMeta;
}

// PaginatedResponse is deprecated - use SuccessResponse instead
// Kept for backwards compatibility
export interface PaginatedResponse<T> {
	data: T;
	pagination: PaginationMeta;
}

import { generateMockAnalytics } from "@/lib/mock-data";
import type { Url } from "./url";

// Convert API Link to app Url type
export function linkToUrl(link: Link): Url {
	const url: Url = {
		id: link.id,
		originalUrl: link.original_url,
		shortCode: link.shortcode,
		createdAt: new Date(link.created_at),
		expiresAt: link.expires_at ? new Date(link.expires_at) : null,
		clicks: link.clicks || 0,
		tags:
			link.tags?.map((tag) => ({
				id: tag.id,
				name: tag.name,
			})) || [], // Handle missing tags (e.g., from create response)
		analytics: {
			clicks_data: [], // TODO: Fetch from analytics endpoint when available
			referrers_data: [], // TODO: Fetch from analytics endpoint when available
		},
		isActive: link.is_active !== false, // Default to true if not set
	};

	// Generate mock analytics if backend doesn't provide them
	if (
		url.analytics.clicks_data.length === 0 &&
		url.analytics.referrers_data.length === 0
	) {
		url.analytics = generateMockAnalytics(url, "7days");
	}

	return url;
}
