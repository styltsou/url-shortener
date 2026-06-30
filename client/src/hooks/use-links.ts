import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { useAuth } from "@clerk/clerk-react";
import { apiClient } from "@/lib/api-client";
import { linkToUrl } from "@/types/api";
import type {
	Link,
	CreateLinkRequest,
	UpdateLinkRequest,
	SuccessResponse,
	PaginationMeta,
} from "@/types/api";
import type { Url } from "@/types/url";
import { toast } from "sonner";
import { DEFAULT_PAGE_SIZE } from "@/lib/constants";

// Query keys
export const linkKeys = {
	all: ["links"] as const,
	lists: () => [...linkKeys.all, "list"] as const,
	list: (filters?: { tagIds?: string[]; status?: string; page?: number }) => {
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
	const { getToken } = useAuth();
	const page = options?.page ?? 1;
	const limit = options?.limit ?? DEFAULT_PAGE_SIZE;

	return useQuery({
		queryKey: linkKeys.list({ ...options, page }),
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

// Fetch single link by ID
export function useLink(id: string) {
	const { getToken } = useAuth();

	return useQuery({
		queryKey: linkKeys.detail(id),
		queryFn: async () => {
			const token = await getToken();
			const response = await apiClient.get<{ data: Link }>(`/api/v1/links/${id}`, token);
			return linkToUrl(response.data);
		},
		enabled: !!id,
	});
}

// Create link mutation
export function useCreateLink() {
	const { getToken } = useAuth();
	const queryClient = useQueryClient();

	return useMutation({
		mutationFn: async (data: CreateLinkRequest) => {
			const token = await getToken();
			const response = await apiClient.post<{ data: Link }>("/api/v1/links", data, token);
			return linkToUrl(response.data);
		},
		onSuccess: () => {
			// Invalidate all list queries to refresh the links list
			queryClient.invalidateQueries({ queryKey: linkKeys.lists() });
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
		onSuccess: (_, variables) => {
			queryClient.invalidateQueries({ queryKey: linkKeys.lists() });
			queryClient.invalidateQueries({
				queryKey: linkKeys.detail(variables.id),
			});
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
			queryClient.invalidateQueries({ queryKey: linkKeys.list() });
			toast.success("Link deleted successfully");
		},
	});
}
