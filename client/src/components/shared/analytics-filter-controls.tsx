import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "@/components/ui/select";
import {
	defaultIntervalForRange,
	type AnalyticsFilter,
	type AnalyticsInterval,
	type AnalyticsRange,
} from "@/lib/analytics-filter";

interface AnalyticsFilterControlsProps {
	filter: AnalyticsFilter;
	onChange: (filter: AnalyticsFilter) => void;
}

export function AnalyticsFilterControls({ filter, onChange }: AnalyticsFilterControlsProps) {
	return (
		<div className="flex flex-wrap gap-2" aria-label="Analytics time filters">
			<Select
				value={filter.range}
				onValueChange={(value) => {
					const range = value as AnalyticsRange;
					onChange({ range, interval: defaultIntervalForRange[range] });
				}}
			>
				<SelectTrigger aria-label="Time range" className="w-[150px]">
					<SelectValue />
				</SelectTrigger>
				<SelectContent>
					<SelectItem value="7d">Last 7 days</SelectItem>
					<SelectItem value="30d">Last 30 days</SelectItem>
					<SelectItem value="90d">Last 90 days</SelectItem>
					<SelectItem value="12m">Last 12 months</SelectItem>
				</SelectContent>
			</Select>
			<Select
				value={filter.interval}
				onValueChange={(value) =>
					onChange({ ...filter, interval: value as AnalyticsInterval })
				}
			>
				<SelectTrigger aria-label="Chart interval" className="w-[120px]">
					<SelectValue />
				</SelectTrigger>
				<SelectContent>
					<SelectItem value="day">Daily</SelectItem>
					<SelectItem value="week">Weekly</SelectItem>
					<SelectItem value="month">Monthly</SelectItem>
				</SelectContent>
			</Select>
			<span className="self-center text-xs text-muted-foreground">UTC</span>
		</div>
	);
}
