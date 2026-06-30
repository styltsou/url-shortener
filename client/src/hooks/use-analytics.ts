import { useQuery } from "@tanstack/react-query";
import { useAuth } from "@clerk/clerk-react";
import { apiClient } from "@/lib/api-client";

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

export function useAnalytics(shortcode: string) {
	const { getToken } = useAuth();

	return useQuery({
		queryKey: analyticsKeys.detail(shortcode),
		queryFn: async () => {
			const token = await getToken();
			const response = await apiClient.get<{ data: AnalyticsData }>(
				`/api/v1/links/${shortcode}/analytics`,
				token,
			);
			return response.data;
		},
		enabled: !!shortcode,
	});
}
