import { Card, CardContent } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import type { Url } from '@/types/url'
import { formatDate } from '@/lib/mock-data'
import { useNavigate } from '@tanstack/react-router'
import { toast } from 'sonner'
import { useState } from 'react'
import { BarChart3, Calendar, Copy, CheckCircle2, MoreHorizontal, Power, Trash2 } from 'lucide-react'

interface UrlCardProps {
  url: Url
}

export function UrlCard({ url }: UrlCardProps) {
  const navigate = useNavigate()
  const isExpired = url.expiresAt && new Date(url.expiresAt) < new Date()
  const shortLink = `short.ly/${url.shortCode}`
  const [copied, setCopied] = useState(false)
  const [isActive, setIsActive] = useState(true) // Local state for UI only

  const handleCopy = (e: React.MouseEvent) => {
    e.stopPropagation()
    navigator.clipboard.writeText(`https://${shortLink}`)
    setCopied(true)
    toast.success('Copied to clipboard')
    setTimeout(() => setCopied(false), 2000)
  }

  const handleClick = () => {
    navigate({ to: '/$shortcode', params: { shortcode: url.shortCode } })
  }

  const handleToggleActive = (e: React.MouseEvent) => {
    e.stopPropagation()
    setIsActive(!isActive)
    toast.success(isActive ? 'Link deactivated' : 'Link activated')
  }

  const handleDelete = (e: React.MouseEvent) => {
    e.stopPropagation()
    toast.info('Delete link action (UI only)')
  }

  return (
    <Card
      onClick={handleClick}
      className="group relative cursor-pointer hover:shadow-md transition-all"
    >
      <CardContent className="py-3 px-5">
        <div className="flex flex-col md:flex-row md:items-center justify-between gap-2">
          <div className="flex-1 min-w-0">
            <div className="flex items-center gap-3 mb-0.5">
              <div className="flex items-center gap-2">
                <h3 className="text-lg font-bold text-foreground tracking-tight">{shortLink}</h3>
                {isExpired && (
                  <Badge variant="destructive" className="text-[10px] font-bold uppercase tracking-wider">
                    Expired
                  </Badge>
                )}
              </div>
            </div>
            <div className="flex items-center text-sm text-muted-foreground truncate">
              <span className="truncate hover:text-foreground transition-colors">{url.originalUrl}</span>
            </div>
          </div>

          <div className="flex items-center gap-6 text-sm text-muted-foreground">
            <div className="flex items-center gap-1.5" title="Total Clicks">
              <BarChart3 className="w-4 h-4 text-muted-foreground" />
              <span className="font-semibold text-foreground">{url.clicks.toLocaleString()}</span>
            </div>
            <div className="flex items-center gap-1.5 hidden sm:flex">
              <Calendar className="w-4 h-4 text-muted-foreground" />
              <span>{formatDate(url.createdAt)}</span>
            </div>

            <div className="flex items-center gap-2 pl-4 border-l border-border">
              <Button
                variant="ghost"
                size="icon"
                onClick={handleCopy}
                className={`h-8 w-8 ${copied ? 'bg-primary/10 text-primary' : ''}`}
              >
                {copied ? <CheckCircle2 className="w-4 h-4" /> : <Copy className="w-4 h-4" />}
              </Button>
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button
                    variant="ghost"
                    size="icon"
                    onClick={(e) => {
                      e.stopPropagation()
                    }}
                    className="h-8 w-8"
                  >
                    <MoreHorizontal className="w-4 h-4" />
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end" onClick={(e) => e.stopPropagation()}>
                  <DropdownMenuItem onClick={handleToggleActive}>
                    <Power className="w-4 h-4 mr-2" />
                    {isActive ? 'Deactivate' : 'Activate'}
                  </DropdownMenuItem>
                  <DropdownMenuSeparator />
                  <DropdownMenuItem onClick={handleDelete} variant="destructive">
                    <Trash2 className="w-4 h-4 mr-2" />
                    Delete link
                  </DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  )
}

