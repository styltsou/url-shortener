import {
	Link as LinkIcon,
	ArrowRight,
	ChevronRight,
	Calendar as CalendarIcon,
	ChevronDownIcon,
	X,
} from "lucide-react";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Spinner } from "@/components/ui/spinner";
import {
	InputGroup,
	InputGroupText,
	InputGroupInput,
} from "@/components/ui/input-group";
import {
	Popover,
	PopoverContent,
	PopoverTrigger,
} from "@/components/ui/popover";
import { Calendar } from "@/components/ui/calendar";
import { formatDate } from "@/lib/mock-data";
import { useState } from "react";
import { SHORT_DOMAIN } from "@/lib/constants";
import { TagCombobox } from "@/components/link/tag-combobox";
import { useTags, useCreateTag } from "@/hooks/use-tags";
import { toast } from "sonner";
import type { Tag } from "@/types/url";

interface NewUrlFormProps {
	onShorten: (
		originalUrl: string,
		customCode?: string,
		expirationDate?: string,
		tagIds?: string[]
	) => Promise<void>;
	isLoading: boolean;
}

interface FormState {
	originalUrl: string;
	customCode: string;
	expirationDate: Date | undefined;
	expirationTime: string;
	selectedTags: Tag[];
	showOptions: boolean;
}

export function NewUrlForm({ onShorten, isLoading }: NewUrlFormProps) {
	const [formState, setFormState] = useState<FormState>({
		originalUrl: "",
		customCode: "",
		expirationDate: undefined,
		expirationTime: "23:59",
		selectedTags: [],
		showOptions: false,
	});

	const { data: availableTags = [], isLoading: isLoadingTags } = useTags();
	const createTag = useCreateTag();

	// Convert API tags to UI tags
	const availableTagsForUI = availableTags.map((tag) => ({
		id: tag.id,
		name: tag.name,
	}));

	const handleSubmit = async (e: React.FormEvent) => {
		e.preventDefault();
		if (!formState.originalUrl) return;

		// Combine date and time if expiration date is set
		let expirationDateString: string | undefined = undefined;
		if (formState.expirationDate) {
			const [hours, minutes] = formState.expirationTime.split(":").map(Number);
			const combinedDate = new Date(formState.expirationDate);
			combinedDate.setUTCHours(hours, minutes, 0, 0);

			// Validate that expiration is in the future
			// Use < instead of <= to allow dates that are at least the current time
			// This accounts for network latency and ensures users can set expiration times
			// that are very close to the current time (e.g., 2 hours from now)
			if (combinedDate < new Date()) {
				toast.error("Expiration date and time must be in the future");
				return;
			}

			expirationDateString = combinedDate.toISOString();
		}

		const tagIds =
			formState.selectedTags.length > 0
				? formState.selectedTags.map((tag) => tag.id)
				: undefined;

		await onShorten(
			formState.originalUrl,
			formState.customCode || undefined,
			expirationDateString,
			tagIds
		);

		// Reset form after successful submission
		setFormState({
			originalUrl: "",
			customCode: "",
			expirationDate: undefined,
			expirationTime: "23:59",
			selectedTags: [],
			showOptions: false,
		});
	};

	const handleCreateTag = async (tagName: string): Promise<Tag> => {
		const newTag = await createTag.mutateAsync(tagName);
		return {
			id: newTag.id,
			name: newTag.name,
		};
	};

	const handleTagSelect = (tag: Tag) => {
		if (!formState.selectedTags.find((t) => t.id === tag.id)) {
			setFormState({
				...formState,
				selectedTags: [...formState.selectedTags, tag],
			});
		}
	};

	const handleTagRemove = (tagId: string) => {
		setFormState({
			...formState,
			selectedTags: formState.selectedTags.filter((t) => t.id !== tagId),
		});
	};

	return (
		<div className='w-full max-w-3xl mx-auto mb-12'>
			<div className='text-center mb-8 space-y-2'>
				<h2 className='text-4xl font-extrabold text-foreground tracking-tight'>
					Shorten your links
				</h2>
				<p className='text-muted-foreground text-lg'>
					Detailed analytics and custom branding included.
				</p>
			</div>

			<form onSubmit={handleSubmit} className='w-full'>
				<div className='flex flex-col sm:flex-row gap-3'>
					<div className='relative flex-1'>
						<div className='absolute left-4 top-1/2 -translate-y-1/2 text-muted-foreground'>
							<LinkIcon className='w-5 h-5' />
						</div>
						<Input
							type='url'
							value={formState.originalUrl}
							onChange={(e) =>
								setFormState({ ...formState, originalUrl: e.target.value })
							}
							placeholder='Paste a long URL here...'
							required
							className='h-12 pl-12 text-base md:text-lg w-full'
						/>
					</div>
					<Button
						type='submit'
						disabled={isLoading}
						className='h-12 px-8 text-base font-semibold shrink-0 transition-colors'
					>
						{isLoading && <Spinner className='w-5 h-5 mr-2' />}
						{isLoading ? "Shortening" : "Shorten"}{" "}
						{!isLoading && <ArrowRight className='w-5 h-5 ml-2' />}
					</Button>
				</div>

				<div className='mt-3 text-center'>
					<Button
						type='button'
						variant='ghost'
						onClick={() =>
							setFormState({
								...formState,
								showOptions: !formState.showOptions,
							})
						}
						className='text-sm font-medium text-muted-foreground hover:text-foreground group'
					>
						{formState.showOptions ? "Hide Options" : "Show options"}
						<ChevronRight
							className={`w-3 h-3 transition-all group-hover:text-foreground ${
								formState.showOptions ? "rotate-90" : "rotate-0"
							}`}
						/>
					</Button>
				</div>

				<div
					className={`overflow-hidden transition-all duration-300 ease-in-out ${
						formState.showOptions
							? "max-h-96 opacity-100 mt-6"
							: "max-h-0 opacity-0"
					}`}
				>
					<div className='grid grid-cols-1 md:grid-cols-2 gap-3 px-1 pb-2'>
						<div>
							<label className='block text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-2'>
								Custom Alias
							</label>
							<InputGroup>
								<InputGroupText className='text-primary'>
									{SHORT_DOMAIN}/
								</InputGroupText>
								<InputGroupInput
									type='text'
									value={formState.customCode}
									onChange={(e) =>
										setFormState({
											...formState,
											customCode: e.target.value,
										})
									}
									placeholder='alias'
								/>
							</InputGroup>
						</div>
						<div>
							<label className='block text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-2'>
								Expiration Date
							</label>
							<div className='flex gap-2'>
								<Popover>
									<PopoverTrigger asChild>
										<Button
											variant='outline'
											className={`flex-1 justify-between font-normal bg-input ${
												!formState.expirationDate ? "text-muted-foreground" : ""
											}`}
										>
											<div className='flex items-center'>
												<CalendarIcon className='mr-2 h-4 w-4' />
												{formState.expirationDate ? (
													formatDate(formState.expirationDate)
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
											selected={formState.expirationDate}
											onSelect={(date) => {
												setFormState({
													...formState,
													expirationDate: date,
												});
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
								<Input
									type='time'
									value={formState.expirationTime}
									onChange={(e) =>
										setFormState({
											...formState,
											expirationTime: e.target.value,
										})
									}
									step='60'
									className='w-32 cursor-text appearance-none [&::-webkit-calendar-picker-indicator]:hidden [&::-webkit-calendar-picker-indicator]:appearance-none'
								/>
							</div>
						</div>
						<div className='md:col-span-2'>
							<label className='block text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-2'>
								Tags
							</label>
							<div className='flex items-center gap-2 flex-wrap'>
								{!isLoadingTags && (
									<div className='w-48 shrink-0'>
										<TagCombobox
											availableTags={availableTagsForUI}
											selectedTags={formState.selectedTags}
											onTagSelect={handleTagSelect}
											onCreateTag={handleCreateTag}
											placeholder='Add a tag...'
										/>
									</div>
								)}
								{formState.selectedTags.length > 0 && (
									<>
										<div className='flex flex-wrap gap-2 flex-1 min-w-0'>
											{formState.selectedTags.map((tag) => (
												<div
													key={tag.id}
													className='inline-flex items-center gap-1.5 px-2.5 py-1 rounded-md border border-border bg-background text-sm shrink-0'
												>
													<span>{tag.name}</span>
													<button
														type='button'
														onClick={() => handleTagRemove(tag.id)}
														className='text-muted-foreground hover:text-foreground transition-colors'
													>
														×
													</button>
												</div>
											))}
										</div>
										<Button
											type='button'
											variant='ghost'
											size='sm'
											onClick={() =>
												setFormState({
													...formState,
													selectedTags: [],
												})
											}
											className='h-8 px-2 text-xs text-muted-foreground hover:text-foreground shrink-0'
										>
											<X className='w-3 h-3 mr-1' />
											Clear all
										</Button>
									</>
								)}
							</div>
						</div>
					</div>
				</div>
			</form>
		</div>
	);
}
