import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { useAuth } from "@clerk/clerk-react";
import { Navigate } from "@tanstack/react-router";
import { Button } from "@/components/ui/button";
import { useLink } from "@/hooks/use-links";
import { ApiError } from "@/lib/api-client";
import { useAnalytics } from "@/hooks/use-analytics";
import { processReferrersData } from "@/lib/referrers";
import { LinkHeader } from "@/components/link/link-header";
import { LinkActions } from "@/components/link/link-actions";
import { LinkDetailsCard } from "@/components/link/link-details-card";
import { LinkQRCode } from "@/components/link/link-qrcode";
import { PerformanceChartCard } from "@/components/link/performance-chart-card";
import { TotalClicksCard } from "@/components/link/total-clicks-card";
import { TopSourcesCard } from "@/components/link/top-sources-card";
import { LoadingState } from "@/components/shared/loading-state";
import { LinkPageSkeleton } from "@/components/link/link-page-skeleton";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { formatAnalyticsDate } from "@/lib/date-utils";

export const Route = createFileRoute("/links/$shortcode")({
	component: LinkDetailPage,
});

function LinkDetailPage() {
	const { shortcode } = Route.useParams();
	const navigate = useNavigate();
	const { isSignedIn, isLoaded } = useAuth();
	const { data: url, isLoading: isLoadingLinks, error } = useLink(shortcode);
	const {
		data: analyticsData,
		isLoading: isLoadingAnalytics,
		isError: isAnalyticsError,
	} = useAnalytics(shortcode);

	const clicksData = (analyticsData?.clicks_over_time ?? []).map((d) => ({
		name: formatAnalyticsDate(d.date),
		clicks: d.clicks,
	}));

	const referrersData = analyticsData?.top_referrers ?? [];

	const processedReferrers = processReferrersData(referrersData);

	if (!isLoaded) {
		return <LoadingState />;
	}

	if (!isSignedIn) {
		return <Navigate to="/login" />;
	}

	if (isLoadingLinks) {
		return (
			<main className="p-4 sm:p-6">
				<div className="max-w-6xl mx-auto">
					<LinkPageSkeleton />
				</div>
			</main>
		);
	}

	if (!url) {
		return (
			<main className="p-4 sm:p-6">
				<div className="max-w-6xl mx-auto text-center py-20">
					<h1 className="text-3xl font-bold mb-4">
						{error && (!(error instanceof ApiError) || error.status !== 404)
							? "Unable to load link"
							: "Link not found"}
					</h1>
					<Button onClick={() => navigate({ to: "/links" })}>Back to links</Button>
				</div>
			</main>
		);
	}

	return (
		<main className="p-4 sm:p-6">
			<div className="max-w-6xl mx-auto">
				<div className="mb-6">
					<div className="flex items-center justify-between gap-4">
						<div className="flex-1">
							<LinkHeader url={url} />
						</div>
						<div className="flex items-center">
							<LinkActions url={url} />
						</div>
					</div>
				</div>

				<div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
					<div className="lg:col-span-2 space-y-6">
						<LinkDetailsCard url={url} />
						{isAnalyticsError ? (
							<Card>
								<CardHeader>
									<CardTitle>Analytics unavailable</CardTitle>
								</CardHeader>
								<CardContent className="text-sm text-muted-foreground">
									Click data cannot be loaded right now. Please try again later.
								</CardContent>
							</Card>
						) : (
							<PerformanceChartCard clicksData={clicksData} isLoading={isLoadingAnalytics} />
						)}
					</div>

					<div className="grid grid-cols-2 lg:grid-cols-1 gap-6">
						{!isAnalyticsError && (
							<>
								<TotalClicksCard
									clicks={analyticsData?.total_clicks ?? 0}
									isLoading={isLoadingAnalytics}
								/>
								<TopSourcesCard referrers={processedReferrers} />
							</>
						)}
						<LinkQRCode shortcode={shortcode} />
					</div>
				</div>
			</div>
		</main>
	);
}
