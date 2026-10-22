import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { ClicksGraph } from "@/components/link/clicks-graph";
import { LoadingState } from "@/components/shared/loading-state";

interface DashboardChartProps {
	data: Array<{ name: string; clicks: number }>;
	isLoading?: boolean;
	analyticsAvailable: boolean;
}

export function DashboardChart({ data, isLoading, analyticsAvailable }: DashboardChartProps) {
	if (isLoading) {
		return (
			<Card>
				<CardHeader>
					<CardTitle className="text-sm font-semibold uppercase tracking-wider text-muted-foreground">
						Clicks Over Time (30d)
					</CardTitle>
				</CardHeader>
				<CardContent>
					<LoadingState />
				</CardContent>
			</Card>
		);
	}

	if (!analyticsAvailable) {
		return (
			<Card>
				<CardHeader>
					<CardTitle className="text-sm font-semibold uppercase tracking-wider text-muted-foreground">
						Clicks Over Time (30d)
					</CardTitle>
				</CardHeader>
				<CardContent>
					<p className="text-sm text-muted-foreground">
						Analytics unavailable. Click data cannot be loaded right now.
					</p>
				</CardContent>
			</Card>
		);
	}

	if (data.every((day) => day.clicks === 0)) {
		return (
			<Card>
				<CardHeader>
					<CardTitle className="text-sm font-semibold uppercase tracking-wider text-muted-foreground">
						Clicks Over Time (30d)
					</CardTitle>
				</CardHeader>
				<CardContent>
					<p className="text-sm text-muted-foreground">No clicks in the last 30 days.</p>
				</CardContent>
			</Card>
		);
	}

	return (
		<Card>
			<CardHeader>
				<CardTitle className="text-sm font-semibold uppercase tracking-wider text-muted-foreground">
					Clicks Over Time (30d)
				</CardTitle>
			</CardHeader>
			<CardContent>
				<ClicksGraph data={data} />
			</CardContent>
		</Card>
	);
}
