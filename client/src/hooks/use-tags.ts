import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { useAuth } from "@clerk/clerk-react";
import { apiClient } from "@/lib/api-client";
import { linkKeys } from "./use-links";
import { toast } from "sonner";
import type { Tag } from "@/types/api";
import type { Url } from "@/types/url";

// Query keys
export const tagKeys = {
	all: ["tags"] as const,
	lists: () => [...tagKeys.all, "list"] as const,
	list: () => [...tagKeys.lists()] as const,
};

// Fetch all tags
export function useTags() {
	const { getToken } = useAuth();

	return useQuery({
		queryKey: tagKeys.list(),
		queryFn: async () => {
			const token = await getToken();
			const response = await apiClient.get<Tag[]>("/api/v1/tags", token);
			return response.data;
		},
	});
}

// Create tag mutation
export function useCreateTag() {
	const { getToken } = useAuth();
	const queryClient = useQueryClient();

	return useMutation({
		mutationFn: async (name: string) => {
			const token = await getToken();
			const response = await apiClient.post<Tag>(
				"/api/v1/tags",
				{ name },
				token
			);
			return response.data;
		},
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: tagKeys.list() });
			toast.success("Tag created successfully");
		},
		onError: (error: Error) => {
			toast.error(error.message || "Failed to create tag");
		},
	});
}

// Add tags to a link
export function useAddTagsToLink() {
	const { getToken } = useAuth();
	const queryClient = useQueryClient();

	return useMutation({
		mutationFn: async ({
			linkId,
			tagIds,
		}: {
			linkId: string;
			tagIds: string[];
		}) => {
			const token = await getToken();
			await apiClient.post(
				`/api/v1/links/${linkId}/tags`,
				{ tag_ids: tagIds },
				token
			);
		},
		onMutate: async ({ linkId, tagIds }) => {
			// Cancel any outgoing refetches
			await queryClient.cancelQueries({ queryKey: linkKeys.list() });
			await queryClient.cancelQueries({ queryKey: linkKeys.detail(linkId) });
			await queryClient.cancelQueries({ queryKey: tagKeys.list() });

			// Snapshot the previous values
			const previousLinks = queryClient.getQueryData<Url[]>(linkKeys.list());
			const previousLink = queryClient.getQueryData<Url>(linkKeys.detail(linkId));
			const availableTags = queryClient.getQueryData<Tag[]>(tagKeys.list()) || [];

			// Get the tag objects for the tagIds being added
			const tagsToAdd = availableTags.filter((tag) => tagIds.includes(tag.id));

			// Optimistically update the link in the list
			if (previousLinks) {
				queryClient.setQueryData<Url[]>(
					linkKeys.list(),
					previousLinks.map((link) => {
						if (link.id === linkId) {
							const existingTagIds = new Set(link.tags.map((t) => t.id));
							const newTags = tagsToAdd
								.filter((tag) => !existingTagIds.has(tag.id))
								.map((tag) => ({
									id: tag.id,
									name: tag.name,
								}));
							return {
								...link,
								tags: [...link.tags, ...newTags],
							};
						}
						return link;
					})
				);
			}

			// Optimistically update the single link detail
			if (previousLink) {
				const existingTagIds = new Set(previousLink.tags.map((t) => t.id));
				const newTags = tagsToAdd
					.filter((tag) => !existingTagIds.has(tag.id))
					.map((tag) => ({
						id: tag.id,
						name: tag.name,
					}));
				queryClient.setQueryData<Url>(linkKeys.detail(linkId), {
					...previousLink,
					tags: [...previousLink.tags, ...newTags],
				});
			}

			// Return context for rollback
			return { previousLinks, previousLink };
		},
		onError: (error: Error, variables, context) => {
			// Rollback on error
			if (context?.previousLinks) {
				queryClient.setQueryData(linkKeys.list(), context.previousLinks);
			}
			if (context?.previousLink) {
				queryClient.setQueryData(linkKeys.detail(variables.linkId), context.previousLink);
			}
			toast.error(error.message || "Failed to add tags");
		},
		onSuccess: (_, variables) => {
			// Invalidate to refetch and ensure consistency
			queryClient.invalidateQueries({ queryKey: linkKeys.list() });
			queryClient.invalidateQueries({
				queryKey: linkKeys.detail(variables.linkId),
			});
			toast.success("Tags added successfully");
		},
	});
}

// Remove tags from a link
export function useRemoveTagsFromLink() {
	const { getToken } = useAuth();
	const queryClient = useQueryClient();

	return useMutation({
		mutationFn: async ({
			linkId,
			tagIds,
		}: {
			linkId: string;
			tagIds: string[];
		}) => {
			const token = await getToken();
			await apiClient.post(
				`/api/v1/links/${linkId}/tags/remove`,
				{ tag_ids: tagIds },
				token
			);
		},
		onMutate: async ({ linkId, tagIds }) => {
			// Cancel any outgoing refetches
			await queryClient.cancelQueries({ queryKey: linkKeys.list() });
			await queryClient.cancelQueries({ queryKey: linkKeys.detail(linkId) });

			// Snapshot the previous values
			const previousLinks = queryClient.getQueryData<Url[]>(linkKeys.list());
			const previousLink = queryClient.getQueryData<Url>(linkKeys.detail(linkId));

			// Optimistically update the link in the list
			if (previousLinks) {
				queryClient.setQueryData<Url[]>(
					linkKeys.list(),
					previousLinks.map((link) =>
						link.id === linkId
							? {
									...link,
									tags: link.tags.filter((tag) => !tagIds.includes(tag.id)),
								}
							: link
					)
				);
			}

			// Optimistically update the single link detail
			if (previousLink) {
				queryClient.setQueryData<Url>(linkKeys.detail(linkId), {
					...previousLink,
					tags: previousLink.tags.filter((tag) => !tagIds.includes(tag.id)),
				});
			}

			// Return context for rollback
			return { previousLinks, previousLink };
		},
		onError: (error: Error, variables, context) => {
			// Rollback on error
			if (context?.previousLinks) {
				queryClient.setQueryData(linkKeys.list(), context.previousLinks);
			}
			if (context?.previousLink) {
				queryClient.setQueryData(linkKeys.detail(variables.linkId), context.previousLink);
			}
			toast.error(error.message || "Failed to remove tags");
		},
		onSuccess: (_, variables) => {
			// Invalidate to refetch and ensure consistency
			queryClient.invalidateQueries({ queryKey: linkKeys.list() });
			queryClient.invalidateQueries({
				queryKey: linkKeys.detail(variables.linkId),
			});
			toast.success("Tags removed successfully");
		},
	});
}
