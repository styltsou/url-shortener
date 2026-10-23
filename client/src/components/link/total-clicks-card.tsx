import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

interface TotalClicksCardProps {
	clicks: number;
	isLoading?: boolean;
	rangeLabel?: string;
}

export function TotalClicksCard({
	clicks,
	isLoading,
	rangeLabel = "30 days",
}: TotalClicksCardProps) {
	if (isLoading) {
		return (
			<Card>
				<CardHeader>
					<CardTitle className="text-sm font-semibold uppercase tracking-wider text-muted-foreground">
						Clicks (last {rangeLabel})
					</CardTitle>
				</CardHeader>
				<CardContent>
					<div className="flex items-baseline justify-between gap-4">
						<p className="text-3xl font-bold tracking-tight text-foreground">-</p>
					</div>
				</CardContent>
			</Card>
		);
	}

	return (
		<Card>
			<CardHeader>
				<CardTitle className="text-sm font-semibold uppercase tracking-wider text-muted-foreground">
					Clicks (last {rangeLabel})
				</CardTitle>
			</CardHeader>
			<CardContent>
				<div className="flex items-baseline justify-between gap-4">
					<p className="text-3xl font-bold tracking-tight text-foreground">
						{clicks.toLocaleString()}
					</p>
				</div>
			</CardContent>
		</Card>
	);
}
