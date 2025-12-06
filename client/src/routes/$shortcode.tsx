import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { useAuth } from "@clerk/clerk-react";
import { Navigate } from "@tanstack/react-router";
import { useState, useEffect, useMemo } from "react";
import {
	ArrowRight,
	Calendar as CalendarIcon,
	Copy,
	CopyCheck,
	Edit,
	Save,
	Globe,
	Clock,
	TrendingUp,
	Trash2,
	Loader2,
	ExternalLink,
	Power,
	PowerOff,
	Facebook,
	Instagram,
	Twitter,
	Linkedin,
	MessageCircle,
	Mail,
	Newspaper,
	MoreVertical as MoreVerticalIcon,
	Tag as TagIcon,
	X,
} from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "@/components/ui/select";
import {
	Popover,
	PopoverContent,
	PopoverTrigger,
} from "@/components/ui/popover";
import {
	DropdownMenu,
	DropdownMenuContent,
	DropdownMenuItem,
	DropdownMenuSeparator,
	DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Calendar } from "@/components/ui/calendar";
import {
	AlertDialog,
	AlertDialogAction,
	AlertDialogCancel,
	AlertDialogContent,
	AlertDialogDescription,
	AlertDialogFooter,
	AlertDialogHeader,
	AlertDialogTitle,
	AlertDialogTrigger,
} from "@/components/ui/alert-dialog";
import { ClicksGraph } from "@/components/ClicksGraph";
import { formatDate, generateMockAnalytics, MOCK_TAGS } from "@/lib/mock-data";
import { useLinks, useUpdateLink, useDeleteLink } from "@/hooks/use-links";
import { toast } from "sonner";
import { Badge } from "@/components/ui/badge";
import { TagCombobox } from "@/components/TagCombobox";
import type { Tag } from "@/types/url";

export const Route = createFileRoute("/$shortcode")({
	component: LinkDetailPage,
});

// Known referrer sources with their icons
const KNOWN_REFERRERS = [
	{ match: /^direct|none$/i, icon: Globe, label: "Direct" },
	{
		match: /twitter|twitter\.com|x\.com/i,
		icon: Twitter,
		label: "X (Twitter)",
	},
	{ match: /instagram|instagram\.com/i, icon: Instagram, label: "Instagram" },
	{ match: /facebook|facebook\.com/i, icon: Facebook, label: "Facebook" },
	{ match: /linkedin|linkedin\.com/i, icon: Linkedin, label: "LinkedIn" },
	{ match: /reddit|reddit\.com/i, icon: MessageCircle, label: "Reddit" },
	{ match: /^email$/i, icon: Mail, label: "Email" },
	{ match: /newsletter/i, icon: Newspaper, label: "Newsletter" },
];

// Get icon component for a referrer
function getReferrerIcon(referrer: string) {
	const normalized = referrer.toLowerCase().trim();
	const known = KNOWN_REFERRERS.find((r) => r.match.test(normalized));
	return known ? known.icon : null;
}

// Get display label for a referrer
function getReferrerLabel(referrer: string): string {
	const normalized = referrer.toLowerCase().trim();
	const known = KNOWN_REFERRERS.find((r) => r.match.test(normalized));
	return known ? known.label : referrer;
}

// Process referrers data: merge unknown sources into "Other"
function processReferrersData(
	referrersData: Array<{ referrer: string; clicks: number }>
): Array<{ referrer: string; clicks: number }> {
	const known: Array<{ referrer: string; clicks: number }> = [];
	let otherClicks = 0;

	referrersData.forEach((item) => {
		const icon = getReferrerIcon(item.referrer);
		if (icon) {
			// Known referrer - use standardized label
			const label = getReferrerLabel(item.referrer);
			const existing = known.find((k) => k.referrer === label);
			if (existing) {
				existing.clicks += item.clicks;
			} else {
				known.push({ referrer: label, clicks: item.clicks });
			}
		} else {
			// Unknown referrer - add to "Other"
			otherClicks += item.clicks;
		}
	});

	// Sort by clicks descending
	known.sort((a, b) => b.clicks - a.clicks);

	// Add "Other" if there are unknown sources
	if (otherClicks > 0) {
		known.push({ referrer: "Other", clicks: otherClicks });
	}

	return known;
}

function LinkDetailPage() {
	const { shortcode } = Route.useParams();
	const navigate = useNavigate();
	const { isSignedIn, isLoaded } = useAuth();
	const { data: urls = [], isLoading: isLoadingLinks } = useLinks();
	const updateLink = useUpdateLink();
	const deleteLink = useDeleteLink();
	const [isEditing, setIsEditing] = useState(false);
	const [copied, setCopied] = useState(false);
	const [originalUrlInput, setOriginalUrlInput] = useState("");
	const [expirationDate, setExpirationDate] = useState<Date | undefined>(
		undefined
	);
	const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
	const [timeRange, setTimeRange] = useState<"7days" | "30days">("7days");
	const [selectedTags, setSelectedTags] = useState<Tag[]>([]);
	const [availableTags, setAvailableTags] = useState<Tag[]>(MOCK_TAGS);
	const [isActive, setIsActive] = useState(true);
	const [destinationCopied, setDestinationCopied] = useState(false);
	const [deactivateDialogOpen, setDeactivateDialogOpen] = useState(false);

	const url = urls.find((u) => u.shortCode === shortcode);

	// Initialize active status from URL
	useEffect(() => {
		if (url) {
			setIsActive(url.isActive !== false);
		}
	}, [url]);

	// Generate analytics data based on selected time range
	const analyticsData = useMemo(() => {
		if (!url) return null;
		return generateMockAnalytics(url, timeRange);
	}, [url, timeRange]);

	// Process referrers data to merge unknown sources into "Other"
	const processedReferrers = useMemo(() => {
		const referrersData =
			analyticsData?.referrers_data || url?.analytics.referrers_data || [];
		return processReferrersData(referrersData);
	}, [analyticsData, url]);

	// Initialize form state when entering edit mode
	useEffect(() => {
		if (isEditing && url && !originalUrlInput) {
			setOriginalUrlInput(url.originalUrl);
			setExpirationDate(url.expiresAt ? new Date(url.expiresAt) : undefined);
			setSelectedTags(url.tags || []);
		}
		if (!isEditing) {
			setOriginalUrlInput("");
			setExpirationDate(undefined);
			setSelectedTags([]);
		}
	}, [isEditing, url, originalUrlInput]);

	if (!isLoaded) {
		return (
			<div className='flex min-h-screen items-center justify-center'>
				<div className='text-lg'>Loading...</div>
			</div>
		);
	}

	if (!isSignedIn) {
		return <Navigate to='/login' />;
	}

	if (isLoadingLinks) {
		return (
			<main className='py-12 px-4 sm:px-6'>
				<div className='max-w-6xl mx-auto text-center py-20'>
					<div className='text-lg text-muted-foreground'>Loading link...</div>
				</div>
			</main>
		);
	}

	if (!url) {
		return (
			<main className='py-12 px-4 sm:px-6'>
				<div className='max-w-6xl mx-auto text-center py-20'>
					<h1 className='text-3xl font-bold mb-4'>Link not found</h1>
					<Button onClick={() => navigate({ to: "/" })}>
						Back to dashboard
					</Button>
				</div>
			</main>
		);
	}

	const handleSave = async () => {
		if (!url) return;

		await updateLink.mutateAsync({
			id: url.id,
			data: {
				expires_at: expirationDate
					? expirationDate.toISOString().split("T")[0]
					: null,
				// TODO: Add original_url update when backend supports it
			},
		});
		setIsEditing(false);
	};

	const handleDelete = async () => {
		if (!url) return;
		await deleteLink.mutateAsync(url.id);
		navigate({ to: "/" });
	};

	const handleCopy = () => {
		navigator.clipboard.writeText(`https://short.ly/${url.shortCode}`);
		setCopied(true);
		toast.success("Copied to clipboard");
		setTimeout(() => setCopied(false), 2000);
	};

	const handleCopyDestination = (e: React.MouseEvent) => {
		e.preventDefault();
		e.stopPropagation();
		navigator.clipboard.writeText(url.originalUrl);
		setDestinationCopied(true);
		toast.success("Destination URL copied to clipboard");
		setTimeout(() => setDestinationCopied(false), 2000);
	};

	// Mock function to create a new tag
	// TODO: Replace with actual API call when backend endpoint is ready
	const handleCreateTag = async (tagName: string): Promise<Tag> => {
		// Simulate API delay
		await new Promise((resolve) => setTimeout(resolve, 500));

		// Mock POST request to create tag
		// In the real implementation, this would be:
		// const response = await fetch('/api/tags', {
		//   method: 'POST',
		//   headers: { 'Content-Type': 'application/json' },
		//   body: JSON.stringify({ name: tagName }),
		// })
		// const newTag = await response.json()

		// For now, create a mock tag
		const newTag: Tag = {
			id: `tag-${Date.now()}-${Math.random().toString(36).substring(2, 9)}`,
			name: tagName.trim(),
		};

		// Add to available tags
		setAvailableTags((prev) => {
			// Check if tag already exists (case-insensitive)
			const exists = prev.some(
				(tag) => tag.name.toLowerCase() === tagName.toLowerCase().trim()
			);
			if (exists) {
				return prev;
			}
			return [...prev, newTag];
		});

		toast.success(`Tag "${tagName}" created successfully`);
		return newTag;
	};

	const handleTagSelect = (tag: Tag) => {
		if (!selectedTags.find((t) => t.id === tag.id)) {
			setSelectedTags([...selectedTags, tag]);
		}
	};

	const handleToggleActive = async () => {
		// TODO: Call API to update link active status
		setIsActive(!isActive);
		toast.success(isActive ? "Link deactivated" : "Link activated");
		setDeactivateDialogOpen(false);
	};

	return (
		<main className='py-12 px-4 sm:px-6'>
			<div className='max-w-6xl mx-auto'>
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

						<div className='flex gap-2'>
							{isEditing ? (
								<>
									<Button variant='secondary' onClick={() => setIsEditing(false)}>
										Cancel
									</Button>
									<Button onClick={handleSave} disabled={updateLink.isPending}>
										{updateLink.isPending ? (
											<Loader2 className='w-4 h-4 animate-spin' />
										) : (
											<Save className='w-4 h-4' />
										)}{" "}
										Save Changes
									</Button>
								</>
							) : (
								<>
									<Button variant='ghost' onClick={() => setIsEditing(true)}>
										<Edit className='w-4 h-4' /> Edit
									</Button>
									<DropdownMenu>
										<DropdownMenuTrigger asChild>
											<Button variant='secondary'>
												<MoreVerticalIcon className='w-4 h-4' />
											</Button>
										</DropdownMenuTrigger>
										<DropdownMenuContent align='end'>
											{isActive ? (
												<AlertDialog
													open={deactivateDialogOpen}
													onOpenChange={setDeactivateDialogOpen}
												>
													<AlertDialogTrigger asChild>
														<DropdownMenuItem
															onSelect={(e) => {
																e.preventDefault();
																setDeactivateDialogOpen(true);
															}}
														>
															<PowerOff className='w-4 h-4 mr-2' />
															Deactivate
														</DropdownMenuItem>
													</AlertDialogTrigger>
													<AlertDialogContent>
														<AlertDialogHeader>
															<AlertDialogTitle>
																Deactivate link?
															</AlertDialogTitle>
															<AlertDialogDescription>
																This will deactivate the short URL{" "}
																<strong>short.ly/{url.shortCode}</strong>. The
																link will stop working and redirects will fail.
																You can reactivate it at any time.
															</AlertDialogDescription>
														</AlertDialogHeader>
														<AlertDialogFooter>
															<AlertDialogCancel>Cancel</AlertDialogCancel>
															<AlertDialogAction onClick={handleToggleActive}>
																Deactivate
															</AlertDialogAction>
														</AlertDialogFooter>
													</AlertDialogContent>
												</AlertDialog>
											) : (
												<DropdownMenuItem onClick={handleToggleActive}>
													<Power className='w-4 h-4 mr-2' />
													Activate
												</DropdownMenuItem>
											)}
											<DropdownMenuSeparator />
											<DropdownMenuItem
												onSelect={(e) => {
													e.preventDefault();
													setDeleteDialogOpen(true);
												}}
												variant='destructive'
											>
												<Trash2 className='w-4 h-4 mr-2' />
												Delete link
											</DropdownMenuItem>
										</DropdownMenuContent>
									</DropdownMenu>
									<AlertDialog
										open={deleteDialogOpen}
										onOpenChange={setDeleteDialogOpen}
									>
										<AlertDialogContent>
											<AlertDialogHeader>
												<AlertDialogTitle>
													Are you absolutely sure?
												</AlertDialogTitle>
												<AlertDialogDescription>
													This action cannot be undone. This will permanently
													delete the short URL{" "}
													<strong>short.ly/{url.shortCode}</strong> and all its
													associated data.
												</AlertDialogDescription>
											</AlertDialogHeader>
											<AlertDialogFooter>
												<AlertDialogCancel>Cancel</AlertDialogCancel>
												<AlertDialogAction
													onClick={handleDelete}
													className='bg-destructive text-white hover:bg-destructive/90 focus-visible:ring-destructive/20 dark:focus-visible:ring-destructive/40 dark:bg-destructive/60'
												>
													Delete
												</AlertDialogAction>
											</AlertDialogFooter>
										</AlertDialogContent>
									</AlertDialog>
								</>
							)}
						</div>
					</div>
				</div>

				<div className='grid grid-cols-1 lg:grid-cols-3 gap-6'>
					{/* Main Content */}
					<div className='lg:col-span-2 space-y-6'>
						{/* Destination Card */}
						<Card>
							<CardHeader>
								<CardTitle className='text-sm font-semibold uppercase tracking-wider flex items-center gap-2 text-muted-foreground'>
									<Globe className='w-4 h-4' /> Destination
								</CardTitle>
							</CardHeader>
							<CardContent>
								{isEditing ? (
									<Input
										type='url'
										value={originalUrlInput}
										onChange={(e) => setOriginalUrlInput(e.target.value)}
									/>
								) : (
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
											<ExternalLink className='w-4 h-4 text-muted-foreground group-hover:text-primary transition-colors' />
										</div>
									</div>
								)}

								<div className='mt-6 pt-6 border-t border-border'>
									<CardTitle className='text-sm font-semibold uppercase tracking-wider mb-4 flex items-center gap-2 text-muted-foreground'>
										<TagIcon className='w-4 h-4' /> Tags
									</CardTitle>
									{isEditing ? (
										<div className='space-y-3'>
											<div className='flex flex-wrap gap-2'>
												{selectedTags.map((tag) => (
													<Badge
														key={tag.id}
														variant='outline'
														className='text-sm gap-1.5 pr-1'
													>
														{tag.name}
														<button
															type='button'
															onClick={() =>
																setSelectedTags(
																	selectedTags.filter((t) => t.id !== tag.id)
																)
															}
															className='ml-0.5 hover:bg-muted rounded-full p-0.5'
														>
															<X className='w-3 h-3' />
														</button>
													</Badge>
												))}
											</div>
											<TagCombobox
												availableTags={availableTags}
												selectedTags={selectedTags}
												onTagSelect={handleTagSelect}
												onCreateTag={handleCreateTag}
												placeholder='Add a tag...'
												className='w-full'
											/>
										</div>
									) : (
										<div className='flex flex-wrap gap-2'>
											{url.tags && url.tags.length > 0 ? (
												url.tags.map((tag) => (
													<Badge
														key={tag.id}
														variant='outline'
														className='text-sm'
													>
														{tag.name}
													</Badge>
												))
											) : (
												<p className='text-sm text-muted-foreground'>No tags</p>
											)}
										</div>
									)}
								</div>

								<div className='mt-6 pt-6 border-t border-border'>
									<CardTitle className='text-sm font-semibold uppercase tracking-wider mb-4 flex items-center gap-2 text-muted-foreground'>
										<Clock className='w-4 h-4' /> Expiration
									</CardTitle>
									{isEditing ? (
										<Popover>
											<PopoverTrigger asChild>
												<Button
													variant='outline'
													className={`w-full justify-start text-left font-normal bg-input ${
														!expirationDate ? "text-muted-foreground" : ""
													}`}
												>
													<CalendarIcon className='mr-2 h-4 w-4' />
													{expirationDate ? (
														formatDate(expirationDate)
													) : (
														<span>Pick a date</span>
													)}
												</Button>
											</PopoverTrigger>
											<PopoverContent className='w-auto p-0' align='start'>
												<Calendar
													mode='single'
													selected={expirationDate}
													onSelect={setExpirationDate}
													disabled={(date) => date < new Date()}
													initialFocus
												/>
											</PopoverContent>
										</Popover>
									) : (
										<p
											className={`font-medium ${
												url.expiresAt && new Date(url.expiresAt) < new Date()
													? "text-destructive"
													: "text-muted-foreground"
											}`}
										>
											{url.expiresAt
												? formatDate(url.expiresAt)
												: "No expiration date set"}
										</p>
									)}
								</div>
							</CardContent>
						</Card>

						{/* Chart Card */}
						<Card>
							<CardHeader>
								<div className='flex items-center justify-between'>
									<CardTitle className='text-sm font-semibold uppercase tracking-wider text-muted-foreground'>
										Performance
									</CardTitle>
									<Select
										value={timeRange}
										onValueChange={(value: "7days" | "30days") =>
											setTimeRange(value)
										}
									>
										<SelectTrigger className='w-[140px]'>
											<SelectValue />
										</SelectTrigger>
										<SelectContent>
											<SelectItem value='7days'>Last 7 days</SelectItem>
											<SelectItem value='30days'>Last 30 days</SelectItem>
										</SelectContent>
									</Select>
								</div>
							</CardHeader>
							<CardContent>
								<ClicksGraph
									data={analyticsData?.clicks_data || url.analytics.clicks_data}
								/>
							</CardContent>
						</Card>
					</div>

					{/* Sidebar Stats */}
					<div className='grid grid-cols-2 lg:grid-cols-1 lg:grid-rows-[auto_1fr] gap-6'>
						<Card>
							<CardHeader>
								<CardTitle className='text-sm font-semibold uppercase tracking-wider text-muted-foreground'>
									Total Clicks
								</CardTitle>
							</CardHeader>
							<CardContent>
								<div className='flex items-baseline justify-between gap-4'>
									<p className='text-3xl font-bold tracking-tight text-foreground'>
										{url.clicks.toLocaleString()}
									</p>
									<div className='flex items-center gap-1.5 text-sm text-muted-foreground whitespace-nowrap'>
										<TrendingUp className='w-4 h-4' />
										<span>+12.5% this week</span>
									</div>
								</div>
							</CardContent>
						</Card>

						<Card className='h-full flex flex-col'>
							<CardHeader>
								<CardTitle className='text-sm font-semibold uppercase tracking-wider text-muted-foreground'>
									Top Sources
								</CardTitle>
							</CardHeader>
							<CardContent className='flex-1'>
								<div className='space-y-6'>
									{processedReferrers.map((item, idx) => {
										const maxClicks = Math.max(
											...processedReferrers.map((d) => d.clicks)
										);
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
													<span className='text-muted-foreground'>
														{item.clicks}
													</span>
												</div>
												<div className='w-full bg-muted/80 dark:bg-input/50 rounded-full h-2 overflow-hidden'>
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
					</div>
				</div>
			</div>
		</main>
	);
}
