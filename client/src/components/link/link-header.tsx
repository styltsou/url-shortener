import { useState } from "react";
import { ArrowRight, Copy, CopyCheck, Calendar as CalendarIcon, PowerOff } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { formatDate } from "@/lib/mock-data";
import { toast } from "sonner";
import { useNavigate } from "@tanstack/react-router";
import type { Url } from "@/types/url";

interface LinkHeaderProps {
	url: Url;
}

export function LinkHeader({ url }: LinkHeaderProps) {
	const navigate = useNavigate();
	const [copied, setCopied] = useState(false);
	const isActive = url.isActive !== false;

	const handleCopy = () => {
		navigator.clipboard.writeText(`https://short.ly/${url.shortCode}`);
		setCopied(true);
		toast.success("Copied to clipboard");
		setTimeout(() => setCopied(false), 2000);
	};

	return (
		<div className='mb-8'>
			<Button
				variant='ghost'
				onClick={() => navigate({ to: "/" })}
				className='mb-4 -ml-2'
			>
				<ArrowRight className='w-4 h-4 rotate-180 mr-0.5' />
				Back to dashboard
			</Button>

			<div className='flex flex-col md:flex-row md:items-center justify-between gap-6'>
				<div>
					<div className='flex items-center gap-3'>
						<h1 className='text-3xl font-bold text-foreground tracking-tight'>
							short.ly/{url.shortCode}
						</h1>
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
						{!isActive && (
							<Badge
								variant='outline'
								className='text-[10px] font-bold uppercase tracking-wider flex items-center gap-1 border-primary text-primary'
							>
								<PowerOff className='w-3 h-3' />
								Inactive
							</Badge>
						)}
					</div>
					<div className='flex items-center gap-2 mt-2 text-muted-foreground text-sm'>
						<CalendarIcon className='w-4 h-4' />
						<span>Created {formatDate(url.createdAt)}</span>
					</div>
				</div>
			</div>
		</div>
	);
}

