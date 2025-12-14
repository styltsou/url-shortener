import { useState } from "react";
import {
	Copy,
	CopyCheck,
	Calendar as CalendarIcon,
	PowerOff,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
	Tooltip,
	TooltipContent,
	TooltipTrigger,
} from "@/components/ui/tooltip";
import { formatDate, getTimePeriod } from "@/lib/mock-data";
import type { Url } from "@/types/url";
import { SHORT_DOMAIN } from "@/lib/constants";

interface LinkHeaderProps {
	url: Url;
}

export function LinkHeader({ url }: LinkHeaderProps) {
	const [copied, setCopied] = useState(false);
	const [tooltipOpen, setTooltipOpen] = useState<boolean | undefined>(
		undefined
	);
	const isExpired = url.expiresAt && new Date(url.expiresAt) < new Date();
	const isManuallyInactive = url.isActive === false;
	// A link is effectively inactive if it's manually deactivated OR expired
	const isEffectivelyInactive = isManuallyInactive || isExpired;

	const handleCopy = () => {
		navigator.clipboard.writeText(`https://${SHORT_DOMAIN}/${url.shortCode}`);
		setCopied(true);
		setTooltipOpen(true);
		setTimeout(() => {
			setTooltipOpen(false);
			setCopied(false);
		}, 2000);
	};

	const handleTooltipOpenChange = (open: boolean) => {
		// Prevent closing if we're in the "copied" state (wait for timeout)
		if (!open && copied) {
			return; // Keep tooltip open
		}
		setTooltipOpen(open);
	};

	return (
		<div className='flex flex-col gap-6'>
			<div>
				<div className='flex items-center gap-3'>
					<h1 className='text-3xl font-bold text-foreground tracking-tight'>
						{SHORT_DOMAIN}/{url.shortCode}
					</h1>
					<Tooltip open={tooltipOpen} onOpenChange={handleTooltipOpenChange}>
						<TooltipTrigger asChild>
							<Button
								variant='ghost'
								size='icon'
								onClick={handleCopy}
								className='text-muted-foreground hover:text-foreground'
							>
								{copied ? (
									<CopyCheck className='w-6 h-6 text-primary' />
								) : (
									<Copy className='w-6 h-6' />
								)}
							</Button>
						</TooltipTrigger>
						<TooltipContent>
							<p>{copied ? "Copied!" : "Copy link"}</p>
						</TooltipContent>
					</Tooltip>
					{isEffectivelyInactive && (
						<Badge
							variant='outline'
							className='text-[10px] font-bold uppercase tracking-wider flex items-center gap-1 border-primary text-primary'
						>
							<PowerOff className='w-3 h-3' />
							{isExpired ? "Expired" : "Inactive"}
						</Badge>
					)}
				</div>
				<div className='flex items-center gap-2 mt-2 text-muted-foreground text-sm'>
					<CalendarIcon className='w-4 h-4' />
					<Tooltip>
						<TooltipTrigger asChild>
							<span>Created {getTimePeriod(url.createdAt)}</span>
						</TooltipTrigger>
						<TooltipContent>
							<p>Created {formatDate(url.createdAt)}</p>
						</TooltipContent>
					</Tooltip>
				</div>
			</div>
		</div>
	);
}
