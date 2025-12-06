import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { useAuth } from "@clerk/clerk-react";
import { apiClient } from "@/lib/api-client";
import { linkToUrl } from "@/types/api";
import type { Link, CreateLinkRequest, UpdateLinkRequest } from "@/types/api";
import type { Url } from "@/types/url";
import { toast } from "sonner";
import { INITIAL_MOCK_URLS } from "@/lib/mock-data";

// Query keys
export const linkKeys = {
	all: ["links"] as const,
	lists: () => [...linkKeys.all, "list"] as const,
	list: () => [...linkKeys.lists()] as const,
	details: () => [...linkKeys.all, "detail"] as const,
	detail: (id: string) => [...linkKeys.details(), id] as const,
};

// Fetch all links
export function useLinks() {
	const { getToken } = useAuth();

	return useQuery({
		queryKey: linkKeys.list(),
		queryFn: async () => {
			try {
				const token = await getToken();
				const response = await apiClient.get<Link[]>("/api/v1/links", token);
				const apiUrls = response.data.map(linkToUrl);

				// TODO: Remove this when backend provides all links with isActive field
				// For now, merge with mock data to show inactive link example
				const mockInactiveLink = INITIAL_MOCK_URLS.find((u) => u.id === "4");
				if (
					mockInactiveLink &&
					!apiUrls.find((u) => u.id === mockInactiveLink.id)
				) {
					return [...apiUrls, mockInactiveLink];
				}

				return apiUrls;
			} catch (error) {
				// If API fails, return mock data for development
				// TODO: Remove this fallback when backend is fully ready
				console.warn("API call failed, using mock data:", error);
				return INITIAL_MOCK_URLS;
			}
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
			queryClient.invalidateQueries({ queryKey: linkKeys.list() });
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
		onSuccess: (_, variables) => {
			queryClient.invalidateQueries({ queryKey: linkKeys.list() });
			queryClient.invalidateQueries({
				queryKey: linkKeys.detail(variables.id),
			});
			toast.success("Link updated successfully");
		},
		onError: (error: Error) => {
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
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: linkKeys.list() });
			toast.success("Link deleted successfully");
		},
		onError: (error: Error) => {
			toast.error(error.message || "Failed to delete link");
		},
	});
}
