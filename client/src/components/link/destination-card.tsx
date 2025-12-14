import { useState } from "react";
import { Globe, Copy, CopyCheck, ExternalLink } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import {
	Tooltip,
	TooltipContent,
	TooltipTrigger,
} from "@/components/ui/tooltip";
import type { Url } from "@/types/url";

interface DestinationCardProps {
	url: Url;
}

export function DestinationCard({ url }: DestinationCardProps) {
	const [destinationCopied, setDestinationCopied] = useState(false);
	const [tooltipOpen, setTooltipOpen] = useState<boolean | undefined>(
		undefined
	);

	const handleCopyDestination = (e: React.MouseEvent) => {
		e.preventDefault();
		e.stopPropagation();
		navigator.clipboard.writeText(url.originalUrl);
		setDestinationCopied(true);
		setTooltipOpen(true);
		setTimeout(() => {
			setTooltipOpen(false);
			setDestinationCopied(false);
		}, 2000);
	};

	const handleTooltipOpenChange = (open: boolean) => {
		// Prevent closing if we're in the "copied" state (wait for timeout)
		if (!open && destinationCopied) {
			return; // Keep tooltip open
		}
		setTooltipOpen(open);
	};

	return (
		<Card>
			<CardHeader>
				<CardTitle className='text-sm font-semibold uppercase tracking-wider flex items-center gap-2 text-muted-foreground'>
					<Globe className='w-4 h-4' /> Destination
				</CardTitle>
			</CardHeader>
			<CardContent>
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
						<Tooltip open={tooltipOpen} onOpenChange={handleTooltipOpenChange}>
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
								<p>{destinationCopied ? "Copied!" : "Copy destination URL"}</p>
							</TooltipContent>
						</Tooltip>
						<ExternalLink className='w-4 h-4 text-muted-foreground group-hover:text-primary transition-colors' />
					</div>
				</div>
			</CardContent>
		</Card>
	);
}
