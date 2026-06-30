import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Link2, Activity, MousePointerClick } from "lucide-react";

interface DashboardStatsProps {
	totalLinks: number;
	activeLinks: number;
	totalClicks: number;
}

export function DashboardStats({ totalLinks, activeLinks, totalClicks }: DashboardStatsProps) {
	return (
		<div className="grid gap-4 sm:grid-cols-3">
			<Card>
				<CardHeader className="flex flex-row items-center justify-between pb-2">
					<CardTitle className="text-sm font-semibold uppercase tracking-wider text-muted-foreground">
						Total Links
					</CardTitle>
					<Link2 className="h-4 w-4 text-muted-foreground" />
				</CardHeader>
				<CardContent>
					<p className="text-3xl font-bold tracking-tight text-foreground">
						{totalLinks.toLocaleString()}
					</p>
				</CardContent>
			</Card>

			<Card>
				<CardHeader className="flex flex-row items-center justify-between pb-2">
					<CardTitle className="text-sm font-semibold uppercase tracking-wider text-muted-foreground">
						Active Links
					</CardTitle>
					<Activity className="h-4 w-4 text-muted-foreground" />
				</CardHeader>
				<CardContent>
					<p className="text-3xl font-bold tracking-tight text-foreground">
						{activeLinks.toLocaleString()}
					</p>
				</CardContent>
			</Card>

			<Card>
				<CardHeader className="flex flex-row items-center justify-between pb-2">
					<CardTitle className="text-sm font-semibold uppercase tracking-wider text-muted-foreground">
						Clicks (30d)
					</CardTitle>
					<MousePointerClick className="h-4 w-4 text-muted-foreground" />
				</CardHeader>
				<CardContent>
					<p className="text-3xl font-bold tracking-tight text-foreground">
						{totalClicks.toLocaleString()}
					</p>
				</CardContent>
			</Card>
		</div>
	);
}
