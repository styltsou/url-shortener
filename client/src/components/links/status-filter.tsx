import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "@/components/ui/select";

export type StatusFilter = "all" | "active" | "inactive";

interface StatusFilterProps {
	value: StatusFilter;
	onValueChange: (value: StatusFilter) => void;
}

export function StatusFilter({ value, onValueChange }: StatusFilterProps) {
	return (
		<Select value={value} onValueChange={onValueChange}>
			<SelectTrigger className="h-8 w-[140px] text-xs" size="sm">
				<SelectValue />
			</SelectTrigger>
			<SelectContent>
				<SelectItem value="all">All links</SelectItem>
				<SelectItem value="active">Active links</SelectItem>
				<SelectItem value="inactive">Inactive links</SelectItem>
			</SelectContent>
		</Select>
	);
}

