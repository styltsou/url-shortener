// API types matching Go backend DTOs

export interface Link {
	id: string;
	shortcode: string;
	original_url: string;
	user_id: string;
	clicks: number | null;
	expires_at: string | null; // ISO date string
	created_at: string; // ISO date string
	updated_at: string; // ISO date string
	is_active?: boolean; // TODO: Add when backend supports it
}

export interface CreateLinkRequest {
	url: string;
}

export interface UpdateLinkRequest {
	shortcode?: string;
	expires_at?: string | null; // ISO date string or null
}

import { generateMockAnalytics, MOCK_TAGS } from "@/lib/mock-data";
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
		tags: [], // TODO: Fetch tags from backend when available
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

	// Add mock tags temporarily for UI development
	// TODO: Remove when backend provides tags
	const tagAssignments: Record<string, number[]> = {
		// Assign tags based on link ID or shortcode
		"1": [0, 1, 6, 8, 9], // marketing, product, promo, urgent, featured
		"2": [2, 4, 10], // documentation, internal, tutorial
		"3": [2, 5, 7, 10, 11], // documentation, external, blog, tutorial, api
	};

	// Try to match by ID first, then by shortcode pattern
	const linkId = link.id;
	const shortcode = link.shortcode.toLowerCase();

	let tagIndices: number[] = [];
	if (tagAssignments[linkId]) {
		tagIndices = tagAssignments[linkId];
	} else if (shortcode.includes("launch") || shortcode.includes("promo")) {
		tagIndices = [0, 1, 6, 8, 9]; // marketing, product, promo, urgent, featured
	} else if (
		shortcode.includes("go") ||
		shortcode.includes("lang") ||
		shortcode.includes("doc")
	) {
		tagIndices = [2, 4, 10]; // documentation, internal, tutorial
	} else if (
		shortcode.includes("react") ||
		shortcode.includes("api") ||
		shortcode.includes("learn")
	) {
		tagIndices = [2, 5, 7, 10, 11]; // documentation, external, blog, tutorial, api
	} else {
		// Default: assign 2-3 random tags
		const randomIndices = new Set<number>();
		while (randomIndices.size < Math.floor(Math.random() * 2) + 2) {
			randomIndices.add(Math.floor(Math.random() * MOCK_TAGS.length));
		}
		tagIndices = Array.from(randomIndices);
	}

	url.tags = tagIndices.map((idx) => MOCK_TAGS[idx]).filter(Boolean);

	return url;
}
