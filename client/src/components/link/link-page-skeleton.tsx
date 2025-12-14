import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";

export function LinkPageSkeleton() {
	return (
		<>
			{/* Header and Actions skeleton */}
			<div className='mb-6'>
				<div className='flex items-center justify-between gap-4'>
					<div className='flex-1'>
						<div className='flex flex-col gap-6'>
							<div>
								<div className='flex items-center gap-3'>
									<Skeleton className='h-9 w-64' />
									<Skeleton className='h-10 w-10 rounded-md' />
									<Skeleton className='h-6 w-20 rounded-full' />
								</div>
								<div className='flex items-center gap-2 mt-2'>
									<Skeleton className='h-4 w-4' />
									<Skeleton className='h-4 w-32' />
								</div>
							</div>
						</div>
					</div>
					<div className='flex items-center'>
						<Skeleton className='h-10 w-10 rounded-md' />
					</div>
				</div>
			</div>

			<div className='grid grid-cols-1 lg:grid-cols-3 gap-6'>
				{/* Main Content skeleton */}
				<div className='lg:col-span-2 space-y-6'>
					{/* Link Details Card skeleton */}
					<Card>
						<CardContent className='space-y-0'>
							{/* Destination Section */}
							<div className='pb-6'>
								<div className='flex items-center gap-2 mb-3'>
									<Skeleton className='h-4 w-24' />
								</div>
								<div className='flex items-center gap-3 p-4 bg-background rounded-lg border border-border'>
									<Skeleton className='h-10 w-10 rounded-md shrink-0' />
									<Skeleton className='h-5 flex-1' />
									<Skeleton className='h-8 w-8 rounded-md shrink-0' />
								</div>
							</div>

							<div className='h-px bg-border' />

							{/* Expiration Section */}
							<div className='py-6'>
								<div className='flex items-center justify-between mb-3'>
									<Skeleton className='h-4 w-24' />
									<Skeleton className='h-8 w-8 rounded-md' />
								</div>
								<Skeleton className='h-10 w-full' />
							</div>

							<div className='h-px bg-border' />

							{/* Tags Section */}
							<div className='py-6'>
								<div className='flex items-center gap-2 mb-3'>
									<Skeleton className='h-4 w-16' />
								</div>
								<div className='flex items-center gap-2 flex-wrap'>
									<Skeleton className='h-8 w-48' />
									<Skeleton className='h-6 w-20 rounded-full' />
									<Skeleton className='h-6 w-24 rounded-full' />
								</div>
							</div>
						</CardContent>
					</Card>

					{/* Performance Chart Card skeleton */}
					<Card>
						<CardHeader>
							<div className='flex items-center justify-between'>
								<Skeleton className='h-4 w-24' />
								<Skeleton className='h-10 w-[140px]' />
							</div>
						</CardHeader>
						<CardContent>
							<Skeleton className='h-[300px] w-full' />
						</CardContent>
					</Card>
				</div>

				{/* Sidebar Stats skeleton */}
				<div className='grid grid-cols-2 lg:grid-cols-1 lg:grid-rows-[auto_1fr] gap-6'>
					{/* Total Clicks Card skeleton */}
					<Card>
						<CardHeader>
							<Skeleton className='h-4 w-32' />
						</CardHeader>
						<CardContent>
							<div className='flex items-baseline justify-between gap-4'>
								<Skeleton className='h-9 w-24' />
								<Skeleton className='h-4 w-32' />
							</div>
						</CardContent>
					</Card>

					{/* Top Sources Card skeleton */}
					<Card className='h-full flex flex-col'>
						<CardHeader>
							<Skeleton className='h-4 w-28' />
						</CardHeader>
						<CardContent className='flex-1'>
							<div className='space-y-6'>
								{[1, 2, 3, 4].map((i) => (
									<div key={i} className='group'>
										<div className='flex justify-between items-center text-sm mb-1.5'>
											<div className='flex items-center gap-2'>
												<Skeleton className='h-4 w-4' />
												<Skeleton className='h-4 w-24' />
											</div>
											<Skeleton className='h-4 w-8' />
										</div>
										<Skeleton className='h-2 w-full rounded-full' />
									</div>
								))}
							</div>
						</CardContent>
					</Card>
				</div>
			</div>
		</>
	);
}
