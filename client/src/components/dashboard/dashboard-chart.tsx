import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { ClicksGraph } from "@/components/link/clicks-graph";
import { LoadingState } from "@/components/shared/loading-state";

interface DashboardChartProps {
	data: Array<{ name: string; clicks: number }>;
	isLoading?: boolean;
}

export function DashboardChart({ data, isLoading }: DashboardChartProps) {
	if (isLoading) {
		return (
			<Card>
				<CardHeader>
					<CardTitle className="text-sm font-semibold uppercase tracking-wider text-muted-foreground">
						Clicks Over Time
					</CardTitle>
				</CardHeader>
				<CardContent>
					<LoadingState />
				</CardContent>
			</Card>
		);
	}

	if (data.length === 0) {
		return null;
	}

	return (
		<Card>
			<CardHeader>
				<CardTitle className="text-sm font-semibold uppercase tracking-wider text-muted-foreground">
					Clicks Over Time
				</CardTitle>
			</CardHeader>
			<CardContent>
				<ClicksGraph data={data} />
			</CardContent>
		</Card>
	);
}
