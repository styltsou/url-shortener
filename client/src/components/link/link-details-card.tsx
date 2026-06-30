import { useState } from "react";
import {
	Globe,
	Copy,
	CopyCheck,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import {
	Tooltip,
	TooltipContent,
	TooltipTrigger,
} from "@/components/ui/tooltip";
import { ExpirationSection } from "./expiration-section";
import { TagsSection } from "./tags-section";
import type { Url } from "@/types/url";

interface LinkDetailsCardProps {
	url: Url;
}

export function LinkDetailsCard({ url }: LinkDetailsCardProps) {
	const [destinationCopied, setDestinationCopied] = useState(false);
	const [destinationTooltipOpen, setDestinationTooltipOpen] = useState<
		boolean | undefined
	>(undefined);

	const handleCopyDestination = (e: React.MouseEvent) => {
		e.preventDefault();
		e.stopPropagation();
		navigator.clipboard.writeText(url.originalUrl);
		setDestinationCopied(true);
		setDestinationTooltipOpen(true);
		setTimeout(() => {
			setDestinationTooltipOpen(false);
			setDestinationCopied(false);
		}, 2000);
	};

	const handleDestinationTooltipOpenChange = (open: boolean) => {
		if (!open && destinationCopied) {
			return;
		}
		setDestinationTooltipOpen(open);
	};

	return (
		<div className='space-y-6'>
			<div className='rounded-lg border border-border bg-card text-card-foreground shadow-sm'>
				<div className='p-6'>
					<div className='flex items-center gap-2 mb-3'>
						<span className='text-sm font-semibold uppercase tracking-wider text-muted-foreground'>
							Destination
						</span>
					</div>
					<div className='flex items-center gap-3 p-4 bg-background rounded-lg border border-border hover:border-primary/50 transition-all group'>
						<div className='shrink-0 p-2 bg-muted rounded-md border border-border group-hover:border-primary/50 transition-colors'>
							<Globe className='w-4 h-4 text-muted-foreground group-hover:text-primary transition-colors' />
						</div>
						<a
							href={url.originalUrl}
							target='_blank'
							rel='noopener noreferrer'
							className='flex-1 text-foreground break-all group-hover:text-primary transition-colors font-medium min-w-0'
						>
							{url.originalUrl}
						</a>
						<div className='flex items-center gap-1 shrink-0'>
							<Tooltip
								open={destinationTooltipOpen}
								onOpenChange={handleDestinationTooltipOpenChange}
							>
								<TooltipTrigger asChild>
									<Button
										variant='ghost'
										size='icon'
										onClick={handleCopyDestination}
										className={`h-8 w-8 text-muted-foreground hover:text-foreground ${
											destinationCopied ? "bg-primary/10 text-primary" : ""
										}`}
									>
										{destinationCopied ? (
											<CopyCheck className='w-4 h-4 text-primary' />
										) : (
											<Copy className='w-4 h-4' />
										)}
									</Button>
								</TooltipTrigger>
								<TooltipContent>
									<p>
										{destinationCopied ? "Copied!" : "Copy destination URL"}
									</p>
								</TooltipContent>
							</Tooltip>
						</div>
					</div>
				</div>
			</div>

			<ExpirationSection url={url} />

			<TagsSection url={url} />
		</div>
	);
}