import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { ClicksGraph } from "@/components/link/clicks-graph";
import { LoadingState } from "@/components/shared/loading-state";

interface DashboardChartProps {
	data: Array<{ name: string; clicks: number }>;
	isLoading?: boolean;
	analyticsAvailable: boolean;
	rangeLabel?: string;
}

export function DashboardChart({
	data,
	isLoading,
	analyticsAvailable,
	rangeLabel = "30 days",
}: DashboardChartProps) {
	if (isLoading) {
		return (
			<Card>
				<CardHeader>
					<CardTitle className="text-sm font-semibold uppercase tracking-wider text-muted-foreground">
						Clicks Over Time (last {rangeLabel})
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
						Clicks Over Time (last {rangeLabel})
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
						Clicks Over Time (last {rangeLabel})
					</CardTitle>
				</CardHeader>
				<CardContent>
					<p className="text-sm text-muted-foreground">No clicks in the last {rangeLabel}.</p>
				</CardContent>
			</Card>
		);
	}

	return (
		<Card>
			<CardHeader>
				<CardTitle className="text-sm font-semibold uppercase tracking-wider text-muted-foreground">
					Clicks Over Time (last {rangeLabel})
				</CardTitle>
			</CardHeader>
			<CardContent>
				<ClicksGraph data={data} />
			</CardContent>
		</Card>
	);
}
