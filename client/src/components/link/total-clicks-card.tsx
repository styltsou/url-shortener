import { TrendingUp } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

interface TotalClicksCardProps {
	clicks: number;
}

export function TotalClicksCard({ clicks }: TotalClicksCardProps) {
	return (
		<Card>
			<CardHeader>
				<CardTitle className='text-sm font-semibold uppercase tracking-wider text-muted-foreground'>
					Total Clicks
				</CardTitle>
			</CardHeader>
			<CardContent>
				<div className='flex items-baseline justify-between gap-4'>
					<p className='text-3xl font-bold tracking-tight text-foreground'>
						{clicks.toLocaleString()}
					</p>
					<div className='flex items-center gap-1.5 text-sm text-muted-foreground whitespace-nowrap'>
						<TrendingUp className='w-4 h-4' />
						<span>+12.5% this week</span>
					</div>
				</div>
			</CardContent>
		</Card>
	);
}

