import { useQuery } from "@tanstack/react-query";
import { useAuth } from "@clerk/clerk-react";
import { apiClient } from "@/lib/api-client";

export interface ClicksOverTime {
	date: string;
	clicks: number;
}

export interface RecentLink {
	id: string;
	shortcode: string;
	original_url: string;
	is_active: boolean;
	expires_at: string | null;
	created_at: string;
	updated_at: string | null;
}

export interface DashboardData {
	total_links: number;
	active_links: number;
	total_clicks: number;
	clicks_over_time: ClicksOverTime[];
	recent_links: RecentLink[];
}

export const dashboardKeys = {
	all: ["dashboard"] as const,
};

export function useDashboard() {
	const { getToken } = useAuth();

	return useQuery({
		queryKey: dashboardKeys.all,
		queryFn: async () => {
			const token = await getToken();
			const response = await apiClient.get<{ data: DashboardData }>("/api/v1/dashboard", token);
			return response.data;
		},
	});
}
