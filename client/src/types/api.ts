// API types matching Go backend DTOs

export interface Link {
  id: string
  shortcode: string
  original_url: string
  user_id: string
  clicks: number | null
  expires_at: string | null // ISO date string
  created_at: string // ISO date string
  updated_at: string // ISO date string
}

export interface CreateLinkRequest {
  url: string
}

export interface UpdateLinkRequest {
  shortcode?: string
  expires_at?: string | null // ISO date string or null
}

import { generateMockAnalytics } from '@/lib/mock-data'
import type { Url } from './url'

// Convert API Link to app Url type
export function linkToUrl(link: Link): Url {
  const url: Url = {
    id: link.id,
    originalUrl: link.original_url,
    shortCode: link.shortcode,
    createdAt: new Date(link.created_at),
    expiresAt: link.expires_at ? new Date(link.expires_at) : null,
    clicks: link.clicks || 0,
    analytics: {
      clicks_data: [], // TODO: Fetch from analytics endpoint when available
      referrers_data: [], // TODO: Fetch from analytics endpoint when available
    },
  }

  // Generate mock analytics if backend doesn't provide them
  if (url.analytics.clicks_data.length === 0 && url.analytics.referrers_data.length === 0) {
    url.analytics = generateMockAnalytics(url, '7days')
  }

  return url
}

