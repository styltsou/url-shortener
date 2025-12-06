interface LinksErrorStateProps {
	error: Error;
}

export function LinksErrorState({ error }: LinksErrorStateProps) {
	return (
		<main className='py-12 px-4 sm:px-6'>
			<div className='max-w-4xl mx-auto text-center py-20'>
				<h2 className='text-2xl font-bold mb-4 text-destructive'>
					Error loading links
				</h2>
				<p className='text-muted-foreground'>{error.message}</p>
			</div>
		</main>
	);
}

