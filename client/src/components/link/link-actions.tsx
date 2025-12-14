import { useState, useEffect } from "react";
import {
	Trash2,
	Power,
	PowerOff,
	MoreVertical as MoreVerticalIcon,
	Pencil,
	Edit2,
	Save,
	X,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Spinner } from "@/components/ui/spinner";
import {
	Tooltip,
	TooltipContent,
	TooltipTrigger,
} from "@/components/ui/tooltip";
import {
	DropdownMenu,
	DropdownMenuContent,
	DropdownMenuItem,
	DropdownMenuSeparator,
	DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
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
import {
	Dialog,
	DialogContent,
	DialogDescription,
	DialogFooter,
	DialogHeader,
	DialogTitle,
} from "@/components/ui/dialog";
import { useUpdateLink, useDeleteLink } from "@/hooks/use-links";
import { useNavigate } from "@tanstack/react-router";
import { toast } from "sonner";
import { useBlockNavigation } from "@/hooks/use-block-navigation";
import type { Url } from "@/types/url";
import { SHORT_DOMAIN } from "@/lib/constants";

interface LinkActionsProps {
	url: Url;
}

export function LinkActions({ url }: LinkActionsProps) {
	const navigate = useNavigate();
	const updateLink = useUpdateLink();
	const deleteLink = useDeleteLink();
	const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
	const [deactivateDialogOpen, setDeactivateDialogOpen] = useState(false);
	const [shortcodeDialogOpen, setShortcodeDialogOpen] = useState(false);
	const [shortcode, setShortcode] = useState(url.shortCode);
	const isExpired = url.expiresAt && new Date(url.expiresAt) < new Date();
	const isManuallyInactive = url.isActive === false;
	// A link is effectively inactive if it's manually deactivated OR expired
	const isEffectivelyInactive = isManuallyInactive || isExpired;

	const handleDelete = async () => {
		try {
			// Wait for the mutation to complete - this shows loading state
			await deleteLink.mutateAsync(url.id);
			// Close dialog only after successful deletion
			setDeleteDialogOpen(false);
			// Navigate only after successful deletion
			navigate({ to: "/links", replace: true });
		} catch (error) {
			// Error is handled by the hook, keep dialog open so user can see the error
		}
	};

	const handleToggleActive = async () => {
		try {
			await updateLink.mutateAsync({
				id: url.id,
				data: {
					is_active: !isManuallyInactive,
				},
			});
			setDeactivateDialogOpen(false);
		} catch (error) {
			// Error is handled by the hook
		}
	};

	const handleSaveShortcode = async () => {
		if (shortcode.trim() === url.shortCode) {
			setShortcodeDialogOpen(false);
			return;
		}

		if (shortcode.trim().length === 0) {
			toast.error("Shortcode cannot be empty");
			return;
		}

		if (shortcode.trim().length > 20) {
			toast.error("Shortcode must be 20 characters or less");
			return;
		}

		try {
			await updateLink.mutateAsync({
				id: url.id,
				data: {
					shortcode: shortcode.trim(),
				},
			});
			setShortcodeDialogOpen(false);
			// Navigate to the new shortcode URL if we're on the detail page
			const currentPath = window.location.pathname;
			if (currentPath === `/links/${url.shortCode}`) {
				navigate({ to: `/links/${shortcode.trim()}`, replace: true });
			}
		} catch (error) {
			// Error is handled by the hook
		}
	};

	const handleCancelShortcode = () => {
		setShortcode(url.shortCode);
		setShortcodeDialogOpen(false);
	};

	// Block navigation when shortcode has unsaved changes
	const shortcodeHasChanged = shortcode.trim() !== url.shortCode;
	useBlockNavigation({
		shouldBlock: shortcodeDialogOpen && shortcodeHasChanged,
		title: "Unsaved Changes",
		message:
			"You have unsaved changes to the shortcode. Are you sure you want to leave?",
		confirmButtonLabel: "Leave",
		cancelButtonLabel: "Stay",
	});

	// Reset shortcode when dialog opens
	useEffect(() => {
		if (shortcodeDialogOpen) {
			setShortcode(url.shortCode);
		}
	}, [shortcodeDialogOpen, url.shortCode]);

	return (
		<div className='flex gap-2'>
			<DropdownMenu>
				<Tooltip>
					<TooltipTrigger asChild>
						<DropdownMenuTrigger asChild>
							<Button variant='secondary'>
								<MoreVerticalIcon className='w-4 h-4' />
							</Button>
						</DropdownMenuTrigger>
					</TooltipTrigger>
					<TooltipContent>
						<p>More options</p>
					</TooltipContent>
				</Tooltip>
				<DropdownMenuContent align='end'>
					<DropdownMenuItem
						onSelect={(e) => {
							e.preventDefault();
							setShortcodeDialogOpen(true);
						}}
					>
						<Pencil className='w-4 h-4 mr-2' />
						Change shortcode
					</DropdownMenuItem>
					<DropdownMenuSeparator />
					{!isEffectivelyInactive ? (
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
									<AlertDialogTitle>Deactivate link?</AlertDialogTitle>
									<AlertDialogDescription>
										This will deactivate the short URL{" "}
										<strong>
											{SHORT_DOMAIN}/{url.shortCode}
										</strong>
										. The link will stop working and redirects will fail. You
										can reactivate it at any time.
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
					) : isExpired ? (
						<DropdownMenuItem disabled>
							<PowerOff className='w-4 h-4 mr-2' />
							Expired (cannot activate)
						</DropdownMenuItem>
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
				onOpenChange={(open) => {
					// Prevent closing the dialog while deletion is in progress
					if (!open && deleteLink.isPending) {
						return;
					}
					setDeleteDialogOpen(open);
				}}
			>
				<AlertDialogContent>
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
						<AlertDialogCancel disabled={deleteLink.isPending}>
							Cancel
						</AlertDialogCancel>
						<AlertDialogAction
							onClick={handleDelete}
							disabled={deleteLink.isPending}
							className='bg-destructive text-white hover:bg-destructive/90 focus-visible:ring-destructive/20 dark:focus-visible:ring-destructive/40 dark:bg-destructive/60'
						>
							{deleteLink.isPending ? (
								<>
									<Spinner className='w-4 h-4 mr-2' />
									Deleting
								</>
							) : (
								"Delete"
							)}
						</AlertDialogAction>
					</AlertDialogFooter>
				</AlertDialogContent>
			</AlertDialog>
			<Dialog open={shortcodeDialogOpen} onOpenChange={setShortcodeDialogOpen}>
				<DialogContent>
					<DialogHeader>
						<DialogTitle>Change shortcode</DialogTitle>
						<DialogDescription>
							Change the shortcode for this link. The new shortcode must be
							unique and 20 characters or less.
						</DialogDescription>
					</DialogHeader>
					<form
						onSubmit={(e) => {
							e.preventDefault();
							handleSaveShortcode();
						}}
					>
						<div className='space-y-4 py-4'>
							<div className='flex items-center gap-2'>
								<span className='text-sm text-muted-foreground'>
									{SHORT_DOMAIN}/
								</span>
								<Input
									value={shortcode}
									onChange={(e) => setShortcode(e.target.value)}
									placeholder='Enter shortcode'
									maxLength={20}
									className='flex-1'
								/>
							</div>
						</div>
						<DialogFooter>
							<Button
								type='button'
								variant='outline'
								onClick={handleCancelShortcode}
								disabled={updateLink.isPending}
							>
								Cancel
							</Button>
							<Button type='submit' disabled={updateLink.isPending}>
								{updateLink.isPending ? (
									<Spinner className='w-4 h-4 mr-1' />
								) : null}
								{updateLink.isPending ? "Saving" : "Save"}
							</Button>
						</DialogFooter>
					</form>
				</DialogContent>
			</Dialog>
		</div>
	);
}
