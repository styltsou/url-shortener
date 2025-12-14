import { ChevronsUpDown } from "lucide-react";
import {
	SidebarMenu,
	SidebarMenuButton,
	SidebarMenuItem,
} from "@/components/ui/sidebar";
import { Skeleton } from "@/components/ui/skeleton";

export function NavUserSkeleton() {
	return (
		<SidebarMenu>
			<SidebarMenuItem>
				<SidebarMenuButton
					size='lg'
					className='data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground'
					disabled
				>
					<Skeleton className='h-8 w-8 rounded-lg shrink-0' />
					<div className='grid flex-1 text-left text-sm leading-tight'>
						<Skeleton className='h-4 w-32 mb-1' />
						<Skeleton className='h-3 w-40' />
					</div>
					<ChevronsUpDown className='ml-auto size-4 opacity-50' />
				</SidebarMenuButton>
			</SidebarMenuItem>
		</SidebarMenu>
	);
}
