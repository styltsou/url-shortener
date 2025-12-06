import { createFileRoute } from "@tanstack/react-router";
import { useAuth } from "@clerk/clerk-react";
import { Navigate } from "@tanstack/react-router";
import { Link as LinkIcon } from "lucide-react";
import { NewUrlForm } from "@/components/NewUrlForm";
import { UrlCard } from "@/components/UrlCard";
import { useLinks, useCreateLink } from "@/hooks/use-links";
import { Card } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";

export const Route = createFileRoute("/")({
	component: LinksPage,
});

function LinksListSkeleton() {
	return (
		<div className='space-y-2'>
			{[1, 2, 3].map((i) => (
				<Card key={i} className='py-1 px-4'>
					<div className='flex flex-col md:flex-row md:items-center justify-between gap-2'>
						<div className='flex-1 space-y-1'>
							<Skeleton className='h-7 w-48 sm:w-64' />
							<Skeleton className='h-4 w-full sm:w-96' />
						</div>
						<div className='flex items-center gap-6 mt-2 md:mt-0'>
							<Skeleton className='h-5 w-16' />
							<Skeleton className='h-5 w-24 hidden sm:block' />
							<div className='flex gap-2 pl-4 border-l border-border'>
								<Skeleton className='h-8 w-8 rounded-md' />
								<Skeleton className='h-8 w-8 rounded-md' />
							</div>
						</div>
					</div>
				</Card>
			))}
		</div>
	);
}

function LinksPage() {
	const { isSignedIn, isLoaded } = useAuth();
	const { data: urls = [], isLoading, error } = useLinks();
	const createLink = useCreateLink();

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

	const handleShorten = async (
		originalUrl: string,
		customCode?: string,
		expirationDate?: string
	) => {
		await createLink.mutateAsync({
			url: originalUrl,
			// TODO: Add customCode and expirationDate when backend supports it
		});
	};

	if (error) {
		return (
			<main className='py-12 px-4 sm:px-6'>
				<div className='max-w-4xl mx-auto text-center py-20'>
					<h2 className='text-2xl font-bold mb-4 text-destructive'>
						Error loading links
					</h2>
					<p className='text-muted-foreground'>{error.message}</p>
				</div>
			</main>
		);
	}

	return (
		<main className='py-12 px-4 sm:px-6'>
			<div className='max-w-4xl mx-auto'>
				<NewUrlForm
					onShorten={handleShorten}
					isLoading={createLink.isPending}
				/>

				<div className='flex items-end justify-between mb-4 pb-2 border-b border-border'>
					<h2 className='text-lg font-bold text-foreground tracking-tight'>
						Links
					</h2>
					<span className='text-xs font-medium text-muted-foreground uppercase tracking-wider'>
						{isLoading ? "Loading..." : `${urls.length} Total`}
					</span>
				</div>

				{isLoading ? (
					<LinksListSkeleton />
				) : (
					<div className='space-y-2'>
						{urls.map((url) => (
							<UrlCard key={url.id} url={url} />
						))}
						{urls.length === 0 && (
							<div className='text-center py-20 bg-card rounded-3xl border border-border'>
								<div className='w-16 h-16 bg-muted rounded-full flex items-center justify-center mx-auto mb-4'>
									<LinkIcon className='w-8 h-8 text-muted-foreground' />
								</div>
								<h3 className='text-foreground font-medium mb-1'>
									No links yet
								</h3>
								<p className='text-muted-foreground'>
									Create your first shortened link above.
								</p>
							</div>
						)}
					</div>
				)}
			</div>
		</main>
	);
}
