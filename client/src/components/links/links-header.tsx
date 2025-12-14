interface LinksHeaderProps {
	isLoading: boolean;
	totalCount: number;
	hasActiveFilters?: boolean;
}

export function LinksHeader({
	isLoading,
	totalCount,
	hasActiveFilters,
}: LinksHeaderProps) {
	return (
		<div className='flex items-end justify-between mb-4 pb-2 border-b border-border'>
			<h2 className='text-lg font-bold text-foreground tracking-tight'>
				Links
			</h2>
			<span className='text-xs font-medium text-muted-foreground uppercase tracking-wider'>
				{isLoading ? (
					"Loading..."
				) : hasActiveFilters ? (
					`${totalCount} results`
				) : (
					`${totalCount} Total`
				)}
			</span>
		</div>
	);
}

