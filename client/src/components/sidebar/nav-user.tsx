import {
	ChevronsUpDown,
	LogOut,
	Settings,
	User,
	CreditCard,
} from "lucide-react";
import { useUser, useClerk } from "@clerk/clerk-react";
import { useNavigate } from "@tanstack/react-router";

import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import {
	DropdownMenu,
	DropdownMenuContent,
	DropdownMenuGroup,
	DropdownMenuItem,
	DropdownMenuLabel,
	DropdownMenuSeparator,
	DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
	SidebarMenu,
	SidebarMenuButton,
	SidebarMenuItem,
	useSidebar,
} from "@/components/ui/sidebar";
import { NavUserSkeleton } from "./nav-user-skeleton";

export function NavUser() {
	const { isMobile } = useSidebar();
	const { user, isLoaded } = useUser();
	const { signOut } = useClerk();
	const navigate = useNavigate();

	if (!isLoaded) {
		return <NavUserSkeleton />;
	}

	if (!user) {
		return null;
	}

	const handleSignOut = async () => {
		await signOut();
		navigate({ to: "/login" });
	};

	const userInitials =
		user.fullName
			?.split(" ")
			.map((n) => n[0])
			.join("")
			.toUpperCase()
			.slice(0, 2) || "U";

	return (
		<SidebarMenu>
			<SidebarMenuItem>
				<DropdownMenu>
					<DropdownMenuTrigger asChild>
						<SidebarMenuButton
							size='lg'
							className='data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground'
						>
							<Avatar className='h-8 w-8 rounded-lg'>
								<AvatarImage
									src={user.imageUrl}
									alt={user.fullName || "User"}
								/>
								<AvatarFallback className='rounded-lg'>
									{userInitials}
								</AvatarFallback>
							</Avatar>
							<div className='grid flex-1 text-left text-sm leading-tight'>
								<span className='truncate font-medium'>
									{user.fullName || "User"}
								</span>
								<span className='truncate text-xs'>
									{user.primaryEmailAddress?.emailAddress}
								</span>
							</div>
							<ChevronsUpDown className='ml-auto size-4' />
						</SidebarMenuButton>
					</DropdownMenuTrigger>
					<DropdownMenuContent
						className='w-(--radix-dropdown-menu-trigger-width) min-w-56 rounded-lg'
						side={isMobile ? "bottom" : "right"}
						align='end'
						sideOffset={4}
					>
						<DropdownMenuLabel className='p-0 font-normal'>
							<div className='flex items-center gap-2 px-1 py-1.5 text-left text-sm'>
								<Avatar className='h-8 w-8 rounded-lg'>
									<AvatarImage
										src={user.imageUrl}
										alt={user.fullName || "User"}
									/>
									<AvatarFallback className='rounded-lg'>
										{userInitials}
									</AvatarFallback>
								</Avatar>
								<div className='grid flex-1 text-left text-sm leading-tight'>
									<span className='truncate font-medium'>
										{user.fullName || "User"}
									</span>
									<span className='truncate text-xs'>
										{user.primaryEmailAddress?.emailAddress}
									</span>
								</div>
							</div>
						</DropdownMenuLabel>
						<DropdownMenuSeparator />
						<DropdownMenuGroup>
							<DropdownMenuItem onClick={() => navigate({ to: "/account" })}>
								<User />
								Account
							</DropdownMenuItem>
							<DropdownMenuItem onClick={() => navigate({ to: "/settings" })}>
								<Settings />
								Settings
							</DropdownMenuItem>
							<DropdownMenuItem onClick={() => navigate({ to: "/billing" })}>
								<CreditCard />
								Billing
							</DropdownMenuItem>
						</DropdownMenuGroup>
						<DropdownMenuSeparator />
						<DropdownMenuItem onClick={handleSignOut}>
							<LogOut />
							Log out
						</DropdownMenuItem>
					</DropdownMenuContent>
				</DropdownMenu>
			</SidebarMenuItem>
		</SidebarMenu>
	);
}
