import { useState, useEffect } from "react";
import { Tag as TagIcon, X, Save, Edit2 } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Spinner } from "@/components/ui/spinner";
import { TagCombobox } from "./tag-combobox";
import {
	useTags,
	useCreateTag,
	useAddTagsToLink,
	useRemoveTagsFromLink,
} from "@/hooks/use-tags";
import type { Url, Tag } from "@/types/url";

interface TagsSectionProps {
	url: Url;
}

export function TagsSection({ url }: TagsSectionProps) {
	const [isEditing, setIsEditing] = useState(false);
	const [selectedTags, setSelectedTags] = useState<Tag[]>(url.tags || []);

	const { data, isLoading: isLoadingTags } = useTags();
	const availableTags = data ?? [];

	const createTag = useCreateTag();
	const addTags = useAddTagsToLink();
	const removeTags = useRemoveTagsFromLink();

	// Initialize selected tags when entering edit mode
	useEffect(() => {
		if (isEditing) {
			setSelectedTags(url.tags || []);
		}
	}, [isEditing, url.tags]);

	// Convert API tags to UI tags
	// React Compiler automatically memoizes this computation
	const availableTagsForUI = !Array.isArray(availableTags)
		? []
		: availableTags.map((tag) => ({
				id: tag.id,
				name: tag.name,
			}));

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

	const handleSave = async () => {
		const currentTagIds = new Set(url.tags.map((t) => t.id));
		const newTagIds = new Set(selectedTags.map((t) => t.id));

		// Find tags to add
		const tagsToAdd = selectedTags.filter((tag) => !currentTagIds.has(tag.id));
		// Find tags to remove
		const tagsToRemove = url.tags.filter((tag) => !newTagIds.has(tag.id));

		try {
			// Add new tags
			if (tagsToAdd.length > 0) {
				await addTags.mutateAsync({
					linkId: url.id,
					tagIds: tagsToAdd.map((t) => t.id),
				});
			}

			// Remove tags
			if (tagsToRemove.length > 0) {
				await removeTags.mutateAsync({
					linkId: url.id,
					tagIds: tagsToRemove.map((t) => t.id),
				});
			}

			// Only close if there were actual changes
			if (tagsToAdd.length > 0 || tagsToRemove.length > 0) {
				setIsEditing(false);
			} else {
				setIsEditing(false);
			}
		} catch (error) {
			// Error is handled by the hooks
		}
	};

	const handleCancel = () => {
		setSelectedTags(url.tags || []);
		setIsEditing(false);
	};

	const isPending =
		addTags.isPending || removeTags.isPending || createTag.isPending;

	return (
		<Card>
			<CardHeader>
				<CardTitle className='text-sm font-semibold uppercase tracking-wider flex items-center gap-2 text-muted-foreground'>
					<TagIcon className='w-4 h-4' /> Tags
				</CardTitle>
			</CardHeader>
			<CardContent>
				{isEditing ? (
					<div className='space-y-3'>
						<div className='flex flex-wrap gap-2'>
							{selectedTags.map((tag) => (
								<Badge
									key={tag.id}
									variant='outline'
									className='pr-1'
								>
									{tag.name}
									<button
										type='button'
										onClick={() => handleTagRemove(tag.id)}
										className='ml-0.5 hover:bg-muted rounded-sm p-0.5 transition-colors'
										disabled={isPending}
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
								variant='secondary'
								size='sm'
								onClick={handleCancel}
								disabled={isPending}
							>
								<X className='w-4 h-4 mr-1' />
								Cancel
							</Button>
							<Button size='sm' onClick={handleSave} disabled={isPending}>
								{isPending ? (
									<Spinner className='w-4 h-4 mr-1' />
								) : (
									<Save className='w-4 h-4 mr-1' />
								)}
								{isPending ? "Saving" : "Save"}
							</Button>
						</div>
					</div>
				) : (
					<div className='flex items-center justify-between'>
						<div className='flex flex-wrap gap-2 flex-1'>
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
