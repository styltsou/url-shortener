import { Link, ExternalLink } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { DATETIME_FORMAT_OPTIONS } from "@/lib/constants";
import { SHORT_DOMAIN } from "@/lib/env";
import type { RecentLink } from "@/hooks/use-dashboard";

interface RecentLinksProps {
	links: RecentLink[];
}

export function RecentLinks({ links }: RecentLinksProps) {
	if (links.length === 0) {
		return null;
	}

	return (
		<Card>
			<CardHeader>
				<CardTitle className="text-sm font-semibold uppercase tracking-wider text-muted-foreground">
					Recent Links
				</CardTitle>
			</CardHeader>
			<CardContent>
				<div className="space-y-3">
					{links.map((link) => (
						<a
							key={link.id}
							href={`/links/${link.shortcode}`}
							className="flex items-center gap-3 p-3 rounded-lg hover:bg-muted/50 transition-colors group border border-transparent hover:border-border"
						>
							<div className="flex h-9 w-9 items-center justify-center rounded-lg bg-primary/10 text-primary shrink-0">
								<Link className="h-4 w-4" />
							</div>
							<div className="flex-1 min-w-0">
								<div className="flex items-center gap-2 flex-wrap">
									<span className="text-sm font-medium text-foreground truncate">
										{SHORT_DOMAIN}/{link.shortcode}
									</span>
									<Badge
										variant={link.is_active ? "default" : "secondary"}
										className="text-[10px] px-1.5 py-0 h-4"
									>
										{link.is_active ? "active" : "inactive"}
									</Badge>
								</div>
								<p className="text-xs text-muted-foreground truncate mt-0.5">{link.original_url}</p>
							</div>
							<div className="flex items-center gap-3 text-xs text-muted-foreground shrink-0">
								<span>
									{new Date(link.created_at).toLocaleDateString("en-US", DATETIME_FORMAT_OPTIONS)}
								</span>
								<ExternalLink className="h-3.5 w-3.5 opacity-0 group-hover:opacity-100 transition-opacity" />
							</div>
						</a>
					))}
				</div>
			</CardContent>
		</Card>
	);
}
