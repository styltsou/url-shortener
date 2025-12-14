"use client";

import * as React from "react";
import { LayoutDashboard, Link2, Link as LinkIcon } from "lucide-react";
import { Link } from "@tanstack/react-router";

import { NavMain } from "@/components/sidebar/nav-main";
import { NavUser } from "@/components/sidebar/nav-user";
import { NavFeedback } from "@/components/sidebar/nav-feedback";
import {
	Sidebar,
	SidebarContent,
	SidebarFooter,
	SidebarHeader,
	SidebarRail,
	SidebarMenu,
	SidebarMenuButton,
	SidebarMenuItem,
} from "@/components/ui/sidebar";

export function AppSidebar({ ...props }: React.ComponentProps<typeof Sidebar>) {
	const navItems = [
		{
			title: "Dashboard",
			url: "/",
			icon: LayoutDashboard,
		},
		{
			title: "Links",
			url: "/links",
			icon: Link2,
		},
	];

	return (
		<Sidebar collapsible='icon' {...props}>
			<SidebarHeader>
				<SidebarMenu>
					<SidebarMenuItem>
						<SidebarMenuButton
							size='lg'
							asChild
							className='data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground hover:bg-transparent hover:text-inherit active:bg-transparent active:text-inherit'
						>
							<Link to='/'>
								<div className='flex aspect-square size-8 items-center justify-center rounded-lg'>
									<LinkIcon className='size-4' />
								</div>
								<div className='grid flex-1 text-left text-sm leading-tight'>
									<span className='truncate font-semibold'>link4.it</span>
								</div>
							</Link>
						</SidebarMenuButton>
					</SidebarMenuItem>
				</SidebarMenu>
			</SidebarHeader>
			<SidebarContent>
				<NavMain items={navItems} />
			</SidebarContent>
			<SidebarFooter>
				<NavFeedback />
				<NavUser />
			</SidebarFooter>
			<SidebarRail />
		</Sidebar>
	);
}
