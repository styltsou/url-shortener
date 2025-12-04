export interface AnalyticsData {
  clicks_data: Array<{ name: string; clicks: number }>
  referrers_data: Array<{ referrer: string; clicks: number }>
}

export interface Url {
  id: string
  originalUrl: string
  shortCode: string
  createdAt: Date
  expiresAt: Date | null
  clicks: number
  analytics: AnalyticsData
}

