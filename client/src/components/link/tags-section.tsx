import { useState, useEffect } from "react";
import { Tag as TagIcon, X } from "lucide-react";
import { CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { TagCombobox } from "./tag-combobox";
import { MOCK_TAGS } from "@/lib/mock-data";
import { toast } from "sonner";
import type { Url, Tag } from "@/types/url";

interface TagsSectionProps {
	url: Url;
	isEditing: boolean;
}

export function TagsSection({ url, isEditing }: TagsSectionProps) {
	const [selectedTags, setSelectedTags] = useState<Tag[]>(url.tags || []);
	const [availableTags, setAvailableTags] = useState<Tag[]>(MOCK_TAGS);

	// Initialize selected tags when entering edit mode
	useEffect(() => {
		if (isEditing) {
			setSelectedTags(url.tags || []);
		}
	}, [isEditing, url.tags]);

	const handleCreateTag = async (tagName: string): Promise<Tag> => {
		// Simulate API delay
		await new Promise((resolve) => setTimeout(resolve, 500));

		const newTag: Tag = {
			id: `tag-${Date.now()}-${Math.random().toString(36).substring(2, 9)}`,
			name: tagName.trim(),
		};

		setAvailableTags((prev) => {
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

	const handleTagRemove = (tagId: string) => {
		setSelectedTags(selectedTags.filter((t) => t.id !== tagId));
	};

	return (
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
									onClick={() => handleTagRemove(tag.id)}
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
							<Badge key={tag.id} variant='outline' className='text-sm'>
								{tag.name}
							</Badge>
						))
					) : (
						<p className='text-sm text-muted-foreground'>No tags</p>
					)}
				</div>
			)}
		</div>
	);
}

