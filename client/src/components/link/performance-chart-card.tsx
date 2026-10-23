import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { ClicksGraph } from "./clicks-graph";
import { LoadingState } from "@/components/shared/loading-state";

interface PerformanceChartCardProps {
	clicksData: Array<{ name: string; clicks: number }>;
	isLoading?: boolean;
	rangeLabel?: string;
}

export function PerformanceChartCard({
	clicksData,
	isLoading,
	rangeLabel = "30 days",
}: PerformanceChartCardProps) {
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

	return (
		<Card>
			<CardHeader>
				<CardTitle className="text-sm font-semibold uppercase tracking-wider text-muted-foreground">
					Clicks Over Time (last {rangeLabel})
				</CardTitle>
			</CardHeader>
			<CardContent>
				{clicksData.every((day) => day.clicks === 0) ? (
					<p className="text-sm text-muted-foreground">No clicks in the last {rangeLabel}.</p>
				) : (
					<ClicksGraph data={clicksData} />
				)}
			</CardContent>
		</Card>
	);
}
