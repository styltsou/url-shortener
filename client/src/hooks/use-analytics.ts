import { useQuery } from "@tanstack/react-query";
import { useAuth } from "@clerk/clerk-react";
import { apiClient } from "@/lib/api-client";
import {
	analyticsQueryString,
	DEFAULT_ANALYTICS_FILTER,
	type AnalyticsFilter,
} from "@/lib/analytics-filter";

export interface AnalyticsData {
	total_clicks: number;
	clicks_over_time: Array<{ date: string; clicks: number }>;
	top_referrers: Array<{ referrer: string; clicks: number }>;
	top_user_agents: Array<{ user_agent: string; clicks: number }>;
}

export const analyticsKeys = {
	all: ["analytics"] as const,
	detail: (shortcode: string) => [...analyticsKeys.all, shortcode] as const,
};

export function useAnalytics(shortcode: string, filter: AnalyticsFilter = DEFAULT_ANALYTICS_FILTER) {
	const { getToken, isLoaded, isSignedIn, userId } = useAuth();
	const authUserId = isLoaded && isSignedIn ? userId : null;

	return useQuery({
		queryKey: [...analyticsKeys.detail(shortcode), authUserId, filter.range, filter.interval],
		queryFn: async () => {
			const token = await getToken();
			const response = await apiClient.get<{ data: AnalyticsData }>(
				`/api/v1/links/${encodeURIComponent(shortcode)}/analytics?${analyticsQueryString(filter)}`,
				token,
			);
			return response.data;
		},
		enabled: !!authUserId && !!shortcode,
	});
}
