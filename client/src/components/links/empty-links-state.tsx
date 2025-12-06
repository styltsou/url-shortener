import { Link as LinkIcon } from "lucide-react";

export function EmptyLinksState() {
	return (
		<div className='text-center py-20 bg-card rounded-3xl border border-border'>
			<div className='w-16 h-16 bg-muted rounded-full flex items-center justify-center mx-auto mb-4'>
				<LinkIcon className='w-8 h-8 text-muted-foreground' />
			</div>
			<h3 className='text-foreground font-medium mb-1'>No links yet</h3>
			<p className='text-muted-foreground'>
				Create your first shortened link above.
			</p>
		</div>
	);
}

