import { MessageSquare } from "lucide-react";
import { Link, useLocation } from "@tanstack/react-router";
import {
	SidebarMenu,
	SidebarMenuButton,
	SidebarMenuItem,
} from "@/components/ui/sidebar";

export function NavFeedback() {
	const location = useLocation();
	const isActive = location.pathname === "/feedback";

	return (
		<SidebarMenu>
			<SidebarMenuItem>
				<SidebarMenuButton asChild tooltip='Feedback' isActive={isActive}>
					<Link to='/feedback'>
						<MessageSquare />
						<span>Feedback</span>
					</Link>
				</SidebarMenuButton>
			</SidebarMenuItem>
		</SidebarMenu>
	);
}
