import { useState, useEffect } from "react"
import { useParams, useNavigate } from "react-router-dom"
import { ArrowLeft, Edit, Trash2, Save, X } from "lucide-react"
import { toast } from "sonner"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Navbar } from "@/components/Navbar"
import { getShortURL, updateShortURL, deleteShortURL, getDomains } from "@/services/api"
import { ShortURL } from "@/types/api"

export function ShortURLDetail() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [shortURL, setShortURL] = useState<ShortURL | null>(null)
  const [domains, setDomains] = useState<string[]>([])
  const [loading, setLoading] = useState<boolean>(true)
  const [error, setError] = useState<string>("")
  const [isEditing, setIsEditing] = useState<boolean>(false)
  const [submitting, setSubmitting] = useState<boolean>(false)
  const [deleting, setDeleting] = useState<boolean>(false)
  const [showDeleteConfirm, setShowDeleteConfirm] = useState<boolean>(false)

  // Form state
  const [editedUrl, setEditedUrl] = useState<string>("")
  const [editedDomain, setEditedDomain] = useState<string>("")
  const [editedSlug, setEditedSlug] = useState<string>("")

  // Load short URL data
  useEffect(() => {
    const loadData = async () => {
      if (!id) {
        setError("Invalid short URL ID")
        setLoading(false)
        return
      }

      setLoading(true)
      setError("")
      try {
        const [urlData, domainList] = await Promise.all([
          getShortURL(id),
          getDomains(),
        ])
        setShortURL(urlData)
        setDomains(domainList)
        setEditedUrl(urlData.url)
        setEditedDomain(urlData.domain)
        setEditedSlug(urlData.slug)
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to load short URL")
      } finally {
        setLoading(false)
      }
    }
    loadData()
  }, [id])

  const handleEdit = () => {
    setIsEditing(true)
  }

  const handleCancel = () => {
    setIsEditing(false)
    if (shortURL) {
      setEditedUrl(shortURL.url)
      setEditedDomain(shortURL.domain)
      setEditedSlug(shortURL.slug)
    }
  }

  const handleSave = async () => {
    if (!id || !shortURL) return

    setSubmitting(true)

    try {
      const updates: { url?: string; domain?: string; slug?: string } = {}
      
      if (editedUrl !== shortURL.url) {
        updates.url = editedUrl
      }
      if (editedDomain !== shortURL.domain) {
        updates.domain = editedDomain
      }
      if (editedSlug !== shortURL.slug) {
        updates.slug = editedSlug
      }

      if (Object.keys(updates).length === 0) {
        setIsEditing(false)
        setSubmitting(false)
        return
      }

      const updated = await updateShortURL(id, updates)
      setShortURL(updated)
      setIsEditing(false)
      toast.success("Short URL saved successfully!")
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : "Failed to update short URL"
      toast.error(errorMessage)
    } finally {
      setSubmitting(false)
    }
  }

  const handleDeleteClick = () => {
    setShowDeleteConfirm(true)
  }

  const handleDeleteCancel = () => {
    setShowDeleteConfirm(false)
  }

  const handleDeleteConfirm = async () => {
    if (!id) return

    setDeleting(true)

    try {
      await deleteShortURL(id)
      toast.success("Short URL deleted successfully!")
      navigate("/")
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : "Failed to delete short URL"
      toast.error(errorMessage)
      setShowDeleteConfirm(false)
    } finally {
      setDeleting(false)
    }
  }

  const formatDate = (dateString: string) => {
    const date = new Date(dateString)
    return date.toLocaleString("en-US", {
      year: "numeric",
      month: "short",
      day: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    })
  }

  if (loading) {
    return (
      <div className="min-h-screen bg-background pt-20">
        <Navbar />
        <div className="p-8">
          <div className="max-w-4xl mx-auto">
            <div className="text-center py-8 text-muted-foreground">Loading...</div>
          </div>
        </div>
      </div>
    )
  }

  if (error && !shortURL) {
    return (
      <div className="min-h-screen bg-background pt-20">
        <Navbar />
        <div className="p-8">
          <div className="max-w-4xl mx-auto">
            <Card className="bg-white dark:bg-background">
              <CardContent className="pt-6">
                <div className="text-center py-8">
                  <p className="text-red-600 mb-4">{error}</p>
                  <Button onClick={() => navigate("/")} variant="outline">
                    <ArrowLeft className="mr-2 h-4 w-4" />
                    Back to Dashboard
                  </Button>
                </div>
              </CardContent>
            </Card>
          </div>
        </div>
      </div>
    )
  }

  if (!shortURL) {
    return null
  }

  return (
    <>
      <div className="min-h-screen bg-background pt-20">
        <Navbar />
        <div className="p-8">
          <div className="max-w-4xl mx-auto space-y-6">
            <div className="flex items-center justify-between">
              <h1 className="text-4xl font-bold">Short URL Details</h1>
              <Button onClick={() => navigate("/")} variant="outline">
                <ArrowLeft className="mr-2 h-4 w-4" />
                Back to Dashboard
              </Button>
            </div>

            <Card className="bg-white dark:bg-background">
              <CardHeader>
                <div className="flex items-center justify-between">
                  <div>
                    <CardTitle>Short URL Information</CardTitle>
                    <CardDescription>
                      View and edit details for your shortened URL
                    </CardDescription>
                  </div>
                  {!isEditing ? (
                    <div className="flex gap-2">
                      <Button onClick={handleEdit} variant="outline">
                        <Edit className="mr-2 h-4 w-4" />
                        Edit
                      </Button>
                      <Button onClick={handleDeleteClick} variant="destructive">
                        <Trash2 className="mr-2 h-4 w-4" />
                        Delete
                      </Button>
                    </div>
                  ) : (
                    <div className="flex gap-2">
                      <Button
                        onClick={handleSave}
                        disabled={submitting}
                      >
                        <Save className="mr-2 h-4 w-4" />
                        {submitting ? "Saving..." : "Save"}
                      </Button>
                      <Button
                        onClick={handleCancel}
                        variant="outline"
                        disabled={submitting}
                      >
                        <X className="mr-2 h-4 w-4" />
                        Cancel
                      </Button>
                    </div>
                  )}
                </div>
              </CardHeader>
              <CardContent className="space-y-6">
                {isEditing && (
                  <div className="p-4 bg-yellow-50 border border-yellow-200 rounded-md">
                  <p className="text-sm text-yellow-800 font-medium">
                    Warning: Since we respond with 301 (Moved Permanently), users who have already visited this URL may not get the update.
                  </p>
                </div>
                )}

                <div className="space-y-4">
                  <div>
                    <label className="block text-sm font-medium mb-1.5">ID</label>
                    <div className="px-3 py-2 bg-muted rounded-md text-sm font-mono">
                      {shortURL.id}
                    </div>
                  </div>

                  <div>
                    <label className="block text-sm font-medium mb-1.5">
                      Original URL <span className="text-red-500">*</span>
                    </label>
                    {isEditing ? (
                      <input
                        type="url"
                        value={editedUrl}
                        onChange={(e) => setEditedUrl(e.target.value)}
                        className="w-full px-3 py-2 border border-border bg-background rounded-md focus:outline-none focus:ring-2 focus:ring-ring text-sm"
                        required
                      />
                    ) : (
                      <div className="px-3 py-2 bg-muted rounded-md text-sm break-all">
                        <a
                          href={shortURL.url}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="text-primary hover:underline"
                        >
                          {shortURL.url}
                        </a>
                      </div>
                    )}
                  </div>

                  <div>
                    <label className="block text-sm font-medium mb-1.5">
                      Domain <span className="text-red-500">*</span>
                    </label>
                    {isEditing ? (
                      <select
                        value={editedDomain}
                        onChange={(e) => setEditedDomain(e.target.value)}
                        className="w-full px-3 py-2 border border-border bg-background rounded-md focus:outline-none focus:ring-2 focus:ring-ring text-sm"
                        required
                      >
                        {domains.map((domain) => (
                          <option key={domain} value={domain}>
                            {domain}
                          </option>
                        ))}
                      </select>
                    ) : (
                      <div className="px-3 py-2 bg-muted rounded-md text-sm">
                        {shortURL.domain}
                      </div>
                    )}
                  </div>

                  <div>
                    <label className="block text-sm font-medium mb-1.5">
                      Slug <span className="text-red-500">*</span>
                    </label>
                    {isEditing ? (
                      <input
                        type="text"
                        value={editedSlug}
                        onChange={(e) => setEditedSlug(e.target.value)}
                        className="w-full px-3 py-2 border border-border bg-background rounded-md focus:outline-none focus:ring-2 focus:ring-ring text-sm font-mono"
                        required
                      />
                    ) : (
                      <div className="px-3 py-2 bg-muted rounded-md text-sm font-mono">
                        {shortURL.slug}
                      </div>
                    )}
                  </div>

                  <div>
                    <label className="block text-sm font-medium mb-1.5">Short URL</label>
                    <div className="px-3 py-2 bg-muted rounded-md text-sm">
                      <a
                        href={`http://${shortURL.domain}/${shortURL.slug}`}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="text-primary hover:underline font-mono"
                      >
                        {shortURL.domain}/{shortURL.slug}
                      </a>
                    </div>
                  </div>

                  <div>
                    <label className="block text-sm font-medium mb-1.5">User ID</label>
                    <div className="px-3 py-2 bg-muted rounded-md text-sm font-mono">
                      {shortURL.user_id}
                    </div>
                  </div>

                  <div>
                    <label className="block text-sm font-medium mb-1.5">Created At</label>
                    <div className="px-3 py-2 bg-muted rounded-md text-sm">
                      {formatDate(shortURL.created_at)}
                    </div>
                  </div>

                  <div>
                    <label className="block text-sm font-medium mb-1.5">Updated At</label>
                    <div className="px-3 py-2 bg-muted rounded-md text-sm">
                      {formatDate(shortURL.updated_at)}
                    </div>
                  </div>
                </div>
              </CardContent>
            </Card>
          </div>
        </div>
      </div>

      {/* Delete Confirmation Dialog */}
      {showDeleteConfirm && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-[60]">
          <Card className="max-w-md w-full mx-4">
            <CardHeader>
              <CardTitle>Delete Short URL</CardTitle>
              <CardDescription>
                Are you sure you want to delete this short URL? This action cannot be undone.
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                <div className="p-3 bg-muted rounded-md">
                  <div className="text-sm font-medium mb-1">Short URL:</div>
                  <div className="text-sm font-mono text-muted-foreground">
                    {shortURL.domain}/{shortURL.slug}
                  </div>
                  <div className="text-sm font-medium mb-1 mt-2">Original URL:</div>
                  <div className="text-sm text-muted-foreground break-all">
                    {shortURL.url}
                  </div>
                </div>
                <div className="flex gap-3">
                  <Button
                    onClick={handleDeleteConfirm}
                    variant="destructive"
                    disabled={deleting}
                    className="flex-1"
                  >
                    {deleting ? "Deleting..." : "Delete"}
                  </Button>
                  <Button
                    onClick={handleDeleteCancel}
                    variant="outline"
                    disabled={deleting}
                    className="flex-1"
                  >
                    Cancel
                  </Button>
                </div>
              </div>
            </CardContent>
          </Card>
        </div>
      )}
    </>
  )
}

