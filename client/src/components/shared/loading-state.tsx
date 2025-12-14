/**
 * Shared loading state component
 * Provides consistent loading UI across the application
 */

import { Spinner } from "@/components/ui/spinner";

interface LoadingStateProps {
	message?: string;
}

export function LoadingState({ message = "Loading..." }: LoadingStateProps) {
	return (
		<div className='flex min-h-screen items-center justify-center'>
			<div className='flex flex-col items-center gap-3'>
				<Spinner className='size-6' />
				<div className='text-lg text-muted-foreground'>{message}</div>
			</div>
		</div>
	);
}
