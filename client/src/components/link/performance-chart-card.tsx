import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { ClicksGraph } from "./clicks-graph";
import { LoadingState } from "@/components/shared/loading-state";

interface PerformanceChartCardProps {
	clicksData: Array<{ name: string; clicks: number }>;
	isLoading?: boolean;
}

export function PerformanceChartCard({ clicksData, isLoading }: PerformanceChartCardProps) {
	if (isLoading) {
		return (
			<Card>
				<CardHeader>
					<CardTitle className='text-sm font-semibold uppercase tracking-wider text-muted-foreground'>
						Performance
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
				<CardTitle className='text-sm font-semibold uppercase tracking-wider text-muted-foreground'>
					Performance
				</CardTitle>
			</CardHeader>
			<CardContent>
				<ClicksGraph data={clicksData} />
			</CardContent>
		</Card>
	);
}

