import { useState, useEffect } from "react"
import { Link } from "react-router-dom"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Navbar } from "@/components/Navbar"
import { getDomains, shorten, listShortURLs } from "@/services/api"
import { ShortURL } from "@/types/api"

export function Dashboard() {
  const [domains, setDomains] = useState<string[]>([])
  const [selectedDomain, setSelectedDomain] = useState<string>("")
  const [url, setUrl] = useState<string>("")
  const [slug, setSlug] = useState<string>("")
  const [shortURLs, setShortURLs] = useState<ShortURL[]>([])
  const [currentPage, setCurrentPage] = useState<number>(1)
  const [totalPages, setTotalPages] = useState<number>(1)
  const [total, setTotal] = useState<number>(0)
  const [loading, setLoading] = useState<boolean>(false)
  const [submitting, setSubmitting] = useState<boolean>(false)
  const [error, setError] = useState<string>("")
  const [success, setSuccess] = useState<string>("")

  const limit = 20

  // Load domains on mount
  useEffect(() => {
    const loadDomains = async () => {
      try {
        const domainList = await getDomains()
        setDomains(domainList)
        if (domainList.length > 0) {
          setSelectedDomain(domainList[0])
        }
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to load domains")
      }
    }
    loadDomains()
  }, [])

  // Load short URLs on mount and when page changes
  useEffect(() => {
    const loadShortURLs = async () => {
      setLoading(true)
      setError("")
      try {
        const response = await listShortURLs(currentPage, limit)
        setShortURLs(response.urls)
        setTotalPages(response.total_pages)
        setTotal(response.total)
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to load short URLs")
      } finally {
        setLoading(false)
      }
    }
    loadShortURLs()
  }, [currentPage])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setSubmitting(true)
    setError("")
    setSuccess("")

    try {
      await shorten(selectedDomain, url, slug || undefined)
      setSuccess("Short URL created successfully!")
      // Reset form
      setUrl("")
      setSlug("")
      // Refresh the list
      const response = await listShortURLs(currentPage, limit)
      setShortURLs(response.urls)
      setTotalPages(response.total_pages)
      setTotal(response.total)
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create short URL")
    } finally {
      setSubmitting(false)
    }
  }

  const formatDate = (dateString: string) => {
    const date = new Date(dateString)
    return date.toLocaleDateString("en-US", {
      year: "numeric",
      month: "short",
      day: "numeric",
    })
  }

  return (
    <div className="min-h-screen bg-background">
      <Navbar />
      <div className="p-8">
        <div className="max-w-6xl mx-auto space-y-8">
          <h1 className="text-4xl font-bold">Dashboard</h1>

          {/* Section 1: Create Short URL */}
          <Card>
            <CardHeader>
              <CardTitle>Create Short URL</CardTitle>
              <CardDescription>
                Create a new shortened URL by selecting a domain and providing the target URL.
              </CardDescription>
            </CardHeader>
            <CardContent>
              <form onSubmit={handleSubmit} className="space-y-4">
                {error && (
                  <div className="p-3 text-sm text-red-600 bg-red-50 border border-red-200 rounded-md">
                    {error}
                  </div>
                )}
                {success && (
                  <div className="p-3 text-sm text-green-600 bg-green-50 border border-green-200 rounded-md">
                    {success}
                  </div>
                )}
                {/* Form inputs: stacked on mobile, inline on desktop */}
                <div className="flex flex-col md:flex-row gap-3 md:items-end">
                  {/* Domain select */}
                  <div className="flex-1 min-w-0">
                    <label htmlFor="domain" className="block text-sm font-medium mb-1.5 md:hidden">
                      Domain
                    </label>
                    <select
                      id="domain"
                      value={selectedDomain}
                      onChange={(e) => setSelectedDomain(e.target.value)}
                      required
                      className="w-full px-3 py-2 border border-input bg-background rounded-md focus:outline-none focus:ring-2 focus:ring-ring text-sm"
                      disabled={submitting || domains.length === 0}
                    >
                      {domains.map((domain) => (
                        <option key={domain} value={domain}>
                          {domain}
                        </option>
                      ))}
                    </select>
                  </div>
                  
                  {/* URL input */}
                  <div className="flex-[2] min-w-0">
                    <label htmlFor="url" className="block text-sm font-medium mb-1.5 md:hidden">
                      URL <span className="text-red-500">*</span>
                    </label>
                    <input
                      id="url"
                      type="url"
                      value={url}
                      onChange={(e) => setUrl(e.target.value)}
                      required
                      placeholder="https://example.com"
                      className="w-full px-3 py-2 border border-input bg-background rounded-md focus:outline-none focus:ring-2 focus:ring-ring text-sm"
                      disabled={submitting}
                    />
                  </div>
                  
                  {/* Slug input */}
                  <div className="flex-1 min-w-0">
                    <label htmlFor="slug" className="block text-sm font-medium mb-1.5 md:hidden">
                      Custom Slug <span className="text-muted-foreground font-normal">(optional)</span>
                    </label>
                    <input
                      id="slug"
                      type="text"
                      value={slug}
                      onChange={(e) => setSlug(e.target.value)}
                      placeholder="Custom slug (optional)"
                      className="w-full px-3 py-2 border border-input bg-background rounded-md focus:outline-none focus:ring-2 focus:ring-ring text-sm"
                      disabled={submitting}
                    />
                  </div>
                  
                  {/* Submit button */}
                  <div className="flex md:flex-shrink-0">
                    <Button 
                      type="submit" 
                      disabled={submitting}
                      className="w-full md:w-auto whitespace-nowrap"
                    >
                      {submitting ? "Creating..." : "Shorten"}
                    </Button>
                  </div>
                </div>
              </form>
            </CardContent>
          </Card>

          {/* Section 2: Short URLs List */}
          <Card>
            <CardHeader>
              <CardTitle>Your Short URLs</CardTitle>
              <CardDescription>
                {total > 0 ? `Showing ${shortURLs.length} of ${total} shortened URLs` : "No shortened URLs yet"}
              </CardDescription>
            </CardHeader>
            <CardContent>
              {loading ? (
                <div className="text-center py-8 text-muted-foreground">Loading...</div>
              ) : shortURLs.length === 0 ? (
                <div className="text-center py-8 text-muted-foreground">
                  No shortened URLs found. Create your first one above!
                </div>
              ) : (
                <>
                  <div className="overflow-x-auto">
                    <table className="w-full border-collapse">
                      <thead>
                        <tr className="border-b">
                          <th className="text-left p-3 font-medium">Short URL</th>
                          <th className="text-left p-3 font-medium">Original URL</th>
                          <th className="text-left p-3 font-medium">Domain</th>
                          <th className="text-left p-3 font-medium">Slug</th>
                          <th className="text-left p-3 font-medium">Created</th>
                          <th className="text-left p-3 font-medium">Actions</th>
                        </tr>
                      </thead>
                      <tbody>
                        {shortURLs.map((shortURL) => (
                          <tr key={shortURL.id} className="border-b hover:bg-muted/50">
                            <td className="p-3">
                              <a
                                href={`http://${shortURL.domain}/${shortURL.slug}`}
                                target="_blank"
                                rel="noopener noreferrer"
                                className="text-primary hover:underline"
                              >
                                {shortURL.domain}/{shortURL.slug}
                              </a>
                            </td>
                            <td className="p-3">
                              <a
                                href={shortURL.url}
                                target="_blank"
                                rel="noopener noreferrer"
                                className="text-primary hover:underline truncate max-w-md block"
                                title={shortURL.url}
                              >
                                {shortURL.url}
                              </a>
                            </td>
                            <td className="p-3 text-muted-foreground">{shortURL.domain}</td>
                            <td className="p-3 text-muted-foreground font-mono text-sm">{shortURL.slug}</td>
                            <td className="p-3 text-muted-foreground">{formatDate(shortURL.created_at)}</td>
                            <td className="p-3">
                              <Link to={`/short-urls/${shortURL.id}`}>
                                <Button variant="outline" size="sm">
                                  View More
                                </Button>
                              </Link>
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                  {totalPages > 1 && (
                    <div className="flex items-center justify-between mt-6">
                      <div className="text-sm text-muted-foreground">
                        Page {currentPage} of {totalPages}
                      </div>
                      <div className="flex gap-2">
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => setCurrentPage((p) => Math.max(1, p - 1))}
                          disabled={currentPage === 1 || loading}
                        >
                          Previous
                        </Button>
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => setCurrentPage((p) => Math.min(totalPages, p + 1))}
                          disabled={currentPage === totalPages || loading}
                        >
                          Next
                        </Button>
                      </div>
                    </div>
                  )}
                </>
              )}
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  )
}

