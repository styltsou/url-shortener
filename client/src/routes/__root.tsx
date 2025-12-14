import {
	createRootRoute,
	Outlet,
	useRouterState,
} from "@tanstack/react-router";
import { ClerkProvider } from "@clerk/clerk-react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { Toaster } from "@/components/ui/sonner";
import { AppSidebar } from "@/components/sidebar/app-sidebar";
import { SidebarProvider, SidebarInset } from "@/components/ui/sidebar";
import { Header } from "@/components/header";
import { ThemeProvider } from "@/components/theme/theme-provider";
import { NavigationBlockerProvider } from "@/hooks/use-block-navigation";
import { getClerkPublishableKey } from "@/lib/env";

// Validate environment variables on module load
try {
	getClerkPublishableKey();
} catch (error) {
	if (error instanceof Error) {
		throw new Error(`Environment validation failed: ${error.message}`);
	}
	throw error;
}

const queryClient = new QueryClient({
	defaultOptions: {
		queries: {
			retry: false, // Disable retries for queries
			refetchOnWindowFocus: false,
		},
		mutations: {
			retry: false, // Disable retries for mutations
		},
	},
});

export const Route = createRootRoute({
	component: RootComponent,
});

function RootComponent() {
	const router = useRouterState();
	const isAuthPage =
		router.location.pathname === "/login" ||
		router.location.pathname === "/sso-callback";

	return (
		<ClerkProvider publishableKey={getClerkPublishableKey()}>
			<ThemeProvider
				attribute='class'
				defaultTheme='system'
				enableSystem
				disableTransitionOnChange
			>
				<QueryClientProvider client={queryClient}>
					<NavigationBlockerProvider>
						<div className='min-h-screen bg-background font-sans text-foreground selection:bg-primary/20 selection:text-primary-foreground'>
							{isAuthPage ? (
								<>
									<Outlet />
									<Toaster />
								</>
							) : (
								<SidebarProvider>
									<AppSidebar />
									<SidebarInset>
										<Header />
										<Outlet />
									</SidebarInset>
									<Toaster />
								</SidebarProvider>
							)}
						</div>
					</NavigationBlockerProvider>
				</QueryClientProvider>
			</ThemeProvider>
		</ClerkProvider>
	);
}
