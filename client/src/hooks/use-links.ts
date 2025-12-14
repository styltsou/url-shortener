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

			// The server returns SuccessResponse<Link[]> directly: {data: Link[], pagination?: {...}}
			// apiClient.get returns the raw JSON response, which IS the SuccessResponse object
			// So response IS the SuccessResponse<Link[]> object, not wrapped
			const response = await apiClient.get<SuccessResponse<Link[]>>(url, token);

			// response IS the SuccessResponse<Link[]> object: {data: Link[], pagination?: {...}}
			// The apiClient.get type says it returns ApiSuccessResponse<T>, but it actually
			// returns the raw JSON which is the SuccessResponse object directly
			const successData = response as unknown as SuccessResponse<Link[]>;

			// Handle case where data might be missing or not an array
			if (!successData || !Array.isArray(successData.data)) {
				return {
					urls: [],
					pagination: successData?.pagination || {
						page: 1,
						limit: limit,
						total: 0,
						total_pages: 0,
					},
				};
			}

			const finalPagination = successData.pagination || {
				page: 1,
				limit: limit,
				total: 0,
				total_pages: 0,
			};

			return {
				urls: successData.data.map(linkToUrl),
				pagination: finalPagination,
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
			const response = await apiClient.get<Link>(`/api/v1/links/${id}`, token);
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
			const response = await apiClient.post<Link>("/api/v1/links", data, token);
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
		mutationFn: async ({
			id,
			data,
		}: {
			id: string;
			data: UpdateLinkRequest;
		}) => {
			const token = await getToken();
			const response = await apiClient.patch<Link>(
				`/api/v1/links/${id}`,
				data,
				token
			);
			return linkToUrl(response.data);
		},
		onMutate: async ({ id, data }) => {
			// Only do optimistic updates for is_active changes
			if (data.is_active === undefined) {
				return;
			}

			// Cancel any outgoing refetches (so they don't overwrite our optimistic update)
			await queryClient.cancelQueries({ queryKey: linkKeys.lists() });
			await queryClient.cancelQueries({ queryKey: linkKeys.detail(id) });

			// Get all list queries from cache and snapshot them
			const queryCache = queryClient.getQueryCache();
			const allListQueries = queryCache.findAll({ queryKey: linkKeys.lists() });
			const previousQueries = new Map<string, UseLinksResult>();

			// Snapshot all list queries
			allListQueries.forEach((query) => {
				const queryData = query.state.data as UseLinksResult | undefined;
				if (queryData) {
					previousQueries.set(JSON.stringify(query.queryKey), { ...queryData });
				}
			});

			// Snapshot the detail query
			const previousLink = queryClient.getQueryData<Url>(linkKeys.detail(id));

			// Optimistically update all list queries
			allListQueries.forEach((query) => {
				const queryData = query.state.data as UseLinksResult | undefined;
				if (queryData && queryData.urls) {
					const updatedUrls = queryData.urls.map((link) =>
						link.id === id ? { ...link, isActive: data.is_active } : link
					);
					queryClient.setQueryData<UseLinksResult>(query.queryKey, {
						...queryData,
						urls: updatedUrls,
					});
				}
			});

			// Optimistically update the detail
			if (previousLink) {
				queryClient.setQueryData<Url>(linkKeys.detail(id), {
					...previousLink,
					isActive: data.is_active,
				});
			}

			// Return a context object with the snapshotted values
			return { previousQueries, previousLink };
		},
		onError: (error: Error, variables, context) => {
			// If the mutation fails, use the context returned from onMutate to roll back
			if (context?.previousQueries) {
				context.previousQueries.forEach((previousData, queryKeyStr) => {
					const queryKey = JSON.parse(queryKeyStr);
					queryClient.setQueryData<UseLinksResult>(queryKey, previousData);
				});
			}
			if (context?.previousLink) {
				queryClient.setQueryData(
					linkKeys.detail(variables.id),
					context.previousLink
				);
			}
			toast.error(error.message || "Failed to update link");
		},
		onSuccess: (updatedLink, variables) => {
			// Update all list queries with the server response
			if (updatedLink) {
				const queryCache = queryClient.getQueryCache();
				const allListQueries = queryCache.findAll({
					queryKey: linkKeys.lists(),
				});

				allListQueries.forEach((query) => {
					const queryData = query.state.data as UseLinksResult | undefined;
					if (queryData && queryData.urls) {
						const updatedUrls = queryData.urls.map((link) =>
							link.id === variables.id ? updatedLink : link
						);
						queryClient.setQueryData<UseLinksResult>(query.queryKey, {
							...queryData,
							urls: updatedUrls,
						});
					}
				});

				// Also update the detail query if it exists
				queryClient.setQueryData<Url>(
					linkKeys.detail(variables.id),
					updatedLink
				);
			}

			// Use refetchQueries instead of invalidateQueries to keep the optimistic update visible
			// during refetch. This prevents the cache from being cleared during refetch.
			queryClient.refetchQueries({ queryKey: linkKeys.lists() });
			queryClient.refetchQueries({
				queryKey: linkKeys.detail(variables.id),
			});
			toast.success("Link updated successfully");
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
