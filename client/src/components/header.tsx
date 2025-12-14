import { SidebarTrigger } from "@/components/ui/sidebar";
import { ThemeToggle } from "@/components/theme/theme-toggle";

export function Header() {
	return (
		<header className='sticky top-0 z-30 flex h-16 shrink-0 items-center gap-2 border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60 px-4 transition-[width,height] duration-150 ease-linear group-has-data-[collapsible=icon]/sidebar-wrapper:h-12'>
			<SidebarTrigger />
			<div className='flex flex-1 items-center justify-end gap-4'>
				<ThemeToggle />
			</div>
		</header>
	);
}
