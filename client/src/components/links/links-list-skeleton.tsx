import { Card, CardContent } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";

export function LinksListSkeleton() {
	return (
		<div className='space-y-2'>
			{[1, 2, 3, 4, 5].map((i) => (
				<Card key={i}>
					<CardContent className='py-1 px-4'>
						<div className='flex flex-col md:flex-row md:items-center justify-between gap-2'>
							<div className='flex-1 min-w-0'>
								<div className='flex items-center gap-2 mb-0.5 flex-wrap'>
									<Skeleton className='h-7 w-48 sm:w-64' />
								</div>
								<div className='flex items-center gap-2 text-sm'>
									<Skeleton className='h-4 w-full sm:w-96' />
									<Skeleton className='h-6 w-6 shrink-0 rounded-md' />
								</div>
							</div>
							<div className='flex items-center gap-4 text-sm'>
								<Skeleton className='h-5 w-16' />
								<Skeleton className='h-5 w-24 hidden sm:block' />
								<div className='flex items-center gap-2 pl-4 border-l border-border'>
									<Skeleton className='h-8 w-8 rounded-md' />
									<Skeleton className='h-8 w-8 rounded-md' />
								</div>
							</div>
						</div>
					</CardContent>
				</Card>
			))}
		</div>
	);
}

