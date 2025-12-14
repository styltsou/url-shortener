import { useState } from "react";
import { Hash, Edit2, Save, X } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Spinner } from "@/components/ui/spinner";
import { useUpdateLink } from "@/hooks/use-links";
import { useNavigate } from "@tanstack/react-router";
import { toast } from "sonner";
import type { Url } from "@/types/url";
import { SHORT_DOMAIN } from "@/lib/constants";

interface ShortcodeSectionProps {
	url: Url;
}

export function ShortcodeSection({ url }: ShortcodeSectionProps) {
	const [isEditing, setIsEditing] = useState(false);
	const [shortcode, setShortcode] = useState(url.shortCode);
	const updateLink = useUpdateLink();
	const navigate = useNavigate();

	const handleSave = async () => {
		if (shortcode.trim() === url.shortCode) {
			setIsEditing(false);
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
			setIsEditing(false);
			// Navigate to the new shortcode URL if we're on the detail page
			const currentPath = window.location.pathname;
			if (currentPath === `/links/${url.shortCode}`) {
				navigate({ to: `/links/${shortcode.trim()}`, replace: true });
			}
		} catch (error) {
			// Error is handled by the hook
		}
	};

	const handleCancel = () => {
		setShortcode(url.shortCode);
		setIsEditing(false);
	};

	return (
		<Card>
			<CardHeader>
				<CardTitle className='text-sm font-semibold uppercase tracking-wider flex items-center gap-2 text-muted-foreground'>
					<Hash className='w-4 h-4' /> Shortcode
				</CardTitle>
			</CardHeader>
			<CardContent>
				{isEditing ? (
					<div className='space-y-3'>
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
						<div className='flex gap-2'>
							<Button
								variant='secondary'
								size='sm'
								onClick={handleCancel}
								disabled={updateLink.isPending}
							>
								<X className='w-4 h-4 mr-1' />
								Cancel
							</Button>
							<Button
								size='sm'
								onClick={handleSave}
								disabled={updateLink.isPending}
							>
								{updateLink.isPending ? (
									<Spinner className='w-4 h-4 mr-1' />
								) : (
									<Save className='w-4 h-4 mr-1' />
								)}
								{updateLink.isPending ? "Saving" : "Save"}
							</Button>
						</div>
					</div>
				) : (
					<div className='flex items-center justify-between'>
						<div className='flex items-center gap-2'>
							<span className='text-sm text-muted-foreground'>
								{SHORT_DOMAIN}/
							</span>
							<span className='font-medium'>{url.shortCode}</span>
						</div>
						<Button
							variant='ghost'
							size='sm'
							onClick={() => setIsEditing(true)}
						>
							<Edit2 className='w-4 h-4 mr-1' />
							Edit
						</Button>
					</div>
				)}
			</CardContent>
		</Card>
	);
}
