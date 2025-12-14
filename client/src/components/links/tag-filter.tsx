import { useState, useEffect } from "react";
import { Filter, X, ChevronDown } from "lucide-react";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
	Command,
	CommandEmpty,
	CommandGroup,
	CommandInput,
	CommandItem,
	CommandList,
} from "@/components/ui/command";
import { cn } from "@/lib/utils";
import type { Tag } from "@/types/url";

interface TagFilterProps {
	availableTags: Tag[];
	selectedTagIds: string[];
	onSelectionChange: (tagIds: string[]) => void;
}

export function TagFilter({
	availableTags,
	selectedTagIds,
	onSelectionChange,
}: TagFilterProps) {
	const [open, setOpen] = useState(false);
	const [searchValue, setSearchValue] = useState("");

	// Reset search when popover closes
	useEffect(() => {
		if (!open) {
			setSearchValue("");
		}
	}, [open]);

	const selectedTags = availableTags.filter((tag) =>
		selectedTagIds.includes(tag.id)
	);

	const availableTagsForSelection = availableTags.filter(
		(tag) => !selectedTagIds.includes(tag.id)
	);

	const filteredAvailableTags = availableTagsForSelection.filter((tag) => {
		if (!searchValue.trim()) return true;
		return tag.name.toLowerCase().includes(searchValue.toLowerCase());
	});

	const handleTagToggle = (tagId: string) => {
		if (selectedTagIds.includes(tagId)) {
			onSelectionChange(selectedTagIds.filter((id) => id !== tagId));
		} else {
			onSelectionChange([...selectedTagIds, tagId]);
		}
	};

	const handleRemoveTag = (tagId: string, e: React.MouseEvent) => {
		e.stopPropagation();
		onSelectionChange(selectedTagIds.filter((id) => id !== tagId));
	};

	const handleClearAll = () => {
		onSelectionChange([]);
	};

	return (
		<div className="flex items-center gap-2 flex-wrap">
			{/* Filter button with popover */}
			<Popover open={open} onOpenChange={setOpen}>
				<PopoverTrigger asChild>
					<Button
						variant="outline"
						size="sm"
						className={cn(
							"h-8 gap-2",
							selectedTagIds.length > 0 && "bg-accent"
						)}
					>
						<Filter className="h-3.5 w-3.5" />
						<span className="text-xs">Filter by tags</span>
						{selectedTagIds.length > 0 && (
							<Badge
								variant="secondary"
								className="ml-1 h-4 px-1.5 text-[10px] font-semibold"
							>
								{selectedTagIds.length}
							</Badge>
						)}
						<ChevronDown className="h-3.5 w-3.5 opacity-50" />
					</Button>
				</PopoverTrigger>
				<PopoverContent className="w-[280px] p-0" align="start">
					<Command shouldFilter={false}>
						<CommandInput
							placeholder="Search tags..."
							value={searchValue}
							onValueChange={setSearchValue}
						/>
						<CommandList className="[&::-webkit-scrollbar]:w-2 [&::-webkit-scrollbar-track]:bg-transparent [&::-webkit-scrollbar-thumb]:bg-border [&::-webkit-scrollbar-thumb]:rounded-full [&::-webkit-scrollbar-thumb]:border-2 [&::-webkit-scrollbar-thumb]:border-transparent [&::-webkit-scrollbar-thumb]:bg-clip-padding hover:[&::-webkit-scrollbar-thumb]:bg-muted-foreground/30 [&::-webkit-scrollbar-thumb]:transition-colors">
							<CommandEmpty>
								{searchValue.trim()
									? "No tags found"
									: "No tags available"}
							</CommandEmpty>
							{filteredAvailableTags.length > 0 && (
								<CommandGroup>
									{filteredAvailableTags.map((tag) => (
										<CommandItem
											key={tag.id}
											value={tag.name}
											onSelect={() => handleTagToggle(tag.id)}
											className="cursor-pointer"
										>
											{tag.name}
										</CommandItem>
									))}
								</CommandGroup>
							)}
						</CommandList>
					</Command>
				</PopoverContent>
			</Popover>

			{/* Selected tags as removable badges */}
			{selectedTags.length > 0 && (
				<>
					{selectedTags.map((tag) => (
						<Badge
							key={tag.id}
							variant="outline"
							className="pr-1 gap-1.5"
						>
							{tag.name}
							<button
								type="button"
								onClick={(e) => handleRemoveTag(tag.id, e)}
								className="ml-0.5 hover:bg-muted rounded-sm p-0.5 transition-colors"
							>
								<X className="w-3 h-3" />
							</button>
						</Badge>
					))}
					<Button
						variant="ghost"
						size="sm"
						onClick={handleClearAll}
						className="h-6 px-2 text-xs"
					>
						Clear all
					</Button>
				</>
			)}
		</div>
	);
}

