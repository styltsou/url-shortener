import { createFileRoute } from "@tanstack/react-router";
import { useState } from "react";
import { useAuth } from "@clerk/clerk-react";
import { Navigate } from "@tanstack/react-router";
import { NewUrlForm } from "@/components/links/new-url-form";
import { UrlCard } from "@/components/links/url-card";
import { TagFilter } from "@/components/links/tag-filter";
import {
	StatusFilter,
	type StatusFilter as StatusFilterType,
} from "@/components/links/status-filter";
import { useLinks, useCreateLink } from "@/hooks/use-links";
import { useTags, useAddTagsToLink } from "@/hooks/use-tags";
import { LinksPageSkeleton } from "@/components/links/links-page-skeleton";
import { LinksHeader } from "@/components/links/links-header";
import { LinksErrorState } from "@/components/links/links-error-state";
import { EmptyLinksState } from "@/components/links/empty-links-state";
import { LinksPagination } from "@/components/links/links-pagination";
import { PageSizeSelector } from "@/components/links/page-size-selector";
import { LoadingState } from "@/components/shared/loading-state";
import { DEFAULT_PAGE_SIZE } from "@/lib/constants";

export const Route = createFileRoute("/links/")({
	component: LinksPage,
});

function LinksPage() {
	const { isSignedIn, isLoaded } = useAuth();
	const [selectedTagIds, setSelectedTagIds] = useState<string[]>([]);
	const [statusFilter, setStatusFilter] = useState<StatusFilterType>("all");
	const [page, setPage] = useState(1);
	const [limit, setLimit] = useState(DEFAULT_PAGE_SIZE);

	const { data, isLoading, error } = useLinks({
		tagIds: selectedTagIds.length > 0 ? selectedTagIds : undefined,
		status: statusFilter,
		page,
		limit,
	});

	const urls = data?.urls ?? [];
	const pagination = data?.pagination;
	const { data: availableTags = [], isLoading: isLoadingTags } = useTags();
	const createLink = useCreateLink();
	const addTagsToLink = useAddTagsToLink();

	// Reset to page 1 when filters or limit change
	const handleTagFilterChange = (tagIds: string[]) => {
		setSelectedTagIds(tagIds);
		setPage(1);
	};

	const handleStatusFilterChange = (status: StatusFilterType) => {
		setStatusFilter(status);
		setPage(1);
	};

	const handleLimitChange = (newLimit: number) => {
		setLimit(newLimit);
		setPage(1); // Reset to first page when changing page size
	};

	if (!isLoaded) {
		return <LoadingState />;
	}

	if (!isSignedIn) {
		return <Navigate to='/login' />;
	}

	const handleShorten = async (
		originalUrl: string,
		customCode?: string,
		expirationDate?: string,
		tagIds?: string[]
	) => {
		// Create the link first
		const createdLink = await createLink.mutateAsync({
			url: originalUrl,
			...(customCode && { shortcode: customCode }),
			...(expirationDate && { expires_at: expirationDate }),
		});

		// Then add tags if any were selected
		if (tagIds && tagIds.length > 0) {
			await addTagsToLink.mutateAsync({
				linkId: createdLink.id,
				tagIds: tagIds,
			});
		}
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

				{isLoading || isLoadingTags ? (
					<LinksPageSkeleton />
				) : (
					<>
						<div className='mb-4'>
							<LinksHeader
								isLoading={isLoading}
								totalCount={pagination?.total ?? 0}
								hasActiveFilters={
									selectedTagIds.length > 0 || statusFilter !== "all"
								}
							/>
							<div className='mt-3 flex items-center justify-between gap-4'>
								{availableTags.length > 0 && (
									<TagFilter
										availableTags={availableTags}
										selectedTagIds={selectedTagIds}
										onSelectionChange={handleTagFilterChange}
									/>
								)}
								<div className='ml-auto'>
									<StatusFilter
										value={statusFilter}
										onValueChange={handleStatusFilterChange}
									/>
								</div>
							</div>
						</div>

						<div className='space-y-2'>
							{urls.map((url) => (
								<UrlCard key={url.id} url={url} />
							))}
							{urls.length === 0 &&
								(selectedTagIds.length > 0 || statusFilter !== "all") && (
									<div className='text-center py-12 text-muted-foreground'>
										<p className='text-sm'>
											No links match the selected filters.
										</p>
									</div>
								)}
							{urls.length === 0 &&
								selectedTagIds.length === 0 &&
								statusFilter === "all" && <EmptyLinksState />}
						</div>
						{pagination && (
							<div className='mt-6 flex items-center justify-between'>
								{pagination.total_pages > 1 ? (
									<LinksPagination
										currentPage={pagination.page}
										totalPages={pagination.total_pages}
										onPageChange={setPage}
									/>
								) : (
									<div /> // Empty div to maintain flex layout
								)}
								<PageSizeSelector
									value={limit}
									onValueChange={handleLimitChange}
								/>
							</div>
						)}
					</>
				)}
			</div>
		</main>
	);
}
