import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { useAuth } from "@clerk/clerk-react";
import { apiClient } from "@/lib/api-client";
import { linkToUrl } from "@/types/api";
import type {
	Link,
	CreateLinkRequest,
	CreateLinkResponse,
	UpdateLinkRequest,
	SuccessResponse,
	PaginationMeta,
} from "@/types/api";
import type { Url } from "@/types/url";
import { toast } from "sonner";
import { DEFAULT_PAGE_SIZE } from "@/lib/constants";
import { dashboardKeys } from "./use-dashboard";
import { analyticsKeys } from "./use-analytics";

// Query keys
export const linkKeys = {
	all: ["links"] as const,
	lists: () => [...linkKeys.all, "list"] as const,
	list: (filters?: { tagIds?: string[]; status?: string; page?: number; limit?: number }) => {
		const key = [...linkKeys.lists()] as const;
		if (filters) {
			return [...key, filters] as const;
		}
		return key;
	},
	details: () => [...linkKeys.all, "detail"] as const,
	detail: (id: string) => [...linkKeys.details(), id] as const,
};

interface UseLinksOptions {
	tagIds?: string[];
	status?: "all" | "active" | "inactive";
	page?: number;
	limit?: number;
}

interface UseLinksResult {
	urls: Url[];
	pagination: PaginationMeta;
}

// Fetch all links
export function useLinks(options?: UseLinksOptions) {
	const { getToken, isLoaded, isSignedIn, userId } = useAuth();
	const authUserId = isLoaded && isSignedIn ? userId : null;
	const page = options?.page ?? 1;
	const limit = options?.limit ?? DEFAULT_PAGE_SIZE;

	return useQuery({
		queryKey: [...linkKeys.list({ ...options, page, limit }), authUserId],
		enabled: !!authUserId,
		queryFn: async (): Promise<UseLinksResult> => {
			const token = await getToken();

			// Build query parameters
			const params = new URLSearchParams();
			if (options?.tagIds && options.tagIds.length > 0) {
				params.append("tags", options.tagIds.join(","));
			}
			if (options?.status && options.status !== "all") {
				params.append("status", options.status);
			}
			params.append("page", page.toString());
			params.append("limit", limit.toString());

			const queryString = params.toString();
			const url = `/api/v1/links?${queryString}`;

			const response = await apiClient.get<SuccessResponse<Link[]>>(url, token);

			return {
				urls: (response.data ?? []).map(linkToUrl),
				pagination: response.pagination ?? {
					page,
					limit,
					total: 0,
					total_pages: 0,
				},
			};
		},
	});
}

// The GET endpoint identifies links by shortcode; mutations use the UUID.
export function useLink(shortcode: string) {
	const { getToken, isLoaded, isSignedIn, userId } = useAuth();
	const authUserId = isLoaded && isSignedIn ? userId : null;

	return useQuery({
		queryKey: [...linkKeys.detail(shortcode), authUserId],
		queryFn: async () => {
			const token = await getToken();
			const response = await apiClient.get<{ data: Link }>(
				`/api/v1/links/${encodeURIComponent(shortcode)}`,
				token,
			);
			return linkToUrl(response.data);
		},
		enabled: !!authUserId && !!shortcode,
	});
}

// Advisory availability check for the custom alias field
export function useShortcodeAvailability(shortcode: string) {
	const { getToken, isLoaded, isSignedIn, userId } = useAuth();
	const authUserId = isLoaded && isSignedIn ? userId : null;

	return useQuery({
		queryKey: [...linkKeys.all, "availability", authUserId, shortcode],
		enabled: !!authUserId && !!shortcode,
		retry: false,
		queryFn: async () => {
			const token = await getToken();
			const response = await apiClient.get<{ data: { available: boolean } }>(
				`/api/v1/shortcodes/availability?shortcode=${encodeURIComponent(shortcode)}`,
				token,
			);
			return response.data.available;
		},
	});
}

// Create link mutation
export function useCreateLink() {
	const { getToken } = useAuth();
	const queryClient = useQueryClient();

	return useMutation({
		mutationFn: async (data: CreateLinkRequest) => {
			const token = await getToken();
			const response = await apiClient.post<CreateLinkResponse>("/api/v1/links", data, token);
			return { link: linkToUrl(response.data), existing: response.existing };
		},
		onSuccess: (result) => {
			if (result.existing) return;
			// Invalidate all list queries to refresh the links list
			queryClient.invalidateQueries({ queryKey: linkKeys.lists() });
			queryClient.invalidateQueries({ queryKey: dashboardKeys.all });
			toast.success("Link created successfully");
		},
		onError: (error: Error) => {
			toast.error(error.message || "Failed to create link");
		},
	});
}

// Update link mutation
export function useUpdateLink() {
	const { getToken } = useAuth();
	const queryClient = useQueryClient();

	return useMutation({
		mutationFn: async ({ id, data }: { id: string; data: UpdateLinkRequest }) => {
			const token = await getToken();
			const response = await apiClient.patch<{ data: Link }>(`/api/v1/links/${id}`, data, token);
			return linkToUrl(response.data);
		},
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: linkKeys.lists() });
			// A changed shortcode also leaves the old detail key behind.
			queryClient.invalidateQueries({ queryKey: linkKeys.details() });
			queryClient.invalidateQueries({ queryKey: analyticsKeys.all });
			queryClient.invalidateQueries({ queryKey: dashboardKeys.all });
			toast.success("Link updated successfully");
		},
		onError: (error) => {
			toast.error(error.message || "Failed to update link");
		},
	});
}

// Delete link mutation
export function useDeleteLink() {
	const { getToken } = useAuth();
	const queryClient = useQueryClient();

	return useMutation({
		mutationFn: async (id: string) => {
			const token = await getToken();
			await apiClient.delete(`/api/v1/links/${id}`, token);
		},
		onError: (error: Error) => {
			toast.error(error.message || "Failed to delete link");
		},
		onSuccess: () => {
			// Invalidate and refetch to get the updated list without the deleted link
			queryClient.invalidateQueries({ queryKey: linkKeys.lists() });
			queryClient.invalidateQueries({ queryKey: linkKeys.details() });
			queryClient.invalidateQueries({ queryKey: analyticsKeys.all });
			queryClient.invalidateQueries({ queryKey: dashboardKeys.all });
			toast.success("Link deleted successfully");
		},
	});
}
