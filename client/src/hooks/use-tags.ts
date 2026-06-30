import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { useAuth } from "@clerk/clerk-react";
import { apiClient } from "@/lib/api-client";
import { linkKeys } from "./use-links";
import { toast } from "sonner";
import type { Tag } from "@/types/api";

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
			const response = await apiClient.get<{ data: Tag[] }>("/api/v1/tags", token);
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
			const response = await apiClient.post<{ data: Tag }>(
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
		onSuccess: (_, variables) => {
			queryClient.invalidateQueries({ queryKey: linkKeys.lists() });
			queryClient.invalidateQueries({
				queryKey: linkKeys.detail(variables.linkId),
			});
			toast.success("Tags added successfully");
		},
		onError: (error) => {
			toast.error(error.message || "Failed to add tags");
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
		onSuccess: (_, variables) => {
			queryClient.invalidateQueries({ queryKey: linkKeys.lists() });
			queryClient.invalidateQueries({
				queryKey: linkKeys.detail(variables.linkId),
			});
			toast.success("Tags removed successfully");
		},
		onError: (error) => {
			toast.error(error.message || "Failed to remove tags");
		},
	});
}
