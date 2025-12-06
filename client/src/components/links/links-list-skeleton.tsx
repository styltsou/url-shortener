import { Card } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";

export function LinksListSkeleton() {
	return (
		<div className='space-y-2'>
			{[1, 2, 3].map((i) => (
				<Card key={i} className='py-1 px-4'>
					<div className='flex flex-col md:flex-row md:items-center justify-between gap-2'>
						<div className='flex-1 space-y-1'>
							<Skeleton className='h-7 w-48 sm:w-64' />
							<Skeleton className='h-4 w-full sm:w-96' />
						</div>
						<div className='flex items-center gap-6 mt-2 md:mt-0'>
							<Skeleton className='h-5 w-16' />
							<Skeleton className='h-5 w-24 hidden sm:block' />
							<div className='flex gap-2 pl-4 border-l border-border'>
								<Skeleton className='h-8 w-8 rounded-md' />
								<Skeleton className='h-8 w-8 rounded-md' />
							</div>
						</div>
					</div>
				</Card>
			))}
		</div>
	);
}

