import { useState, useEffect } from "react";
import {
	Clock,
	Calendar as CalendarIcon,
	Save,
	X,
	ChevronDownIcon,
} from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Spinner } from "@/components/ui/spinner";
import {
	Popover,
	PopoverContent,
	PopoverTrigger,
} from "@/components/ui/popover";
import { Calendar } from "@/components/ui/calendar";
import { Input } from "@/components/ui/input";
import { formatDate, formatDateTime } from "@/lib/mock-data";
import { useUpdateLink } from "@/hooks/use-links";
import { toast } from "sonner";
import type { Url } from "@/types/url";

interface ExpirationSectionProps {
	url: Url;
}

export function ExpirationSection({ url }: ExpirationSectionProps) {
	const [isEditing, setIsEditing] = useState(false);
	const [expirationDate, setExpirationDate] = useState<Date | undefined>(
		url.expiresAt ? new Date(url.expiresAt) : undefined
	);
	const [expirationTime, setExpirationTime] = useState<string>(() => {
		if (url.expiresAt) {
			const date = new Date(url.expiresAt);
			// Format as HH:MM for time input (24-hour format)
			const hours = date.getUTCHours().toString().padStart(2, "0");
			const minutes = date.getUTCMinutes().toString().padStart(2, "0");
			return `${hours}:${minutes}`;
		}
		return "23:59";
	});
	const updateLink = useUpdateLink();

	// Initialize expiration date and time when entering edit mode
	useEffect(() => {
		if (isEditing) {
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
	}, [isEditing, url.expiresAt]);

	const handleSave = async () => {
		if (!expirationDate) {
			setIsEditing(false);
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
			setIsEditing(false);
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
			setIsEditing(false);
		} catch (error) {
			// Error is handled by the hook
		}
	};

	const handleCancel = () => {
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
		setIsEditing(false);
	};

	return (
		<Card>
			<CardHeader>
				<CardTitle className='text-sm font-semibold uppercase tracking-wider flex items-center gap-2 text-muted-foreground'>
					<Clock className='w-4 h-4' /> Expiration
				</CardTitle>
			</CardHeader>
			<CardContent>
				{isEditing ? (
					<div className='space-y-3'>
						<div className='flex gap-2'>
							<div className='flex-1'>
								<Popover>
									<PopoverTrigger asChild>
										<Button
											variant='outline'
											className={`w-full justify-between font-normal bg-input ${
												!expirationDate ? "text-muted-foreground" : ""
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
												// Disable past dates, but allow today if time is in the future
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
								disabled={updateLink.isPending || !expirationDate}
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
						<Button
							variant='ghost'
							size='sm'
							onClick={() => setIsEditing(true)}
						>
							<CalendarIcon className='w-4 h-4 mr-1' />
							Edit
						</Button>
					</div>
				)}
			</CardContent>
		</Card>
	);
}
