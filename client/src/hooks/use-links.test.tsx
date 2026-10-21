// @vitest-environment jsdom
import { renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import type { ReactNode } from "react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { ApiError } from "@/lib/api-client";
import { useLink } from "./use-links";

const get = vi.hoisted(() => vi.fn());
vi.mock("@clerk/clerk-react", () => ({ useAuth: () => ({ getToken: async () => "test-token" }) }));
vi.mock("@/lib/api-client", async (importOriginal) => {
	const original = await importOriginal<typeof import("@/lib/api-client")>();
	return { ...original, apiClient: { get } };
});

const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
function wrapper({ children }: { children: ReactNode }) {
	return <QueryClientProvider client={client}>{children}</QueryClientProvider>;
}

describe("useLink shortcode contract", () => {
	beforeEach(() => {
		get.mockReset();
		client.clear();
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
});
