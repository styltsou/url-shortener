import { useState, useEffect } from "react";
import {
	Globe,
	Copy,
	CopyCheck,
	Calendar as CalendarIcon,
	Save,
	X,
	Edit2,
	ChevronDownIcon,
} from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Spinner } from "@/components/ui/spinner";
import {
	Popover,
	PopoverContent,
	PopoverTrigger,
} from "@/components/ui/popover";
import { Calendar } from "@/components/ui/calendar";
import { Input } from "@/components/ui/input";
import { Separator } from "@/components/ui/separator";
import {
	Tooltip,
	TooltipContent,
	TooltipTrigger,
} from "@/components/ui/tooltip";
import { formatDate, formatDateTime } from "@/lib/mock-data";
import { useUpdateLink } from "@/hooks/use-links";
import { toast } from "sonner";
import { TagCombobox } from "./tag-combobox";
import {
	useTags,
	useCreateTag,
	useAddTagsToLink,
	useRemoveTagsFromLink,
} from "@/hooks/use-tags";
import { useBlockNavigation } from "@/hooks/use-block-navigation";
import type { Url, Tag } from "@/types/url";

interface LinkDetailsCardProps {
	url: Url;
}

export function LinkDetailsCard({ url }: LinkDetailsCardProps) {
	// Destination state
	const [destinationCopied, setDestinationCopied] = useState(false);
	const [destinationTooltipOpen, setDestinationTooltipOpen] = useState<
		boolean | undefined
	>(undefined);

	// Expiration state
	const [isEditingExpiration, setIsEditingExpiration] = useState(false);
	const [expirationDate, setExpirationDate] = useState<Date | undefined>(
		url.expiresAt ? new Date(url.expiresAt) : undefined
	);
	const [expirationTime, setExpirationTime] = useState<string>(() => {
		if (url.expiresAt) {
			const date = new Date(url.expiresAt);
			const hours = date.getUTCHours().toString().padStart(2, "0");
			const minutes = date.getUTCMinutes().toString().padStart(2, "0");
			return `${hours}:${minutes}`;
		}
		return "23:59";
	});

	// Tags state
	const [isEditingTags, setIsEditingTags] = useState(false);
	const [selectedTags, setSelectedTags] = useState<Tag[]>(url.tags || []);

	const updateLink = useUpdateLink();
	const { data, isLoading: isLoadingTags } = useTags();
	const availableTags = data ?? [];
	const createTag = useCreateTag();
	const addTags = useAddTagsToLink();
	const removeTags = useRemoveTagsFromLink();

	// Initialize expiration date and time when entering edit mode
	useEffect(() => {
		if (isEditingExpiration) {
			if (url.expiresAt) {
				const date = new Date(url.expiresAt);
				setExpirationDate(date);
				const hours = date.getUTCHours().toString().padStart(2, "0");
				const minutes = date.getUTCMinutes().toString().padStart(2, "0");
				setExpirationTime(`${hours}:${minutes}`);
			} else {
				setExpirationDate(undefined);
				setExpirationTime("23:59");
			}
		}
	}, [isEditingExpiration, url.expiresAt]);

	// Initialize selected tags when entering edit mode
	useEffect(() => {
		if (isEditingTags) {
			setSelectedTags(url.tags || []);
		}
	}, [isEditingTags, url.tags]);

	// Convert API tags to UI tags
	// React Compiler automatically memoizes this computation
	const availableTagsForUI = !Array.isArray(availableTags)
		? []
		: availableTags.map((tag) => ({
				id: tag.id,
				name: tag.name,
		  }));

	// Destination handlers
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
		// Prevent closing if we're in the "copied" state (wait for timeout)
		if (!open && destinationCopied) {
			return; // Keep tooltip open
		}
		setDestinationTooltipOpen(open);
	};

	// Expiration handlers
	const handleSaveExpiration = async () => {
		if (!expirationDate) {
			setIsEditingExpiration(false);
			return;
		}

		// Combine date and time into a single datetime
		const [hours, minutes] = expirationTime.split(":").map(Number);
		const combinedDate = new Date(expirationDate);
		combinedDate.setUTCHours(hours, minutes, 0, 0);

		const currentExpiration = url.expiresAt ? new Date(url.expiresAt) : null;

		// Check if the date/time actually changed
		if (
			currentExpiration?.getTime() === combinedDate.getTime() ||
			(!currentExpiration && !expirationDate)
		) {
			setIsEditingExpiration(false);
			return;
		}

		// Validate that expiration is in the future
		// Use < instead of <= to allow dates that are at least the current time
		// This accounts for network latency and ensures users can set expiration times
		// that are very close to the current time (e.g., 2 hours from now)
		if (combinedDate < new Date()) {
			toast.error("Expiration date and time must be in the future");
			return;
		}

		try {
			await updateLink.mutateAsync({
				id: url.id,
				data: {
					expires_at: combinedDate.toISOString(),
				},
			});
			setIsEditingExpiration(false);
		} catch (error) {
			// Error is handled by the hook
		}
	};

	const handleCancelExpiration = () => {
		if (url.expiresAt) {
			const date = new Date(url.expiresAt);
			setExpirationDate(date);
			const hours = date.getUTCHours().toString().padStart(2, "0");
			const minutes = date.getUTCMinutes().toString().padStart(2, "0");
			setExpirationTime(`${hours}:${minutes}`);
		} else {
			setExpirationDate(undefined);
			setExpirationTime("23:59");
		}
		setIsEditingExpiration(false);
	};

	// Tags handlers
	const handleCreateTag = async (tagName: string): Promise<Tag> => {
		const newTag = await createTag.mutateAsync(tagName);
		return {
			id: newTag.id,
			name: newTag.name,
		};
	};

	const handleTagSelect = (tag: Tag) => {
		if (!selectedTags.find((t) => t.id === tag.id)) {
			setSelectedTags([...selectedTags, tag]);
		}
	};

	const handleTagRemove = (tagId: string) => {
		setSelectedTags(selectedTags.filter((t) => t.id !== tagId));
	};

	const handleSaveTags = async () => {
		const currentTagIds = new Set(url.tags.map((t) => t.id));
		const newTagIds = new Set(selectedTags.map((t) => t.id));

		const tagsToAdd = selectedTags.filter((tag) => !currentTagIds.has(tag.id));
		const tagsToRemove = url.tags.filter((tag) => !newTagIds.has(tag.id));

		try {
			if (tagsToAdd.length > 0) {
				await addTags.mutateAsync({
					linkId: url.id,
					tagIds: tagsToAdd.map((t) => t.id),
				});
			}

			if (tagsToRemove.length > 0) {
				await removeTags.mutateAsync({
					linkId: url.id,
					tagIds: tagsToRemove.map((t) => t.id),
				});
			}

			setIsEditingTags(false);
		} catch (error) {
			// Error is handled by the hooks
		}
	};

	const handleCancelTags = () => {
		setSelectedTags(url.tags || []);
		setIsEditingTags(false);
	};

	// Check if tags have changed
	const tagsHaveChanged = () => {
		const currentTagIds = new Set((url.tags || []).map((t) => t.id));
		const newTagIds = new Set(selectedTags.map((t) => t.id));

		if (currentTagIds.size !== newTagIds.size) return true;

		for (const tagId of currentTagIds) {
			if (!newTagIds.has(tagId)) return true;
		}

		return false;
	};

	// Check if expiration has changed
	const expirationHasChanged = () => {
		if (!expirationDate && !url.expiresAt) return false;
		if (!expirationDate && url.expiresAt) return true;
		if (expirationDate && !url.expiresAt) return true;

		if (expirationDate && url.expiresAt) {
			const [hours, minutes] = expirationTime.split(":").map(Number);
			const combinedDate = new Date(expirationDate);
			combinedDate.setUTCHours(hours, minutes, 0, 0);
			const currentExpiration = new Date(url.expiresAt);
			return combinedDate.getTime() !== currentExpiration.getTime();
		}

		return false;
	};

	// Block navigation when editing
	const hasUnsavedChanges =
		(isEditingExpiration && expirationHasChanged()) ||
		(isEditingTags && tagsHaveChanged());

	useBlockNavigation({
		shouldBlock: hasUnsavedChanges,
		title: "Unsaved Changes",
		message:
			"You have unsaved changes. Are you sure you want to leave? Your changes will be lost.",
		confirmButtonLabel: "Leave",
		cancelButtonLabel: "Stay",
	});

	// ESC key handler for expiration editing
	useEffect(() => {
		if (!isEditingExpiration) return;

		const handleEscape = (e: KeyboardEvent) => {
			if (e.key === "Escape" && !updateLink.isPending) {
				handleCancelExpiration();
			}
		};

		window.addEventListener("keydown", handleEscape);
		return () => window.removeEventListener("keydown", handleEscape);
	}, [isEditingExpiration, updateLink.isPending, handleCancelExpiration]);

	const isPendingTags =
		addTags.isPending || removeTags.isPending || createTag.isPending;

	// ESC key handler for tags editing
	useEffect(() => {
		if (!isEditingTags) return;

		const handleEscape = (e: KeyboardEvent) => {
			if (e.key === "Escape" && !isPendingTags) {
				handleCancelTags();
			}
		};

		window.addEventListener("keydown", handleEscape);
		return () => window.removeEventListener("keydown", handleEscape);
	}, [isEditingTags, isPendingTags, handleCancelTags]);

	return (
		<Card>
			<CardContent className='space-y-0'>
				{/* Destination Section */}
				<div className='pb-6'>
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

				<Separator />

				{/* Expiration Section */}
				<div className='py-6'>
					<div className='flex items-center justify-between mb-3'>
						<div className='flex items-center gap-2'>
							<span className='text-sm font-semibold uppercase tracking-wider text-muted-foreground'>
								Expiration
							</span>
						</div>
						{!isEditingExpiration && (
							<Tooltip>
								<TooltipTrigger asChild>
									<Button
										variant='ghost'
										size='icon'
										onClick={() => setIsEditingExpiration(true)}
										className='h-8 w-8'
									>
										<Edit2 className='w-4 h-4' />
									</Button>
								</TooltipTrigger>
								<TooltipContent>
									<p>Edit expiration date</p>
								</TooltipContent>
							</Tooltip>
						)}
					</div>
					{isEditingExpiration ? (
						<div className='space-y-3'>
							<div className='flex gap-2'>
								<div className='flex-1'>
									<Popover>
										<PopoverTrigger asChild>
											<Button
												variant='outline'
												className={`w-full justify-between font-normal bg-input dark:bg-input/30 ${
													!expirationDate
														? "text-muted-foreground"
														: "text-foreground"
												}`}
											>
												<div className='flex items-center'>
													<CalendarIcon className='mr-2 h-4 w-4' />
													{expirationDate ? (
														formatDate(expirationDate)
													) : (
														<span>Pick a date</span>
													)}
												</div>
												<ChevronDownIcon className='h-4 w-4 opacity-50' />
											</Button>
										</PopoverTrigger>
										<PopoverContent className='w-auto p-0' align='start'>
											<Calendar
												mode='single'
												selected={expirationDate}
												onSelect={(date) => {
													setExpirationDate(date);
												}}
												disabled={(date) => {
													const today = new Date();
													today.setHours(0, 0, 0, 0);
													return date < today;
												}}
												initialFocus
											/>
										</PopoverContent>
									</Popover>
								</div>
								<div className='w-32'>
									<Input
										type='time'
										value={expirationTime}
										onChange={(e) => setExpirationTime(e.target.value)}
										step='60'
										className='cursor-text appearance-none [&::-webkit-calendar-picker-indicator]:hidden [&::-webkit-calendar-picker-indicator]:appearance-none'
									/>
								</div>
							</div>
							<div className='flex gap-2'>
								<Button
									variant='outline'
									size='sm'
									onClick={handleCancelExpiration}
									disabled={updateLink.isPending}
								>
									Cancel
								</Button>
								<Button
									size='sm'
									onClick={handleSaveExpiration}
									disabled={updateLink.isPending || !expirationDate}
								>
									{updateLink.isPending ? (
										<Spinner className='w-4 h-4 mr-1' />
									) : null}
									{updateLink.isPending ? "Saving" : "Save"}
								</Button>
							</div>
						</div>
					) : (
						<p
							className={`font-medium ${
								url.expiresAt && new Date(url.expiresAt) < new Date()
									? "text-destructive"
									: "text-muted-foreground"
							}`}
						>
							{url.expiresAt
								? formatDateTime(url.expiresAt)
								: "No expiration date set"}
						</p>
					)}
				</div>

				<Separator />

				{/* Tags Section */}
				<div className='pt-6'>
					<div className='flex items-center justify-between mb-3'>
						<div className='flex items-center gap-2'>
							<span className='text-sm font-semibold uppercase tracking-wider text-muted-foreground'>
								Tags
							</span>
						</div>
						{!isEditingTags && (
							<Tooltip>
								<TooltipTrigger asChild>
									<Button
										variant='ghost'
										size='icon'
										onClick={() => setIsEditingTags(true)}
										className='h-8 w-8'
									>
										<Edit2 className='w-4 h-4' />
									</Button>
								</TooltipTrigger>
								<TooltipContent>
									<p>Edit tags</p>
								</TooltipContent>
							</Tooltip>
						)}
					</div>
					{isEditingTags ? (
						<div className='space-y-3'>
							<div className='flex flex-wrap gap-2'>
								{selectedTags.map((tag) => (
									<Badge key={tag.id} variant='outline' className='pr-1'>
										{tag.name}
										<button
											type='button'
											onClick={() => handleTagRemove(tag.id)}
											className='ml-0.5 hover:bg-muted rounded-sm p-0.5 transition-colors'
											disabled={isPendingTags}
										>
											<X className='w-3 h-3' />
										</button>
									</Badge>
								))}
							</div>
							{isLoadingTags ? (
								<div className='text-sm text-muted-foreground'>
									Loading tags...
								</div>
							) : (
								<TagCombobox
									availableTags={availableTagsForUI}
									selectedTags={selectedTags}
									onTagSelect={handleTagSelect}
									onCreateTag={handleCreateTag}
									placeholder='Add a tag...'
									className='w-full'
								/>
							)}
							<div className='flex gap-2'>
								<Button
									variant='outline'
									size='sm'
									onClick={handleCancelTags}
									disabled={isPendingTags}
								>
									Cancel
								</Button>
								<Button
									size='sm'
									onClick={handleSaveTags}
									disabled={isPendingTags || !tagsHaveChanged()}
								>
									{isPendingTags ? <Spinner className='w-4 h-4 mr-1' /> : null}
									{isPendingTags ? "Saving" : "Save"}
								</Button>
							</div>
						</div>
					) : (
						<div className='flex flex-wrap gap-2'>
							{url.tags && url.tags.length > 0 ? (
								url.tags.map((tag) => (
									<Badge key={tag.id} variant='outline'>
										{tag.name}
									</Badge>
								))
							) : (
								<p className='text-sm text-muted-foreground'>No tags</p>
							)}
						</div>
					)}
				</div>
			</CardContent>
		</Card>
	);
}
