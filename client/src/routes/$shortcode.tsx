import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { useAuth } from "@clerk/clerk-react";
import { Navigate } from "@tanstack/react-router";
import { useState, useMemo, useRef } from "react";
import { Button } from "@/components/ui/button";
import { useLinks, useUpdateLink } from "@/hooks/use-links";
import { processReferrersData } from "@/lib/referrers";
import { LinkHeader } from "@/components/link/link-header";
import { LinkActions } from "@/components/link/link-actions";
import { DestinationCard } from "@/components/link/destination-card";
import { PerformanceChartCard } from "@/components/link/performance-chart-card";
import { TotalClicksCard } from "@/components/link/total-clicks-card";
import { TopSourcesCard } from "@/components/link/top-sources-card";

export const Route = createFileRoute("/$shortcode")({
	component: LinkDetailPage,
});

function LinkDetailPage() {
	const { shortcode } = Route.useParams();
	const navigate = useNavigate();
	const { isSignedIn, isLoaded } = useAuth();
	const { data: urls = [], isLoading: isLoadingLinks } = useLinks();
	const updateLink = useUpdateLink();
	const [isEditing, setIsEditing] = useState(false);
	const destinationCardRef = useRef<{
		getFormData: () => { expirationDate?: Date };
	}>(null);

	const url = urls.find((u) => u.shortCode === shortcode);

	// Process referrers data to merge unknown sources into "Other"
	const processedReferrers = useMemo(() => {
		if (!url) return [];
		const referrersData = url.analytics.referrers_data || [];
		return processReferrersData(referrersData);
	}, [url]);

	if (!isLoaded) {
		return (
			<div className='flex min-h-screen items-center justify-center'>
				<div className='text-lg'>Loading...</div>
			</div>
		);
	}

	if (!isSignedIn) {
		return <Navigate to='/login' />;
	}

	if (isLoadingLinks) {
		return (
			<main className='py-12 px-4 sm:px-6'>
				<div className='max-w-6xl mx-auto text-center py-20'>
					<div className='text-lg text-muted-foreground'>Loading link...</div>
				</div>
			</main>
		);
	}

	if (!url) {
		return (
			<main className='py-12 px-4 sm:px-6'>
				<div className='max-w-6xl mx-auto text-center py-20'>
					<h1 className='text-3xl font-bold mb-4'>Link not found</h1>
					<Button onClick={() => navigate({ to: "/" })}>
						Back to dashboard
					</Button>
				</div>
			</main>
		);
	}

	const handleSave = async () => {
		if (!url) return;

		const formData = destinationCardRef.current?.getFormData();
		await updateLink.mutateAsync({
			id: url.id,
			data: {
				expires_at: formData?.expirationDate
					? formData.expirationDate.toISOString().split("T")[0]
					: null,
				// TODO: Add original_url update when backend supports it
			},
		});
		setIsEditing(false);
	};

	return (
		<main className='py-12 px-4 sm:px-6'>
			<div className='max-w-6xl mx-auto'>
				<div className='mb-8'>
					<LinkHeader url={url} />
					<div className='flex justify-end mt-4'>
						<LinkActions
							url={url}
							isEditing={isEditing}
							onEdit={() => setIsEditing(true)}
							onCancel={() => setIsEditing(false)}
							onSave={handleSave}
						/>
					</div>
				</div>

				<div className='grid grid-cols-1 lg:grid-cols-3 gap-6'>
					{/* Main Content */}
					<div className='lg:col-span-2 space-y-6'>
						<DestinationCard
							ref={destinationCardRef}
							url={url}
							isEditing={isEditing}
						/>

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
