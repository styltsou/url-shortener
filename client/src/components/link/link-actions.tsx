import { useState } from "react";
import {
	Edit,
	Save,
	Trash2,
	Loader2,
	Power,
	PowerOff,
	MoreVertical as MoreVerticalIcon,
} from "lucide-react";
import { Button } from "@/components/ui/button";
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
import { useUpdateLink, useDeleteLink } from "@/hooks/use-links";
import { useNavigate } from "@tanstack/react-router";
import { toast } from "sonner";
import type { Url } from "@/types/url";

interface LinkActionsProps {
	url: Url;
	isEditing: boolean;
	onEdit: () => void;
	onCancel: () => void;
	onSave: () => void;
}

export function LinkActions({
	url,
	isEditing,
	onEdit,
	onCancel,
	onSave,
}: LinkActionsProps) {
	const navigate = useNavigate();
	const updateLink = useUpdateLink();
	const deleteLink = useDeleteLink();
	const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
	const [deactivateDialogOpen, setDeactivateDialogOpen] = useState(false);
	const [isActive, setIsActive] = useState(url.isActive !== false);

	const handleDelete = async () => {
		await deleteLink.mutateAsync(url.id);
		navigate({ to: "/" });
	};

	const handleToggleActive = async () => {
		// TODO: Call API to update link active status
		setIsActive(!isActive);
		toast.success(isActive ? "Link deactivated" : "Link activated");
		setDeactivateDialogOpen(false);
	};
	if (isEditing) {
		return (
			<div className='flex gap-2'>
				<Button variant='secondary' onClick={onCancel}>
					Cancel
				</Button>
				<Button onClick={onSave} disabled={updateLink.isPending}>
					{updateLink.isPending ? (
						<Loader2 className='w-4 h-4 animate-spin' />
					) : (
						<Save className='w-4 h-4' />
					)}{" "}
					Save Changes
				</Button>
			</div>
		);
	}

	return (
		<div className='flex gap-2'>
			<Button variant='ghost' onClick={onEdit}>
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
									<AlertDialogTitle>Deactivate link?</AlertDialogTitle>
									<AlertDialogDescription>
										This will deactivate the short URL{" "}
										<strong>short.ly/{url.shortCode}</strong>. The link will
										stop working and redirects will fail. You can reactivate it
										at any time.
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
			<AlertDialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
				<AlertDialogContent>
					<AlertDialogHeader>
						<AlertDialogTitle>Are you absolutely sure?</AlertDialogTitle>
						<AlertDialogDescription>
							This action cannot be undone. This will permanently delete the
							short URL <strong>short.ly/{url.shortCode}</strong> and all its
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
		</div>
	);
}

