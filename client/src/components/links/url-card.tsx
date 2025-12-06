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
import { formatDate } from "@/lib/mock-data";
import { useNavigate } from "@tanstack/react-router";
import { toast } from "sonner";
import { useState } from "react";
import {
	BarChart3,
	Calendar,
	Copy,
	CopyCheck,
	MoreVertical,
	Power,
	PowerOff,
	Trash2,
} from "lucide-react";
import { useDeleteLink } from "@/hooks/use-links";

interface UrlCardProps {
	url: Url;
}

export function UrlCard({ url }: UrlCardProps) {
	const navigate = useNavigate();
	const deleteLink = useDeleteLink();
	const isExpired = url.expiresAt && new Date(url.expiresAt) < new Date();
	const isActive = url.isActive !== false; // Default to true if not set
	const shortLink = `short.ly/${url.shortCode}`;
	const [copied, setCopied] = useState(false);
	const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
	const [tagsPopoverOpen, setTagsPopoverOpen] = useState(false);

	const handleCopy = (e: React.MouseEvent) => {
		e.stopPropagation();
		navigator.clipboard.writeText(`https://${shortLink}`);
		setCopied(true);
		toast.success("Copied to clipboard");
		setTimeout(() => setCopied(false), 2000);
	};

	const handleClick = () => {
		navigate({ to: "/$shortcode", params: { shortcode: url.shortCode } });
	};

	const handleToggleActive = (e: React.MouseEvent) => {
		e.stopPropagation();
		// TODO: Call API to update link active status
		toast.success(isActive ? "Link deactivated" : "Link activated");
	};

	const handleDeleteClick = (e: React.MouseEvent) => {
		e.stopPropagation();
		setDeleteDialogOpen(true);
	};

	const handleDelete = async () => {
		await deleteLink.mutateAsync(url.id);
		setDeleteDialogOpen(false);
	};

	return (
		<Card
			onClick={handleClick}
			className="group relative cursor-pointer hover:shadow-md hover:border-primary/50 transition-all"
		>
			<CardContent className='py-1 px-4'>
				<div className='flex flex-col md:flex-row md:items-center justify-between gap-2'>
					<div className='flex-1 min-w-0'>
						<div className='flex items-center gap-2 mb-0.5 flex-wrap'>
							<div className='flex items-center gap-2'>
								<h3 className='text-lg font-bold text-foreground tracking-tight group-hover:text-primary transition-colors'>
									{shortLink}
								</h3>
								{!isActive && (
									<Badge
										variant='outline'
										className='text-[10px] font-bold uppercase tracking-wider flex items-center gap-1 border-primary text-primary'
									>
										<PowerOff className='w-3 h-3' />
										Inactive
									</Badge>
								)}
								{isExpired && (
									<Badge
										variant='outline'
										className='text-[10px] font-bold uppercase tracking-wider'
									>
										Expired
									</Badge>
								)}
							</div>
							{url.tags && url.tags.length > 0 && (
								<div className='flex items-center gap-1.5 flex-wrap'>
									{url.tags.length <= 3 ? (
										// Show all tags if 3 or fewer
										url.tags.map((tag) => (
											<Badge
												key={tag.id}
												variant='outline'
												className='text-xs px-2 py-0.5 h-6'
												onClick={(e) => {
													e.stopPropagation();
													// TODO: Filter by tag
												}}
											>
												{tag.name}
											</Badge>
										))
									) : (
										// Show 2 tags + popover if more than 3
										<>
											{url.tags.slice(0, 2).map((tag) => (
												<Badge
													key={tag.id}
													variant='outline'
													className='text-xs px-2 py-0.5 h-6'
													onClick={(e) => {
														e.stopPropagation();
														// TODO: Filter by tag
													}}
												>
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
															<Badge
																key={tag.id}
																variant='outline'
																className='text-xs px-2 py-0.5 h-6'
																onClick={(e) => {
																	e.stopPropagation();
																	// TODO: Filter by tag
																}}
															>
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
						<div className='flex items-center text-sm text-muted-foreground truncate'>
							<a
								href={url.originalUrl}
								target='_blank'
								rel='noopener noreferrer'
								onClick={(e) => e.stopPropagation()}
								className='truncate hover:text-foreground transition-colors'
							>
								{url.originalUrl}
							</a>
						</div>
					</div>

					<div className='flex items-center gap-4 text-sm text-muted-foreground'>
						<div className='flex items-center gap-1.5' title='Total Clicks'>
							<BarChart3 className='w-4 h-4 text-muted-foreground' />
							<span className='font-semibold text-foreground'>
								{url.clicks.toLocaleString()}
							</span>
						</div>
						<div className='hidden sm:flex items-center gap-1.5'>
							<Calendar className='w-4 h-4 text-muted-foreground' />
							<span>{formatDate(url.createdAt)}</span>
						</div>

						<div className='flex items-center gap-2 pl-4 border-l border-border'>
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
							<DropdownMenu>
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
								<DropdownMenuContent
									align='end'
									onClick={(e) => e.stopPropagation()}
								>
									<DropdownMenuItem onClick={handleToggleActive}>
										{isActive ? (
											<>
												<PowerOff className='w-4 h-4 mr-2' />
												Deactivate
											</>
										) : (
											<>
												<Power className='w-4 h-4 mr-2' />
												Activate
											</>
										)}
									</DropdownMenuItem>
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

			<AlertDialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
				<AlertDialogContent onClick={(e) => e.stopPropagation()}>
					<AlertDialogHeader>
						<AlertDialogTitle>Are you absolutely sure?</AlertDialogTitle>
						<AlertDialogDescription>
							This action cannot be undone. This will permanently delete the
							short URL <strong>short.ly/{url.shortCode}</strong> and all its
							associated data.
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
							{deleteLink.isPending ? "Deleting..." : "Delete"}
						</AlertDialogAction>
					</AlertDialogFooter>
				</AlertDialogContent>
			</AlertDialog>
		</Card>
	);
}
