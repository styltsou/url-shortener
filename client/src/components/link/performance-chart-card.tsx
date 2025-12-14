import { useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "@/components/ui/select";
import { ClicksGraph } from "./clicks-graph";
import { generateMockAnalytics } from "@/lib/mock-data";
import type { Url } from "@/types/url";

interface PerformanceChartCardProps {
	url: Url;
}

export function PerformanceChartCard({ url }: PerformanceChartCardProps) {
	const [timeRange, setTimeRange] = useState<"7days" | "30days">("7days");

	// React Compiler automatically memoizes this computation
	const analyticsData = generateMockAnalytics(url, timeRange);

	return (
		<Card>
			<CardHeader>
				<div className='flex items-center justify-between'>
					<CardTitle className='text-sm font-semibold uppercase tracking-wider text-muted-foreground'>
						Performance
					</CardTitle>
					<Select value={timeRange} onValueChange={setTimeRange}>
						<SelectTrigger className='w-[140px]'>
							<SelectValue />
						</SelectTrigger>
						<SelectContent>
							<SelectItem value='7days'>Last 7 days</SelectItem>
							<SelectItem value='30days'>Last 30 days</SelectItem>
						</SelectContent>
					</Select>
				</div>
			</CardHeader>
			<CardContent>
				<ClicksGraph data={analyticsData.clicks_data} />
			</CardContent>
		</Card>
	);
}

