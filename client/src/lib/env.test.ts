import { describe, expect, it } from "vitest";
import { buildShortURL } from "./env";

describe("short link URL", () => {
	it("uses the local HTTP redirect endpoint", () => {
		expect(buildShortURL("abc123", "http://localhost:8080"))
			.toBe("http://localhost:8080/abc123");
	});
	it("accepts an HTTPS origin or an existing domain setting", () => {
		expect(buildShortURL("abc123", "https://go.example/"))
			.toBe("https://go.example/abc123");
		expect(buildShortURL("abc123", "go.example"))
			.toBe("https://go.example/abc123");
	});
});
