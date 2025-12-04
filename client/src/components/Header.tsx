import { Server } from 'lucide-react'
import { useNavigate } from '@tanstack/react-router'
import { useUser } from '@clerk/clerk-react'
import { ThemeToggle } from '@/components/theme-toggle'

export function Header() {
  const navigate = useNavigate()
  const { user } = useUser()

  return (
    <header className="sticky top-0 z-30 bg-background/80 backdrop-blur-md border-b border-border">
      <div className="max-w-6xl mx-auto px-4 sm:px-6 h-16 flex justify-between items-center">
        <div
          className="flex items-center gap-2 cursor-pointer group"
          onClick={() => navigate({ to: '/' })}
        >
          <div className="bg-primary text-primary-foreground p-1.5 rounded-lg group-hover:bg-primary/90 transition-colors">
            <Server className="w-5 h-5" />
          </div>
          <h1 className="text-xl font-bold text-foreground tracking-tight">GoShortener</h1>
        </div>

        <div className="flex items-center gap-4">
          <div className="hidden sm:block text-right">
            <p className="text-sm font-semibold text-foreground">Workspace</p>
          </div>
          <ThemeToggle />
          <div className="h-9 w-9 rounded-full bg-primary border-2 border-border shadow-sm cursor-pointer hover:scale-105 transition-transform">
            {user?.imageUrl && (
              <img
                src={user.imageUrl}
                alt={user.fullName || 'User'}
                className="h-full w-full rounded-full object-cover"
              />
            )}
          </div>
        </div>
      </div>
    </header>
  )
}

