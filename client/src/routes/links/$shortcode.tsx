import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { useAuth } from "@clerk/clerk-react";
import { Navigate } from "@tanstack/react-router";
import { Button } from "@/components/ui/button";
import { useLinks } from "@/hooks/use-links";
import { processReferrersData } from "@/lib/referrers";
import { LinkHeader } from "@/components/link/link-header";
import { LinkActions } from "@/components/link/link-actions";
import { LinkDetailsCard } from "@/components/link/link-details-card";
import { PerformanceChartCard } from "@/components/link/performance-chart-card";
import { TotalClicksCard } from "@/components/link/total-clicks-card";
import { TopSourcesCard } from "@/components/link/top-sources-card";
import { LoadingState } from "@/components/shared/loading-state";

export const Route = createFileRoute("/links/$shortcode")({
	component: LinkDetailPage,
});

function LinkDetailPage() {
	const { shortcode } = Route.useParams();
	const navigate = useNavigate();
	const { isSignedIn, isLoaded } = useAuth();
	const { data: linksData, isLoading: isLoadingLinks } = useLinks();
	const urls = linksData?.urls ?? [];

	const url = urls.find((u) => u.shortCode === shortcode);

	// Process referrers data to merge unknown sources into "Other"
	// React Compiler automatically memoizes this computation
	const processedReferrers = !url
		? []
		: processReferrersData(url.analytics.referrers_data || []);

	if (!isLoaded) {
		return <LoadingState />;
	}

	if (!isSignedIn) {
		return <Navigate to='/login' />;
	}

	if (isLoadingLinks) {
		return (
			<main className='p-4 sm:p-6'>
				<div className='max-w-6xl mx-auto text-center py-20'>
					<div className='text-lg text-muted-foreground'>Loading link...</div>
				</div>
			</main>
		);
	}

	if (!url) {
		return (
			<main className='p-4 sm:p-6'>
				<div className='max-w-6xl mx-auto text-center py-20'>
					<h1 className='text-3xl font-bold mb-4'>Link not found</h1>
					<Button onClick={() => navigate({ to: "/links" })}>
						Back to links
					</Button>
				</div>
			</main>
		);
	}

	return (
		<main className='p-4 sm:p-6'>
			<div className='max-w-6xl mx-auto'>
				<div className='mb-6'>
					<div className='flex items-center justify-between gap-4'>
						<div className='flex-1'>
							<LinkHeader url={url} />
						</div>
						<div className='flex items-center'>
							<LinkActions url={url} />
						</div>
					</div>
				</div>

				<div className='grid grid-cols-1 lg:grid-cols-3 gap-6'>
					{/* Main Content */}
					<div className='lg:col-span-2 space-y-6'>
						<LinkDetailsCard url={url} />
						<PerformanceChartCard url={url} />
					</div>

					{/* Sidebar Stats */}
					<div className='grid grid-cols-2 lg:grid-cols-1 lg:grid-rows-[auto_1fr] gap-6'>
						<TotalClicksCard clicks={url.clicks} />
						<TopSourcesCard referrers={processedReferrers} />
					</div>
				</div>
			</div>
		</main>
	);
}
