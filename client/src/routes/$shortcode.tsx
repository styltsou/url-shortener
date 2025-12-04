import { createFileRoute, useNavigate } from '@tanstack/react-router'
import { useAuth } from '@clerk/clerk-react'
import { Navigate } from '@tanstack/react-router'
import { useState, useEffect } from 'react'
import {
  ArrowRight,
  Calendar,
  Copy,
  CheckCircle2,
  Edit,
  Save,
  Globe,
  Clock,
  TrendingUp,
  Trash2,
  Loader2,
} from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { ClicksGraph } from '@/components/ClicksGraph'
import { formatDate } from '@/lib/mock-data'
import { useLinks, useUpdateLink, useDeleteLink } from '@/hooks/use-links'
import { toast } from 'sonner'

export const Route = createFileRoute('/$shortcode')({
  component: LinkDetailPage,
})

function LinkDetailPage() {
  const { shortcode } = Route.useParams()
  const navigate = useNavigate()
  const { isSignedIn, isLoaded } = useAuth()
  const { data: urls = [], isLoading: isLoadingLinks } = useLinks()
  const updateLink = useUpdateLink()
  const deleteLink = useDeleteLink()
  const [isEditing, setIsEditing] = useState(false)
  const [copied, setCopied] = useState(false)
  const [originalUrlInput, setOriginalUrlInput] = useState('')
  const [expirationDateInput, setExpirationDateInput] = useState('')

  const url = urls.find((u) => u.shortCode === shortcode)

  // Initialize form state when entering edit mode
  useEffect(() => {
    if (isEditing && url && !originalUrlInput) {
      setOriginalUrlInput(url.originalUrl)
      setExpirationDateInput(url.expiresAt ? new Date(url.expiresAt).toISOString().split('T')[0] : '')
    }
    if (!isEditing) {
      setOriginalUrlInput('')
      setExpirationDateInput('')
    }
  }, [isEditing, url, originalUrlInput])

  if (!isLoaded) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <div className="text-lg">Loading...</div>
      </div>
    )
  }

  if (!isSignedIn) {
    return <Navigate to="/login" />
  }

  if (isLoadingLinks) {
    return (
      <main className="py-12 px-4 sm:px-6">
        <div className="max-w-6xl mx-auto text-center py-20">
          <div className="text-lg text-muted-foreground">Loading link...</div>
        </div>
      </main>
    )
  }

  if (!url) {
    return (
      <main className="py-12 px-4 sm:px-6">
        <div className="max-w-6xl mx-auto text-center py-20">
          <h1 className="text-3xl font-bold mb-4">Link not found</h1>
          <Button onClick={() => navigate({ to: '/' })}>Back to dashboard</Button>
        </div>
      </main>
    )
  }

  const handleSave = async () => {
    if (!url) return
    
    await updateLink.mutateAsync({
      id: url.id,
      data: {
        expires_at: expirationDateInput || null,
        // TODO: Add original_url update when backend supports it
      },
    })
    setIsEditing(false)
  }

  const handleDelete = async () => {
    if (!url) return
    
    if (window.confirm('Are you sure you want to delete this short URL? This action cannot be undone.')) {
      await deleteLink.mutateAsync(url.id)
      navigate({ to: '/' })
    }
  }

  const handleCopy = () => {
    navigator.clipboard.writeText(`https://short.ly/${url.shortCode}`)
    setCopied(true)
    toast.success('Copied to clipboard')
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <main className="py-12 px-4 sm:px-6">
      <div className="max-w-6xl mx-auto">
        <div className="mb-8">
          <Button
            variant="ghost"
            onClick={() => navigate({ to: '/' })}
            className="group flex items-center text-sm font-medium text-muted-foreground hover:text-foreground mb-4"
          >
            <div className="p-1 rounded-full bg-muted group-hover:bg-accent mr-2 transition-colors">
              <ArrowRight className="w-4 h-4 rotate-180" />
            </div>
            Back to dashboard
          </Button>

          <div className="flex flex-col md:flex-row md:items-center justify-between gap-6">
            <div>
              <div className="flex items-center gap-3">
                <h1 className="text-3xl font-bold text-foreground tracking-tight">short.ly/{url.shortCode}</h1>
                <Button variant="ghost" size="icon" onClick={handleCopy} className="text-muted-foreground hover:text-foreground">
                  {copied ? <CheckCircle2 className="w-6 h-6 text-primary" /> : <Copy className="w-6 h-6" />}
                </Button>
              </div>
              <div className="flex items-center gap-2 mt-2 text-muted-foreground text-sm">
                <Calendar className="w-4 h-4" />
                <span>Created {formatDate(url.createdAt)}</span>
              </div>
            </div>

            <div className="flex gap-3">
              {isEditing ? (
                <>
                  <Button variant="outline" onClick={() => setIsEditing(false)}>
                    Cancel
                  </Button>
                  <Button onClick={handleSave} disabled={updateLink.isPending}>
                    {updateLink.isPending ? <Loader2 className="w-4 h-4 animate-spin" /> : <Save className="w-4 h-4" />} Save Changes
                  </Button>
                </>
              ) : (
                <>
                  <Button variant="outline" onClick={() => setIsEditing(true)}>
                    <Edit className="w-4 h-4" /> Edit
                  </Button>
                  <Button variant="destructive" onClick={handleDelete} disabled={deleteLink.isPending}>
                    {deleteLink.isPending ? <Loader2 className="w-4 h-4 animate-spin" /> : <Trash2 className="w-4 h-4" />} Delete
                  </Button>
                </>
              )}
            </div>
          </div>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
          {/* Main Content */}
          <div className="lg:col-span-2 space-y-6">
            {/* Destination Card */}
            <Card>
              <CardHeader>
                <CardTitle className="text-sm font-semibold uppercase tracking-wider flex items-center gap-2 text-muted-foreground">
                  <Globe className="w-4 h-4" /> Destination
                </CardTitle>
              </CardHeader>
              <CardContent>
                {isEditing ? (
                  <Input
                    type="url"
                    value={originalUrlInput}
                    onChange={(e) => setOriginalUrlInput(e.target.value)}
                  />
                ) : (
                  <div className="flex items-center gap-3 p-4 bg-muted rounded-lg">
                    <a
                      href={url.originalUrl}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="text-foreground break-all hover:text-primary transition-colors font-medium"
                    >
                      {url.originalUrl}
                    </a>
                  </div>
                )}

                <div className="mt-6 pt-6 border-t border-border">
                  <CardTitle className="text-sm font-semibold uppercase tracking-wider mb-4 flex items-center gap-2 text-muted-foreground">
                    <Clock className="w-4 h-4" /> Expiration
                  </CardTitle>
                  {isEditing ? (
                    <Input
                      type="date"
                      value={expirationDateInput}
                      onChange={(e) => setExpirationDateInput(e.target.value)}
                      className="bg-muted"
                    />
                  ) : (
                    <p className={`font-medium ${url.expiresAt && new Date(url.expiresAt) < new Date() ? 'text-destructive' : 'text-muted-foreground'}`}>
                      {url.expiresAt ? formatDate(url.expiresAt) : 'No expiration date set'}
                    </p>
                  )}
                </div>
              </CardContent>
            </Card>

            {/* Chart Card */}
            <Card>
              <CardHeader>
                <div className="flex items-center justify-between">
                  <CardTitle className="text-sm font-semibold uppercase tracking-wider flex items-center gap-2 text-muted-foreground">
                    <TrendingUp className="w-4 h-4" /> Performance
                  </CardTitle>
                  <Select defaultValue="7days">
                    <SelectTrigger className="w-[140px]">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="7days">Last 7 days</SelectItem>
                      <SelectItem value="30days">Last 30 days</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
              </CardHeader>
              <CardContent>
                <ClicksGraph data={url.analytics.clicks_data} />
              </CardContent>
            </Card>
          </div>

          {/* Sidebar Stats */}
          <div className="grid grid-cols-2 lg:grid-cols-1 gap-6">
            <Card>
              <CardHeader>
                <CardTitle className="text-sm font-semibold uppercase tracking-wider flex items-center gap-2 text-muted-foreground">
                  <TrendingUp className="w-4 h-4" /> Total Clicks
                </CardTitle>
              </CardHeader>
              <CardContent>
                <p className="text-4xl font-bold tracking-tight text-foreground mb-4">{url.clicks.toLocaleString()}</p>
                <div className="flex items-center gap-2 text-sm text-muted-foreground">
                  <TrendingUp className="w-4 h-4" />
                  <span>+12.5% this week</span>
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle className="text-sm font-semibold uppercase tracking-wider text-muted-foreground">Top Sources</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-4">
                  {url.analytics.referrers_data.map((item, idx) => {
                    const maxClicks = Math.max(...url.analytics.referrers_data.map((d) => d.clicks))
                    return (
                      <div key={idx} className="group">
                        <div className="flex justify-between items-center text-sm mb-1.5">
                          <span className="font-medium text-foreground">{item.referrer}</span>
                          <span className="text-muted-foreground">{item.clicks}</span>
                        </div>
                        <div className="w-full bg-muted rounded-full h-2 overflow-hidden">
                          <div
                            className="h-full bg-primary rounded-full transition-all duration-500"
                            style={{ width: `${(item.clicks / maxClicks) * 100}%` }}
                          />
                        </div>
                      </div>
                    )
                  })}
                </div>
              </CardContent>
            </Card>
          </div>
        </div>
      </div>
    </main>
  )
}
