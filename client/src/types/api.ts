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

import type { Url } from "./url";

// Convert API Link to app Url type
export function linkToUrl(link: Link): Url {
	return {
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
			})) || [],
		isActive: link.is_active !== false,
	};
}
