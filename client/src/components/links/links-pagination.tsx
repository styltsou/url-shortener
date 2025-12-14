import {
	Pagination,
	PaginationContent,
	PaginationEllipsis,
	PaginationItem,
	PaginationLink,
	PaginationNext,
	PaginationPrevious,
} from "@/components/ui/pagination";
import { cn } from "@/lib/utils";

interface LinksPaginationProps {
	currentPage: number;
	totalPages: number;
	onPageChange: (page: number) => void;
}

export function LinksPagination({
	currentPage,
	totalPages,
	onPageChange,
}: LinksPaginationProps) {
	const getPageNumbers = () => {
		const pages: (number | "ellipsis")[] = [];
		const maxVisible = 7;

		if (totalPages <= maxVisible) {
			// Show all pages if total is less than max visible
			for (let i = 1; i <= totalPages; i++) {
				pages.push(i);
			}
		} else {
			// Always show first page
			pages.push(1);

			if (currentPage <= 3) {
				// Near the start
				for (let i = 2; i <= 4; i++) {
					pages.push(i);
				}
				pages.push("ellipsis");
				pages.push(totalPages);
			} else if (currentPage >= totalPages - 2) {
				// Near the end
				pages.push("ellipsis");
				for (let i = totalPages - 3; i <= totalPages; i++) {
					pages.push(i);
				}
			} else {
				// In the middle
				pages.push("ellipsis");
				for (let i = currentPage - 1; i <= currentPage + 1; i++) {
					pages.push(i);
				}
				pages.push("ellipsis");
				pages.push(totalPages);
			}
		}

		return pages;
	};

	const pageNumbers = getPageNumbers();

	const handlePreviousClick = (e: React.MouseEvent<HTMLAnchorElement>) => {
		e.preventDefault();
		if (currentPage > 1) {
			onPageChange(currentPage - 1);
		}
	};

	const handleNextClick = (e: React.MouseEvent<HTMLAnchorElement>) => {
		e.preventDefault();
		if (currentPage < totalPages) {
			onPageChange(currentPage + 1);
		}
	};

	const handlePageClick = (page: number) => (e: React.MouseEvent<HTMLAnchorElement>) => {
		e.preventDefault();
		onPageChange(page);
	};

	return (
		<Pagination>
			<PaginationContent>
				<PaginationItem>
					<PaginationPrevious
						href="#"
						onClick={handlePreviousClick}
						className={cn(
							currentPage === 1 && "pointer-events-none opacity-50"
						)}
					/>
				</PaginationItem>
				{pageNumbers.map((page, index) => {
					if (page === "ellipsis") {
						return (
							<PaginationItem key={`ellipsis-${index}`}>
								<PaginationEllipsis />
							</PaginationItem>
						);
					}
					return (
						<PaginationItem key={page}>
							<PaginationLink
								href="#"
								onClick={handlePageClick(page)}
								isActive={currentPage === page}
							>
								{page}
							</PaginationLink>
						</PaginationItem>
					);
				})}
				<PaginationItem>
					<PaginationNext
						href="#"
						onClick={handleNextClick}
						className={cn(
							currentPage === totalPages && "pointer-events-none opacity-50"
						)}
					/>
				</PaginationItem>
			</PaginationContent>
		</Pagination>
	);
}

