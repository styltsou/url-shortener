import { createFileRoute } from "@tanstack/react-router";
import { useAuth } from "@clerk/clerk-react";
import { Navigate } from "@tanstack/react-router";
import { NewUrlForm } from "@/components/links/new-url-form";
import { UrlCard } from "@/components/links/url-card";
import { useLinks, useCreateLink } from "@/hooks/use-links";
import { LinksListSkeleton } from "@/components/links/links-list-skeleton";
import { LinksHeader } from "@/components/links/links-header";
import { LinksErrorState } from "@/components/links/links-error-state";
import { EmptyLinksState } from "@/components/links/empty-links-state";

export const Route = createFileRoute("/")({
	component: LinksPage,
});

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
		// eslint-disable-next-line @typescript-eslint/no-unused-vars
		_customCode?: string,
		// eslint-disable-next-line @typescript-eslint/no-unused-vars
		_expirationDate?: string
	) => {
		await createLink.mutateAsync({
			url: originalUrl,
			// TODO: Add customCode and expirationDate when backend supports it
		});
	};

	if (error) {
		return <LinksErrorState error={error} />;
	}

	return (
		<main className='py-12 px-4 sm:px-6'>
			<div className='max-w-4xl mx-auto'>
				<NewUrlForm
					onShorten={handleShorten}
					isLoading={createLink.isPending}
				/>

				<LinksHeader isLoading={isLoading} totalCount={urls.length} />

				{isLoading ? (
					<LinksListSkeleton />
				) : (
					<div className='space-y-2'>
						{urls.map((url) => (
							<UrlCard key={url.id} url={url} />
						))}
						{urls.length === 0 && <EmptyLinksState />}
					</div>
				)}
			</div>
		</main>
	);
}
