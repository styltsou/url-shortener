import type { Url, Tag } from "@/types/url";

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
		tags: [MOCK_TAGS[0], MOCK_TAGS[1], MOCK_TAGS[6], MOCK_TAGS[9], MOCK_TAGS[10]],
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
		tags: [MOCK_TAGS[2], MOCK_TAGS[5], MOCK_TAGS[7], MOCK_TAGS[11], MOCK_TAGS[12]],
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
