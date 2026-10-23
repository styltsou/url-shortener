import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
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
	const today = new Date().toISOString().slice(0, 10);
	const [from, setFrom] = useState(filter.from ?? today);
	const [to, setTo] = useState(filter.to ?? today);
	const fromTime = Date.parse(`${from}T00:00:00Z`);
	const toTime = Date.parse(`${to}T00:00:00Z`);
	const validCustomDates =
		!!from &&
		!!to &&
		Number.isFinite(fromTime) &&
		Number.isFinite(toTime) &&
		toTime >= fromTime &&
		toTime - fromTime <= 365 * 24 * 60 * 60 * 1000;

	return (
		<div className="flex flex-wrap items-center gap-2" aria-label="Analytics time filters">
			<Select
				value={filter.range}
				onValueChange={(value) => {
					const range = value as AnalyticsRange;
					if (range === "custom") {
						const start = filter.from ?? today;
						const end = filter.to ?? today;
						setFrom(start);
						setTo(end);
						onChange({ range, interval: "day", from: start, to: end });
					} else {
						onChange({ range, interval: defaultIntervalForRange[range] });
					}
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
					<SelectItem value="custom">Custom dates</SelectItem>
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
			{filter.range === "custom" && (
				<>
					<label className="flex items-center gap-2 text-sm">
						From
						<Input
							type="date"
							value={from}
							onChange={(event) => setFrom(event.target.value)}
							className="w-[155px]"
						/>
					</label>
					<label className="flex items-center gap-2 text-sm">
						To
						<Input
							type="date"
							value={to}
							onChange={(event) => setTo(event.target.value)}
							className="w-[155px]"
						/>
					</label>
					<Button
						type="button"
						variant="outline"
						disabled={!validCustomDates || (filter.from === from && filter.to === to)}
						onClick={() => onChange({ ...filter, from, to })}
					>
						Apply
					</Button>
					{!validCustomDates && (
						<span role="alert" className="text-xs text-destructive">
							Choose dates in order, covering at most 366 days.
						</span>
					)}
				</>
			)}
			<span className="self-center text-xs text-muted-foreground">UTC</span>
		</div>
	);
}
