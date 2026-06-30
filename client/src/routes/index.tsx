import { createFileRoute, Navigate } from "@tanstack/react-router";
import { useAuth } from "@clerk/clerk-react";
import { useDashboard } from "@/hooks/use-dashboard";
import { LoadingState } from "@/components/shared/loading-state";
import { DashboardSkeleton } from "@/components/dashboard/dashboard-skeleton";
import { DashboardStats } from "@/components/dashboard/dashboard-stats";
import { DashboardChart } from "@/components/dashboard/dashboard-chart";
import { RecentLinks } from "@/components/dashboard/recent-links";
import { Button } from "@/components/ui/button";
import { Link2, AlertCircle } from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";

export const Route = createFileRoute("/")({
	component: DashboardPage,
});

function DashboardPage() {
	const { isSignedIn, isLoaded } = useAuth();
	const { data, isLoading, error } = useDashboard();

	if (!isLoaded) {
		return <LoadingState />;
	}

	if (!isSignedIn) {
		return <Navigate to="/login" />;
	}

	if (isLoading) {
		return (
			<main className="py-12 px-4 sm:px-6">
				<div className="max-w-6xl mx-auto">
					<DashboardSkeleton />
				</div>
			</main>
		);
	}

	if (error) {
		return (
			<main className="py-12 px-4 sm:px-6">
				<div className="max-w-6xl mx-auto">
					<Card>
						<CardContent className="flex flex-col items-center justify-center py-12">
							<AlertCircle className="h-8 w-8 text-destructive mb-4" />
							<p className="text-sm text-muted-foreground mb-4">
								Failed to load dashboard data. Please try again.
							</p>
							<Button variant="outline" onClick={() => window.location.reload()}>
								Retry
							</Button>
						</CardContent>
					</Card>
				</div>
			</main>
		);
	}

	const clicksOverTime =
		data?.clicks_over_time?.map((item) => ({
			name: new Date(item.date).toLocaleDateString("en-US", {
				month: "short",
				day: "numeric",
			}),
			clicks: item.clicks,
		})) ?? [];

	return (
		<main className="py-12 px-4 sm:px-6">
			<div className="max-w-6xl mx-auto">
				<div className="mb-8">
					<h1 className="text-3xl font-bold text-foreground">Dashboard</h1>
					<p className="text-muted-foreground mt-2">
						Overview of your shortened links and their performance.
					</p>
				</div>

				{data && data.total_links === 0 ? (
					<Card>
						<CardContent className="flex flex-col items-center justify-center py-16">
							<Link2 className="h-12 w-12 text-muted-foreground mb-4" />
							<h2 className="text-xl font-semibold text-foreground mb-2">No links yet</h2>
							<p className="text-sm text-muted-foreground mb-6 max-w-sm text-center">
								Create your first shortened link to start tracking clicks and analytics.
							</p>
							<Button asChild>
								<a href="/links">Create your first link</a>
							</Button>
						</CardContent>
					</Card>
				) : (
					<>
						<div className="mb-6">
							<DashboardStats
								totalLinks={data?.total_links ?? 0}
								activeLinks={data?.active_links ?? 0}
								totalClicks={data?.total_clicks ?? 0}
							/>
						</div>

						<div className="mb-6">
							<DashboardChart data={clicksOverTime} />
						</div>

						<RecentLinks links={data?.recent_links ?? []} />
					</>
				)}
			</div>
		</main>
	);
}
