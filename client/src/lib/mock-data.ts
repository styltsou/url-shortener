import type { Url, AnalyticsData } from '@/types/url'

export const MOCK_ANALYTICS: AnalyticsData = {
  clicks_data: [
    { name: 'Mon', clicks: 120 },
    { name: 'Tue', clicks: 230 },
    { name: 'Wed', clicks: 180 },
    { name: 'Thu', clicks: 340 },
    { name: 'Fri', clicks: 290 },
    { name: 'Sat', clicks: 150 },
    { name: 'Sun', clicks: 190 },
  ],
  referrers_data: [
    { referrer: 'Direct/None', clicks: 450 },
    { referrer: 'Google Search', clicks: 320 },
    { referrer: 'Twitter.com', clicks: 180 },
    { referrer: 'Linkedin.com', clicks: 90 },
    { referrer: 'Newsletter', clicks: 45 },
  ],
}

export const INITIAL_MOCK_URLS: Url[] = [
  {
    id: '1',
    originalUrl: 'https://www.example.com/very/long/path/to/product/launch-v2',
    shortCode: 'launch24',
    createdAt: new Date(Date.now() - 1000 * 60 * 60 * 24 * 2),
    expiresAt: null,
    clicks: 1245,
    analytics: MOCK_ANALYTICS,
  },
  {
    id: '2',
    originalUrl: 'https://golang.org/doc/tutorial/getting-started',
    shortCode: 'goLang',
    createdAt: new Date(Date.now() - 1000 * 60 * 60 * 5),
    expiresAt: new Date(Date.now() + 1000 * 60 * 60 * 24 * 5),
    clicks: 42,
    analytics: MOCK_ANALYTICS,
  },
  {
    id: '3',
    originalUrl: 'https://react.dev/learn',
    shortCode: 'react',
    createdAt: new Date(Date.now() - 1000 * 60 * 60 * 24 * 10),
    expiresAt: new Date(Date.now() - 1000 * 60 * 60 * 24),
    clicks: 890,
    analytics: MOCK_ANALYTICS,
  },
]

export const generateShortCode = () => Math.random().toString(36).substring(2, 8)

export const formatDate = (date: Date | null) => {
  if (!date) return null
  return new Date(date).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })
}

