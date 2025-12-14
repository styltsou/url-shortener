import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "@/components/ui/select";

interface PageSizeSelectorProps {
	value: number;
	onValueChange: (value: number) => void;
}

const PAGE_SIZE_OPTIONS = [5, 10, 15, 20] as const;

export function PageSizeSelector({
	value,
	onValueChange,
}: PageSizeSelectorProps) {
	return (
		<div className="flex items-center gap-2">
			<label className="text-xs text-muted-foreground whitespace-nowrap">
				Items per page:
			</label>
			<Select
				value={value.toString()}
				onValueChange={(val) => onValueChange(Number(val))}
			>
				<SelectTrigger className="h-8 w-[80px] text-xs">
					<SelectValue />
				</SelectTrigger>
				<SelectContent>
					{PAGE_SIZE_OPTIONS.map((size) => (
						<SelectItem key={size} value={size.toString()}>
							{size}
						</SelectItem>
					))}
				</SelectContent>
			</Select>
		</div>
	);
}

