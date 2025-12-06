import { useState, useEffect, useImperativeHandle, forwardRef } from "react";
import { Clock, Calendar as CalendarIcon } from "lucide-react";
import { CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import {
	Popover,
	PopoverContent,
	PopoverTrigger,
} from "@/components/ui/popover";
import { Calendar } from "@/components/ui/calendar";
import { formatDate } from "@/lib/mock-data";
import type { Url } from "@/types/url";

interface ExpirationSectionProps {
	url: Url;
	isEditing: boolean;
}

export interface ExpirationSectionRef {
	getExpirationDate: () => Date | undefined;
}

export const ExpirationSection = forwardRef<ExpirationSectionRef, ExpirationSectionProps>(
	({ url, isEditing }, ref) => {
	const [expirationDate, setExpirationDate] = useState<Date | undefined>(
		url.expiresAt ? new Date(url.expiresAt) : undefined
	);

	// Initialize expiration date when entering edit mode
	useEffect(() => {
		if (isEditing) {
			setExpirationDate(url.expiresAt ? new Date(url.expiresAt) : undefined);
		}
	}, [isEditing, url.expiresAt]);

	useImperativeHandle(ref, () => ({
		getExpirationDate: () => expirationDate,
	}));

	return (
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
		);
	}
);

