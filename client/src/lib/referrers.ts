import {
	Globe,
	Twitter,
	Instagram,
	Facebook,
	Linkedin,
	MessageCircle,
	Mail,
	Newspaper,
} from "lucide-react";
import type { LucideIcon } from "lucide-react";

export interface ReferrerConfig {
	match: RegExp;
	icon: LucideIcon;
	label: string;
}

// Known referrer sources with their icons
export const KNOWN_REFERRERS: ReferrerConfig[] = [
	{ match: /^direct|none$/i, icon: Globe, label: "Direct" },
	{
		match: /twitter|twitter\.com|x\.com/i,
		icon: Twitter,
		label: "X (Twitter)",
	},
	{ match: /instagram|instagram\.com/i, icon: Instagram, label: "Instagram" },
	{ match: /facebook|facebook\.com/i, icon: Facebook, label: "Facebook" },
	{ match: /linkedin|linkedin\.com/i, icon: Linkedin, label: "LinkedIn" },
	{ match: /reddit|reddit\.com/i, icon: MessageCircle, label: "Reddit" },
	{ match: /^email$/i, icon: Mail, label: "Email" },
	{ match: /newsletter/i, icon: Newspaper, label: "Newsletter" },
];

// Get icon component for a referrer
export function getReferrerIcon(referrer: string): LucideIcon | null {
	const normalized = referrer.toLowerCase().trim();
	const known = KNOWN_REFERRERS.find((r) => r.match.test(normalized));
	return known ? known.icon : null;
}

// Get display label for a referrer
export function getReferrerLabel(referrer: string): string {
	const normalized = referrer.toLowerCase().trim();
	const known = KNOWN_REFERRERS.find((r) => r.match.test(normalized));
	return known ? known.label : referrer;
}

// Process referrers data: merge unknown sources into "Other"
export function processReferrersData(
	referrersData: Array<{ referrer: string; clicks: number }>
): Array<{ referrer: string; clicks: number }> {
	const known: Array<{ referrer: string; clicks: number }> = [];
	let otherClicks = 0;

	referrersData.forEach((item) => {
		const icon = getReferrerIcon(item.referrer);
		if (icon) {
			// Known referrer - use standardized label
			const label = getReferrerLabel(item.referrer);
			const existing = known.find((k) => k.referrer === label);
			if (existing) {
				existing.clicks += item.clicks;
			} else {
				known.push({ referrer: label, clicks: item.clicks });
			}
		} else {
			// Unknown referrer - add to "Other"
			otherClicks += item.clicks;
		}
	});

	// Sort by clicks descending
	known.sort((a, b) => b.clicks - a.clicks);

	// Add "Other" if there are unknown sources
	if (otherClicks > 0) {
		known.push({ referrer: "Other", clicks: otherClicks });
	}

	return known;
}

