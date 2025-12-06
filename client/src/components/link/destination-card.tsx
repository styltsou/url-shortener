import React, { useState, useImperativeHandle, forwardRef } from "react";
import { Globe, Copy, CopyCheck, ExternalLink } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { toast } from "sonner";
import type { Url } from "@/types/url";
import { TagsSection } from "./tags-section";
import { ExpirationSection } from "./expiration-section";

interface DestinationCardProps {
	url: Url;
	isEditing: boolean;
}

export interface DestinationCardRef {
	getFormData: () => { expirationDate?: Date };
}

export const DestinationCard = forwardRef<DestinationCardRef, DestinationCardProps>(
	({ url, isEditing }, ref) => {
		const [originalUrlInput, setOriginalUrlInput] = useState(url.originalUrl);
		const [destinationCopied, setDestinationCopied] = useState(false);
		const expirationRef = React.useRef<{ getExpirationDate: () => Date | undefined }>(null);

		useImperativeHandle(ref, () => ({
			getFormData: () => ({
				expirationDate: expirationRef.current?.getExpirationDate(),
			}),
		}));

		const handleCopyDestination = (e: React.MouseEvent) => {
			e.preventDefault();
			e.stopPropagation();
			navigator.clipboard.writeText(url.originalUrl);
			setDestinationCopied(true);
			toast.success("Destination URL copied to clipboard");
			setTimeout(() => setDestinationCopied(false), 2000);
		};

	return (
		<Card>
			<CardHeader>
				<CardTitle className='text-sm font-semibold uppercase tracking-wider flex items-center gap-2 text-muted-foreground'>
					<Globe className='w-4 h-4' /> Destination
				</CardTitle>
			</CardHeader>
			<CardContent>
				{isEditing ? (
					<Input
						type='url'
						value={originalUrlInput}
						onChange={(e) => setOriginalUrlInput(e.target.value)}
					/>
				) : (
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
							<ExternalLink className='w-4 h-4 text-muted-foreground group-hover:text-primary transition-colors' />
						</div>
					</div>
				)}

				<TagsSection url={url} isEditing={isEditing} />
				<ExpirationSection ref={expirationRef} url={url} isEditing={isEditing} />
			</CardContent>
		</Card>
		);
	}
);

