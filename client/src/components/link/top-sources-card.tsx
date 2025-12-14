import { MoreVertical as MoreVerticalIcon } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { getReferrerIcon } from "@/lib/referrers";

interface TopSourcesCardProps {
	referrers: Array<{ referrer: string; clicks: number }>;
}

export function TopSourcesCard({ referrers }: TopSourcesCardProps) {
	const maxClicks = Math.max(...referrers.map((d) => d.clicks));

	return (
		<Card className='h-full flex flex-col'>
			<CardHeader>
				<CardTitle className='text-sm font-semibold uppercase tracking-wider text-muted-foreground'>
					Top Sources
				</CardTitle>
			</CardHeader>
			<CardContent className='flex-1'>
				<div className='space-y-6'>
					{referrers.map((item, idx) => {
						const IconComponent = getReferrerIcon(item.referrer);
						return (
							<div key={idx} className='group'>
								<div className='flex justify-between items-center text-sm mb-1.5'>
									<div className='flex items-center gap-2'>
										{IconComponent ? (
											<IconComponent className='w-4 h-4 text-muted-foreground' />
										) : (
											<MoreVerticalIcon className='w-4 h-4 text-muted-foreground' />
										)}
										<span className='font-medium text-foreground'>
											{item.referrer}
										</span>
									</div>
									<span className='text-muted-foreground'>{item.clicks}</span>
								</div>
								<div className='w-full bg-muted dark:bg-input/50 rounded-full h-2 overflow-hidden'>
									<div
										className='h-full bg-primary rounded-full transition-all duration-150'
										style={{
											width: `${(item.clicks / maxClicks) * 100}%`,
										}}
									/>
								</div>
							</div>
						);
					})}
				</div>
			</CardContent>
		</Card>
	);
}

