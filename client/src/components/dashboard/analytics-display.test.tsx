// @vitest-environment jsdom
import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { DashboardChart } from "./dashboard-chart";
import { DashboardStats } from "./dashboard-stats";
import { TotalClicksCard } from "@/components/link/total-clicks-card";
import { AnalyticsFilterControls } from "@/components/shared/analytics-filter-controls";
import { formatAnalyticsBucketDate, formatAnalyticsDate } from "@/lib/date-utils";
import { analyticsPeriodLabel, analyticsQueryString } from "@/lib/analytics-filter";

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
		expect(screen.getByText("Clicks (last 30 days)")).toBeTruthy();
	});

	it("labels a selected range and serializes both analytics filters", () => {
		render(<TotalClicksCard clicks={12} periodLabel="last 12 months" />);
		expect(screen.getByText("Clicks (last 12 months)")).toBeTruthy();
		expect(analyticsQueryString({ range: "12m", interval: "month" })).toBe(
			"range=12m&interval=month",
		);
	});

	it("serializes and labels a custom period, including one day", () => {
		const period = { range: "custom", interval: "day", from: "2026-10-20", to: "2026-10-20" } as const;
		expect(analyticsQueryString(period)).toBe(
			"range=custom&interval=day&from=2026-10-20&to=2026-10-20",
		);
		expect(analyticsPeriodLabel(period)).toBe("Oct 20, 2026");
		expect(analyticsPeriodLabel({ ...period, to: "2026-10-23" })).toBe(
			"Oct 20, 2026 – Oct 23, 2026",
		);
		render(<TotalClicksCard clicks={5} periodLabel={analyticsPeriodLabel(period)} />);
		expect(screen.getByText("Clicks (Oct 20, 2026)")).toBeTruthy();
	});

	it("applies valid custom dates and blocks reversed dates", () => {
		const onChange = vi.fn();
		render(
			<AnalyticsFilterControls
				filter={{ range: "custom", interval: "day", from: "2026-10-20", to: "2026-10-20" }}
				onChange={onChange}
			/>,
		);
		fireEvent.change(screen.getByLabelText("To"), { target: { value: "2026-10-23" } });
		fireEvent.click(screen.getByRole("button", { name: "Apply" }));
		expect(onChange).toHaveBeenCalledWith({
			range: "custom",
			interval: "day",
			from: "2026-10-20",
			to: "2026-10-23",
		});
		fireEvent.change(screen.getByLabelText("From"), { target: { value: "2026-10-24" } });
		expect((screen.getByRole("button", { name: "Apply" }) as HTMLButtonElement).disabled).toBe(true);
	});

	it("formats analytics dates as calendar dates instead of repeating weekdays", () => {
		expect(formatAnalyticsDate("2026-10-20T00:00:00Z")).toBe("Oct 20");
		expect(formatAnalyticsDate("2026-10-21T00:00:00Z")).toBe("Oct 21");
		expect(formatAnalyticsBucketDate("2026-10-01T00:00:00Z", "month")).toBe("Oct 2026");
	});
});
