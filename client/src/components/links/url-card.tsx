import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
	DropdownMenu,
	DropdownMenuContent,
	DropdownMenuItem,
	DropdownMenuSeparator,
	DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
	Popover,
	PopoverContent,
	PopoverTrigger,
} from "@/components/ui/popover";
import {
	Tooltip,
	TooltipContent,
	TooltipTrigger,
} from "@/components/ui/tooltip";
import {
	AlertDialog,
	AlertDialogAction,
	AlertDialogCancel,
	AlertDialogContent,
	AlertDialogDescription,
	AlertDialogFooter,
	AlertDialogHeader,
	AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import type { Url } from "@/types/url";
import { formatDateTime, getTimePeriod } from "@/lib/mock-data";
import { useNavigate } from "@tanstack/react-router";
import { useState } from "react";
import {
	BarChart3,
	Clock,
	Copy,
	CopyCheck,
	ExternalLink,
	MoreVertical,
	Power,
	PowerOff,
	Trash2,
} from "lucide-react";
import { useDeleteLink, useUpdateLink } from "@/hooks/use-links";
import {
	SHORT_DOMAIN,
	MAX_VISIBLE_TAGS,
	EXPIRATION_WARNING_HOURS,
} from "@/lib/constants";

interface UrlCardProps {
	url: Url;
}

export function UrlCard({ url }: UrlCardProps) {
	const navigate = useNavigate();
	const deleteLink = useDeleteLink();
	const updateLink = useUpdateLink();
	const isExpired = url.expiresAt && new Date(url.expiresAt) < new Date();
	const isManuallyInactive = url.isActive === false;
	// A link is effectively inactive if it's manually deactivated OR expired
	const isEffectivelyInactive = isManuallyInactive || isExpired;
	const shortLink = `${SHORT_DOMAIN}/${url.shortCode}`;

	// Determine expiration color based on status
	const expirationColorClass = url.expiresAt
		? isExpired
			? "text-destructive"
			: new Date(url.expiresAt) <
			  new Date(Date.now() + EXPIRATION_WARNING_HOURS * 60 * 60 * 1000)
			? "text-orange-500"
			: "text-muted-foreground"
		: "text-muted-foreground";

	// Determine expiration display text (relative time if within 7 days, otherwise absolute)
	const expirationDisplayText = url.expiresAt
		? (() => {
				const expiresAt = new Date(url.expiresAt);
				const now = new Date();
				const diffDays = Math.abs(
					(expiresAt.getTime() - now.getTime()) / (1000 * 60 * 60 * 24)
				);
				// Show relative time if within 7 days
				if (diffDays <= 7) {
					return getTimePeriod(url.expiresAt);
				}
				return formatDateTime(url.expiresAt);
		  })()
		: null;
	const [copied, setCopied] = useState(false);
	const [tooltipOpen, setTooltipOpen] = useState<boolean | undefined>(
		undefined
	);
	const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
	const [tagsPopoverOpen, setTagsPopoverOpen] = useState(false);
	const [linkButtonHovered, setLinkButtonHovered] = useState(false);

	const handleCopy = (e: React.MouseEvent) => {
		e.stopPropagation();
		navigator.clipboard.writeText(`https://${shortLink}`);
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

	const handleClick = () => {
		navigate({ to: "/links/$shortcode", params: { shortcode: url.shortCode } });
	};

	const handleToggleActive = async (e: React.MouseEvent) => {
		e.stopPropagation();
		try {
			await updateLink.mutateAsync({
				id: url.id,
				data: {
					is_active: !isManuallyInactive,
				},
			});
		} catch (error) {
			// Error is handled by the hook
		}
	};

	const handleDeleteClick = (e: React.MouseEvent) => {
		e.stopPropagation();
		setDeleteDialogOpen(true);
	};

	const handleDelete = async () => {
		try {
			await deleteLink.mutateAsync(url.id);
			// Close dialog only after successful deletion
			setDeleteDialogOpen(false);
		} catch (error) {
			// Error is handled by the hook, keep dialog open so user can see the error
		}
	};

	const handleOpenOriginalUrl = (e: React.MouseEvent) => {
		e.stopPropagation();
		window.open(url.originalUrl, "_blank", "noopener,noreferrer");
	};

	return (
		<Card
			onClick={handleClick}
			className='group relative cursor-pointer hover:shadow-md hover:border-primary/50 transition-all overflow-hidden'
		>
			{isEffectivelyInactive && (
				<div className='absolute top-0 left-0 z-10'>
					<div
						className={`text-[10px] font-bold uppercase tracking-wider px-3 py-1 shadow-md rounded-br-md ${
							isExpired
								? "bg-destructive text-destructive-foreground"
								: "bg-primary text-primary-foreground"
						}`}
					>
						{isExpired ? "Expired" : "Inactive"}
					</div>
				</div>
			)}
			<CardContent className='py-1 px-4'>
				<div className='flex flex-col md:flex-row md:items-center justify-between gap-2'>
					<div className='flex-1 min-w-0'>
						<div className='flex items-center gap-2 mb-0.5 flex-wrap'>
							<div className='flex items-center gap-2'>
								<h3 className='text-lg font-bold text-foreground tracking-tight group-hover:text-primary transition-colors'>
									{shortLink}
								</h3>
							</div>
							{url.tags && url.tags.length > 0 && (
								<div className='flex items-center gap-1.5 flex-wrap'>
									{url.tags.length <= MAX_VISIBLE_TAGS ? (
										// Show all tags if MAX_VISIBLE_TAGS or fewer
										url.tags.map((tag) => (
											<Badge key={tag.id} variant='outline'>
												{tag.name}
											</Badge>
										))
									) : (
										// Show 2 tags + popover if more than MAX_VISIBLE_TAGS
										<>
											{url.tags.slice(0, 2).map((tag) => (
												<Badge key={tag.id} variant='outline'>
													{tag.name}
												</Badge>
											))}
											<Popover
												open={tagsPopoverOpen}
												onOpenChange={setTagsPopoverOpen}
											>
												<PopoverTrigger asChild>
													<span
														className='text-[10px] text-muted-foreground cursor-pointer hover:text-foreground transition-colors'
														onMouseEnter={() => setTagsPopoverOpen(true)}
														onMouseLeave={() => setTagsPopoverOpen(false)}
														onClick={(e) => e.stopPropagation()}
													>
														+{url.tags.length - 2}
													</span>
												</PopoverTrigger>
												<PopoverContent
													className='w-auto p-2'
													side='bottom'
													align='start'
													onMouseEnter={() => setTagsPopoverOpen(true)}
													onMouseLeave={() => setTagsPopoverOpen(false)}
													onClick={(e) => e.stopPropagation()}
												>
													<div className='flex flex-wrap gap-1.5'>
														{url.tags.slice(2).map((tag) => (
															<Badge key={tag.id} variant='outline'>
																{tag.name}
															</Badge>
														))}
													</div>
												</PopoverContent>
											</Popover>
										</>
									)}
								</div>
							)}
						</div>
						<div className='flex items-center gap-2 text-sm text-muted-foreground truncate'>
							<span
								className={`truncate transition-colors ${
									linkButtonHovered ? "text-foreground" : ""
								}`}
							>
								{url.originalUrl}
							</span>
							<Tooltip>
								<TooltipTrigger asChild>
									<Button
										variant='ghost'
										size='icon'
										onClick={handleOpenOriginalUrl}
										onMouseEnter={() => setLinkButtonHovered(true)}
										onMouseLeave={() => setLinkButtonHovered(false)}
										className='h-6 w-6 shrink-0 text-muted-foreground hover:text-foreground'
									>
										<ExternalLink className='w-3.5 h-3.5' />
									</Button>
								</TooltipTrigger>
								<TooltipContent>
									<p>Open original URL</p>
								</TooltipContent>
							</Tooltip>
						</div>
					</div>

					<div className='flex items-center gap-4 text-sm text-muted-foreground'>
						<Tooltip>
							<TooltipTrigger asChild>
								<div className='flex items-center gap-1.5'>
									<BarChart3 className='w-4 h-4 text-muted-foreground' />
									<span className='font-semibold text-foreground'>
										{url.clicks.toLocaleString()}
									</span>
								</div>
							</TooltipTrigger>
							<TooltipContent>
								<p>Total Clicks</p>
							</TooltipContent>
						</Tooltip>
						{url.expiresAt && (
							<Tooltip>
								<TooltipTrigger asChild>
									<div className='hidden sm:flex items-center gap-1.5'>
										<Clock className={`w-4 h-4 ${expirationColorClass}`} />
										<span className={expirationColorClass}>
											{expirationDisplayText}
										</span>
									</div>
								</TooltipTrigger>
								<TooltipContent>
									<p>
										{isExpired
											? "Expired " + formatDateTime(url.expiresAt)
											: "Expires " + formatDateTime(url.expiresAt)}
									</p>
								</TooltipContent>
							</Tooltip>
						)}

						<div className='flex items-center gap-2 pl-4 border-l border-border'>
							<Tooltip
								open={tooltipOpen}
								onOpenChange={handleTooltipOpenChange}
							>
								<TooltipTrigger asChild>
									<Button
										variant='ghost'
										size='icon'
										onClick={handleCopy}
										className={`h-8 w-8 ${
											copied ? "bg-primary/10 text-primary" : ""
										}`}
									>
										{copied ? (
											<CopyCheck className='w-4 h-4 text-primary' />
										) : (
											<Copy className='w-4 h-4' />
										)}
									</Button>
								</TooltipTrigger>
								<TooltipContent>
									<p>{copied ? "Copied!" : "Copy link"}</p>
								</TooltipContent>
							</Tooltip>
							<DropdownMenu>
								<Tooltip>
									<TooltipTrigger asChild>
										<DropdownMenuTrigger asChild>
											<Button
												variant='ghost'
												size='icon'
												onClick={(e) => {
													e.stopPropagation();
												}}
												className='h-8 w-8'
											>
												<MoreVertical className='w-4 h-4' />
											</Button>
										</DropdownMenuTrigger>
									</TooltipTrigger>
									<TooltipContent>
										<p>More options</p>
									</TooltipContent>
								</Tooltip>
								<DropdownMenuContent
									align='end'
									onClick={(e) => e.stopPropagation()}
								>
									{!isExpired && (
										<DropdownMenuItem onClick={handleToggleActive}>
											{isManuallyInactive ? (
												<>
													<Power className='w-4 h-4 mr-2' />
													Activate
												</>
											) : (
												<>
													<PowerOff className='w-4 h-4 mr-2' />
													Deactivate
												</>
											)}
										</DropdownMenuItem>
									)}
									{isExpired && (
										<DropdownMenuItem disabled>
											<PowerOff className='w-4 h-4 mr-2' />
											Expired (cannot activate)
										</DropdownMenuItem>
									)}
									<DropdownMenuSeparator />
									<DropdownMenuItem
										onClick={handleDeleteClick}
										variant='destructive'
									>
										<Trash2 className='w-4 h-4 mr-2' />
										Delete link
									</DropdownMenuItem>
								</DropdownMenuContent>
							</DropdownMenu>
						</div>
					</div>
				</div>
			</CardContent>

			<AlertDialog
				open={deleteDialogOpen}
				onOpenChange={(open) => {
					// Prevent closing the dialog while deletion is in progress
					if (!open && deleteLink.isPending) {
						return;
					}
					setDeleteDialogOpen(open);
				}}
			>
				<AlertDialogContent onClick={(e) => e.stopPropagation()}>
					<AlertDialogHeader>
						<AlertDialogTitle>Are you absolutely sure?</AlertDialogTitle>
						<AlertDialogDescription>
							This action cannot be undone. This will permanently delete the
							short URL{" "}
							<strong>
								{SHORT_DOMAIN}/{url.shortCode}
							</strong>{" "}
							and all its associated data.
						</AlertDialogDescription>
					</AlertDialogHeader>
					<AlertDialogFooter>
						<AlertDialogCancel onClick={(e) => e.stopPropagation()}>
							Cancel
						</AlertDialogCancel>
						<AlertDialogAction
							onClick={(e) => {
								e.stopPropagation();
								handleDelete();
							}}
							className='bg-destructive text-white hover:bg-destructive/90 focus-visible:ring-destructive/20 dark:focus-visible:ring-destructive/40 dark:bg-destructive/60'
							disabled={deleteLink.isPending}
						>
							{deleteLink.isPending ? "Deleting" : "Delete"}
						</AlertDialogAction>
					</AlertDialogFooter>
				</AlertDialogContent>
			</AlertDialog>
		</Card>
	);
}
