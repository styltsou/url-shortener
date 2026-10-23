import { useQuery } from "@tanstack/react-query";
import { useAuth } from "@clerk/clerk-react";
import { apiClient } from "@/lib/api-client";
import {
	analyticsQueryString,
	DEFAULT_ANALYTICS_FILTER,
	type AnalyticsFilter,
} from "@/lib/analytics-filter";

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
	analytics_available: boolean;
	clicks_over_time: ClicksOverTime[];
	recent_links: RecentLink[];
}

export const dashboardKeys = {
	all: ["dashboard"] as const,
};

export function useDashboard(filter: AnalyticsFilter = DEFAULT_ANALYTICS_FILTER) {
	const { getToken, isLoaded, isSignedIn, userId } = useAuth();
	const authUserId = isLoaded && isSignedIn ? userId : null;

	return useQuery({
		queryKey: [...dashboardKeys.all, authUserId, filter.range, filter.interval],
		enabled: !!authUserId,
		queryFn: async () => {
			const token = await getToken();
			const response = await apiClient.get<{ data: DashboardData }>(
				`/api/v1/dashboard?${analyticsQueryString(filter)}`,
				token,
			);
			return response.data;
		},
	});
}
