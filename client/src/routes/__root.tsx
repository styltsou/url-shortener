import { createRootRoute, Outlet, useRouterState } from '@tanstack/react-router'
import { ClerkProvider } from '@clerk/clerk-react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { Toaster } from '@/components/ui/sonner'
import { Header } from '@/components/Header'
import { ThemeProvider } from '@/components/theme-provider'

const clerkPubKey = import.meta.env.VITE_CLERK_PUBLISHABLE_KEY

if (!clerkPubKey) {
  throw new Error('Missing VITE_CLERK_PUBLISHABLE_KEY environment variable')
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
})

export const Route = createRootRoute({
  component: RootComponent,
})

function RootComponent() {
  const router = useRouterState()
  const isAuthPage = router.location.pathname === '/login' || router.location.pathname === '/sso-callback'

  return (
    <ClerkProvider publishableKey={clerkPubKey}>
      <ThemeProvider
        attribute="class"
        defaultTheme="system"
        enableSystem
        disableTransitionOnChange
      >
        <QueryClientProvider client={queryClient}>
          <div className="min-h-screen bg-background font-sans text-foreground selection:bg-primary/20 selection:text-primary-foreground">
            {!isAuthPage && <Header />}
            <Outlet />
            <Toaster />
          </div>
        </QueryClientProvider>
      </ThemeProvider>
    </ClerkProvider>
  )
}

