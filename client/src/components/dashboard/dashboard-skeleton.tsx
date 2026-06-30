import { Card, CardContent } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";

export function DashboardSkeleton() {
	return (
		<>
			<div className="mb-8">
				<Skeleton className="h-9 w-40 mb-2" />
				<Skeleton className="h-5 w-72" />
			</div>

			<div className="grid gap-4 sm:grid-cols-3 mb-6">
				{[1, 2, 3].map((i) => (
					<Card key={i}>
						<CardContent className="pt-6">
							<div className="flex items-center justify-between mb-2">
								<Skeleton className="h-4 w-24" />
								<Skeleton className="h-4 w-4" />
							</div>
							<Skeleton className="h-9 w-20" />
						</CardContent>
					</Card>
				))}
			</div>

			<Card className="mb-6">
				<CardContent className="pt-6">
					<Skeleton className="h-[300px] w-full" />
				</CardContent>
			</Card>

			<Card>
				<CardContent className="pt-6">
					<Skeleton className="h-5 w-28 mb-4" />
					<div className="space-y-3">
						{[1, 2, 3, 4, 5].map((i) => (
							<div key={i} className="flex items-center justify-between">
								<div className="flex-1">
									<Skeleton className="h-5 w-48 mb-1" />
									<Skeleton className="h-4 w-96" />
								</div>
								<Skeleton className="h-4 w-20" />
							</div>
						))}
					</div>
				</CardContent>
			</Card>
		</>
	);
}
