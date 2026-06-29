import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

interface TotalClicksCardProps {
	clicks: number;
	isLoading?: boolean;
}

export function TotalClicksCard({ clicks, isLoading }: TotalClicksCardProps) {
	if (isLoading) {
		return (
			<Card>
				<CardHeader>
					<CardTitle className='text-sm font-semibold uppercase tracking-wider text-muted-foreground'>
						Total Clicks
					</CardTitle>
				</CardHeader>
				<CardContent>
					<div className='flex items-baseline justify-between gap-4'>
						<p className='text-3xl font-bold tracking-tight text-foreground'>-</p>
					</div>
				</CardContent>
			</Card>
		);
	}

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
				</div>
			</CardContent>
		</Card>
	);
}

