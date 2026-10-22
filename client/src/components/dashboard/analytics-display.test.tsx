// @vitest-environment jsdom
import { cleanup, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it } from "vitest";
import { DashboardChart } from "./dashboard-chart";
import { DashboardStats } from "./dashboard-stats";
import { TotalClicksCard } from "@/components/link/total-clicks-card";
import { formatAnalyticsDate } from "@/lib/date-utils";

afterEach(cleanup);

describe("analytics display", () => {
	it("distinguishes unavailable analytics from a genuine zero", () => {
		const { rerender } = render(
			<DashboardStats totalLinks={2} activeLinks={1} totalClicks={0} analyticsAvailable={false} />,
		);
		expect(screen.getByText("Analytics unavailable")).toBeTruthy();
		expect(screen.queryByText("0")).toBeNull();
		rerender(<DashboardStats totalLinks={2} activeLinks={1} totalClicks={0} analyticsAvailable />);
		expect(screen.getByText("0")).toBeTruthy();
		expect(screen.queryByText("Analytics unavailable")).toBeNull();
	});

	it("shows separate unavailable and no-click chart states", () => {
		const { rerender } = render(<DashboardChart data={[]} analyticsAvailable={false} />);
		expect(screen.getByText(/Analytics unavailable\. Click data/)).toBeTruthy();
		rerender(<DashboardChart data={[]} analyticsAvailable />);
		expect(screen.getByText("No clicks in the last 30 days.")).toBeTruthy();
		rerender(<DashboardChart data={[{ name: "Oct 21", clicks: 0 }]} analyticsAvailable />);
		expect(screen.getByText("No clicks in the last 30 days.")).toBeTruthy();
	});

	it("labels the link click total with its reporting period", () => {
		render(<TotalClicksCard clicks={0} />);
		expect(screen.getByText("Clicks (30d)")).toBeTruthy();
	});

	it("formats analytics dates as calendar dates instead of repeating weekdays", () => {
		expect(formatAnalyticsDate("2026-10-20T00:00:00Z")).toBe("Oct 20");
		expect(formatAnalyticsDate("2026-10-21T00:00:00Z")).toBe("Oct 21");
	});
});
