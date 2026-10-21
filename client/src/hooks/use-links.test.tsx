// @vitest-environment jsdom
import { renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import type { ReactNode } from "react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { ApiError } from "@/lib/api-client";
import {
	useLink,
	useLinks,
	useCreateLink,
	useDeleteLink,
	useShortcodeAvailability,
	linkKeys,
} from "./use-links";
import { useDashboard, dashboardKeys } from "./use-dashboard";

const { get, post, remove, auth } = vi.hoisted(() => ({
	get: vi.fn(),
	post: vi.fn(),
	remove: vi.fn(),
	auth: { isLoaded: true, isSignedIn: true, userId: "user-a", getToken: async () => "test-token" },
}));
vi.mock("@clerk/clerk-react", () => ({ useAuth: () => auth }));
vi.mock("@/lib/api-client", async (importOriginal) => {
	const original = await importOriginal<typeof import("@/lib/api-client")>();
	return { ...original, apiClient: { get, post, delete: remove } };
});

const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
function wrapper({ children }: { children: ReactNode }) {
	return <QueryClientProvider client={client}>{children}</QueryClientProvider>;
}

describe("useLink shortcode contract", () => {
	beforeEach(() => {
		get.mockReset();
		post.mockReset();
		remove.mockReset();
		client.clear();
		auth.isLoaded = true;
		auth.isSignedIn = true;
		auth.userId = "user-a";
	});

	it("loads a link directly, regardless of list pagination", async () => {
		get.mockResolvedValue({
			data: {
				id: "link-id",
				shortcode: "older-link",
				original_url: "https://example.com",
				expires_at: null,
				created_at: "2026-01-01T00:00:00Z",
				updated_at: null,
				is_active: true,
				tags: [],
			},
		});
		const { result } = renderHook(() => useLink("older-link"), { wrapper });
		await waitFor(() => expect(result.current.data?.id).toBe("link-id"));
		expect(get).toHaveBeenCalledWith("/api/v1/links/older-link", "test-token");
		expect(result.current.data?.shortCode).toBe("older-link");
	});

	it("reports a missing link from the detail endpoint", async () => {
		get.mockRejectedValue(new ApiError(404, "link_not_found"));
		const { result } = renderHook(() => useLink("missing"), { wrapper });
		await waitFor(() => expect(result.current.isError).toBe(true));
		expect((result.current.error as ApiError).status).toBe(404);
		expect(get).toHaveBeenCalledWith("/api/v1/links/missing", "test-token");
	});

	it("waits for Clerk and keeps each user's link data separate", async () => {
		get.mockImplementation(async () => ({
			data: {
				id: auth.userId,
				shortcode: "shared",
				original_url: "https://example.com",
				expires_at: null,
				created_at: "2026-01-01T00:00:00Z",
				updated_at: null,
				is_active: true,
				tags: [],
			},
		}));
		auth.isLoaded = false;
		const { result, rerender } = renderHook(() => useLink("shared"), { wrapper });
		expect(get).not.toHaveBeenCalled();
		auth.isLoaded = true;
		rerender();
		await waitFor(() => expect(result.current.data?.id).toBe("user-a"));
		auth.userId = "user-b";
		rerender();
		expect(result.current.data?.id).not.toBe("user-a");
		await waitFor(() => expect(result.current.data?.id).toBe("user-b"));
		auth.isSignedIn = false;
		rerender();
		expect(result.current.data).toBeUndefined();
	});

	it("uses the effective page size in the cache key", async () => {
		get.mockResolvedValue({
			data: [],
			pagination: { page: 1, limit: 5, total: 0, total_pages: 0 },
		});
		const { rerender } = renderHook(({ limit }) => useLinks({ page: 1, limit }), {
			wrapper,
			initialProps: { limit: 5 },
		});
		await waitFor(() => expect(get).toHaveBeenCalledTimes(1));
		rerender({ limit: 10 });
		await waitFor(() => expect(get).toHaveBeenCalledTimes(2));
		expect(get.mock.calls[0][0]).toContain("limit=5");
		expect(get.mock.calls[1][0]).toContain("limit=10");
	});

	it("waits for Clerk before loading dashboard data", async () => {
		get.mockResolvedValue({
			data: {
				total_links: 1,
				active_links: 1,
				total_clicks: 0,
				clicks_over_time: [],
				recent_links: [],
			},
		});
		auth.isLoaded = false;
		const { result, rerender } = renderHook(() => useDashboard(), { wrapper });
		expect(get).not.toHaveBeenCalled();
		auth.isLoaded = true;
		rerender();
		await waitFor(() => expect(result.current.data?.total_links).toBe(1));
		auth.isSignedIn = false;
		rerender();
		expect(result.current.data).toBeUndefined();
	});

	it("invalidates every cached list page after deletion", async () => {
		remove.mockResolvedValue(undefined);
		const firstPage = [...linkKeys.list({ page: 1, limit: 5 }), "user-a"];
		const secondPage = [...linkKeys.list({ page: 2, limit: 10 }), "user-a"];
		client.setQueryData(firstPage, { urls: [] });
		client.setQueryData(secondPage, { urls: [] });
		const dashboard = [...dashboardKeys.all, "user-a"];
		client.setQueryData(dashboard, { total_links: 2 });
		const { result } = renderHook(() => useDeleteLink(), { wrapper });
		await result.current.mutateAsync("link-id");
		expect(client.getQueryState(firstPage)?.isInvalidated).toBe(true);
		expect(client.getQueryState(secondPage)?.isInvalidated).toBe(true);
		expect(client.getQueryState(dashboard)?.isInvalidated).toBe(true);
	});

	it("returns an existing link without invalidating the list", async () => {
		const page = [...linkKeys.list({ page: 1, limit: 5 }), "user-a"];
		client.setQueryData(page, { urls: [] });
		post.mockResolvedValue({
			existing: true,
			data: {
				id: "link-id",
				shortcode: "abc123",
				original_url: "https://example.com",
				expires_at: null,
				created_at: "2026-01-01T00:00:00Z",
				updated_at: null,
				is_active: true,
				tags: [],
			},
		});
		const { result } = renderHook(() => useCreateLink(), { wrapper });
		const response = await result.current.mutateAsync({ url: "https://example.com" });
		expect(response).toMatchObject({
			existing: true,
			link: { id: "link-id", shortCode: "abc123" },
		});
		expect(client.getQueryState(page)?.isInvalidated).toBe(false);
	});

	it("checks an encoded custom alias with the signed-in user's token", async () => {
		get.mockResolvedValue({ data: { available: false } });
		const { result } = renderHook(() => useShortcodeAvailability("my_alias"), { wrapper });
		await waitFor(() => expect(result.current.data).toBe(false));
		expect(get).toHaveBeenCalledWith(
			"/api/v1/shortcodes/availability?shortcode=my_alias",
			"test-token",
		);
	});
});
