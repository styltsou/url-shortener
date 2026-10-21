// @vitest-environment jsdom
import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { NewUrlForm } from "./new-url-form";
import type { Url } from "@/types/url";

vi.mock("@/hooks/use-tags", () => ({
	useTags: () => ({ data: [], isLoading: false }),
	useCreateTag: () => ({ mutateAsync: vi.fn() }),
}));
vi.mock("@tanstack/react-router", () => ({
	Link: ({ children, params }: { children: React.ReactNode; params: { shortcode: string } }) => (
		<a href={`/links/${params.shortcode}`}>{children}</a>
	),
}));

const existingLink: Url = {
	id: "link-1",
	originalUrl: "https://example.com",
	shortCode: "abc123",
	createdAt: new Date("2026-01-01T00:00:00Z"),
	expiresAt: null,
	clicks: 0,
	tags: [],
	isActive: true,
};

describe("NewUrlForm duplicate choice", () => {
	beforeEach(() => vi.clearAllMocks());
	afterEach(() => cleanup());

	it("offers the existing link and only creates another after explicit choice", async () => {
		const onShorten = vi
			.fn()
			.mockResolvedValueOnce({ link: existingLink, existing: true })
			.mockResolvedValueOnce({
				link: { ...existingLink, id: "link-2", shortCode: "xyz789" },
				existing: false,
			});
		render(<NewUrlForm onShorten={onShorten} isLoading={false} />);
		fireEvent.change(screen.getByPlaceholderText("Paste a long URL here..."), {
			target: { value: "https://example.com" },
		});
		fireEvent.click(screen.getByRole("button", { name: /Shorten/ }));
		await screen.findByText("You already have a link for this URL");
		expect(onShorten).toHaveBeenCalledTimes(1);
		expect(onShorten).toHaveBeenCalledWith({ url: "https://example.com" });
		expect(screen.getByRole("link", { name: "Open existing link" }).getAttribute("href")).toBe(
			"/links/abc123",
		);
		fireEvent.click(screen.getByRole("button", { name: "Create another link anyway" }));
		await waitFor(() =>
			expect(onShorten).toHaveBeenCalledWith({ url: "https://example.com", force_new: true }),
		);
		await waitFor(() =>
			expect(screen.queryByText("You already have a link for this URL")).toBeNull(),
		);
		expect(
			(screen.getByPlaceholderText("Paste a long URL here...") as HTMLInputElement).value,
		).toBe("");
	});

	it("keeps the form and does not create a second link when the choice is dismissed", async () => {
		const onShorten = vi.fn().mockResolvedValue({ link: existingLink, existing: true });
		render(<NewUrlForm onShorten={onShorten} isLoading={false} />);
		fireEvent.change(screen.getByPlaceholderText("Paste a long URL here..."), {
			target: { value: "https://example.com" },
		});
		fireEvent.click(screen.getByRole("button", { name: /Shorten/ }));
		await screen.findByText("You already have a link for this URL");
		fireEvent.keyDown(document, { key: "Escape" });
		await waitFor(() =>
			expect(screen.queryByText("You already have a link for this URL")).toBeNull(),
		);
		expect(onShorten).toHaveBeenCalledTimes(1);
		expect(
			(screen.getByPlaceholderText("Paste a long URL here...") as HTMLInputElement).value,
		).toBe("https://example.com");
	});
});
