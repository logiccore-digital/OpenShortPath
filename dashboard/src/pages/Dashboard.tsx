import { useState, useEffect } from "react"
import { Link } from "react-router-dom"
import { Scissors, Eye, ChevronLeft, ChevronRight, ChevronDown } from "lucide-react"
import { toast } from "sonner"
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
      toast.success("Short URL created successfully!")
      // Reset form
      setUrl("")
      setSlug("")
      // Refresh the list
      const response = await listShortURLs(currentPage, limit)
      setShortURLs(response.urls)
      setTotalPages(response.total_pages)
      setTotal(response.total)
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : "Failed to create short URL"
      setError(errorMessage)
      toast.error(errorMessage)
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
      <div className="px-6 md:px-8 lg:px-12 pb-6 md:pb-8 lg:pb-12 pt-24 transition-colors duration-300">
        <div className="max-w-7xl mx-auto w-full space-y-8">

          {/* Section 1: Create Short URL */}
          <Card>
            <CardHeader>
              <CardTitle>Create Short URL</CardTitle>
              <CardDescription>
                Create a new shortened URL by selecting a domain and providing the target URL.
              </CardDescription>
            </CardHeader>
            <CardContent>
              <form onSubmit={handleSubmit} className="space-y-2">
                {error && (
                  <div className="p-3 text-sm text-red-400 bg-red-950/20 border border-red-900/50">
                    {error}
                  </div>
                )}
                {success && (
                  <div className="p-3 text-sm text-emerald-400 bg-emerald-950/20 border border-emerald-900/50">
                    {success}
                  </div>
                )}
                {/* Form inputs: stacked on mobile, inline on desktop */}
                <div className="flex flex-col sm:flex-row gap-0">
                  {/* Domain select */}
                  <div className="relative flex items-center min-w-[140px] bg-input border border-border sm:border-r-0 h-[57px]">
                    <select
                      id="domain"
                      value={selectedDomain}
                      onChange={(e) => setSelectedDomain(e.target.value)}
                      required
                      className="w-full h-full appearance-none bg-transparent text-foreground pl-4 pr-10 focus:outline-none cursor-pointer text-sm"
                      disabled={submitting || domains.length === 0}
                      style={{ colorScheme: 'dark' }}
                    >
                      {domains.map((domain) => (
                        <option key={domain} value={domain}>
                          {domain}
                        </option>
                      ))}
                    </select>
                    <ChevronDown
                      size={14}
                      className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground pointer-events-none"
                    />
                  </div>
                  
                  {/* URL input */}
                  <input
                    id="url"
                    type="url"
                    value={url}
                    onChange={(e) => setUrl(e.target.value)}
                    required
                    placeholder="https://example.com/long-url"
                    className="flex-1 p-4 bg-input text-foreground border-y sm:border-y border-x sm:border-l border-t-0 sm:border-t border-border focus:outline-none focus:border-emerald-500 transition-colors placeholder:text-muted-foreground text-sm h-[57px]"
                    disabled={submitting}
                  />
                  
                  {/* Slug input */}
                  <input
                    id="slug"
                    type="text"
                    value={slug}
                    onChange={(e) => setSlug(e.target.value)}
                    placeholder="Custom slug (optional)"
                    className="flex-1 min-w-[120px] p-4 bg-input text-foreground border-y sm:border-y border-x sm:border-l border-t-0 sm:border-t border-border focus:outline-none focus:border-emerald-500 transition-colors placeholder:text-muted-foreground text-sm h-[57px]"
                    disabled={submitting}
                  />
                  
                  {/* Submit button */}
                  <Button 
                    type="submit" 
                    disabled={submitting}
                    className="bg-emerald-500 hover:bg-emerald-400 text-black px-6 h-[57px] font-bold flex items-center justify-center gap-2 transition-colors disabled:opacity-50 disabled:cursor-not-allowed whitespace-nowrap"
                  >
                    {submitting ? (
                      <>
                        <Scissors className="h-4 w-4 animate-pulse" />
                        <span className="hidden sm:inline">Processing</span>
                        <span className="sm:hidden">Processing</span>
                      </>
                    ) : (
                      <>
                        <Scissors className="h-4 w-4" />
                        <span className="hidden sm:inline">Shorten</span>
                        <span className="sm:hidden">Shorten Link</span>
                      </>
                    )}
                  </Button>
                </div>
              </form>
            </CardContent>
          </Card>

          {/* Section 2: Short URLs List */}
          <Card className="bg-white dark:bg-background border-0">
            <CardHeader>
              <CardTitle>Your Short URLs</CardTitle>
              <CardDescription>
                {total > 0 ? `Showing ${shortURLs.length} of ${total} shortened URLs` : "No shortened URLs yet"}
              </CardDescription>
            </CardHeader>
            <CardContent>
              {loading ? (
                <div className="text-center py-8 text-muted-foreground text-sm">Loading...</div>
              ) : shortURLs.length === 0 ? (
                <div className="text-center py-8 text-muted-foreground text-sm leading-relaxed">
                  No shortened URLs found. Create your first one above!
                </div>
              ) : (
                <>
                  <div className="overflow-x-auto">
                    <table className="w-full border-collapse">
                      <thead>
                        <tr className="border-b border-border">
                          <th className="text-left p-3 font-bold text-sm">Short URL</th>
                          <th className="text-left p-3 font-bold text-sm">Original URL</th>
                          <th className="text-left p-3 font-bold text-sm">Domain</th>
                          <th className="text-left p-3 font-bold text-sm">Slug</th>
                          <th className="text-left p-3 font-bold text-sm">Created</th>
                          <th className="text-left p-3 font-bold text-sm">Actions</th>
                        </tr>
                      </thead>
                      <tbody>
                        {shortURLs.map((shortURL) => (
                          <tr key={shortURL.id} className="border-b border-border hover:bg-muted/30 transition-colors">
                            <td className="p-3">
                              <a
                                href={`http://${shortURL.domain}/${shortURL.slug}`}
                                target="_blank"
                                rel="noopener noreferrer"
                                className="text-emerald-500 hover:text-emerald-400 hover:underline text-sm"
                              >
                                {shortURL.domain}/{shortURL.slug}
                              </a>
                            </td>
                            <td className="p-3">
                              <a
                                href={shortURL.url}
                                target="_blank"
                                rel="noopener noreferrer"
                                className="text-emerald-500 hover:text-emerald-400 hover:underline truncate max-w-md block text-sm"
                                title={shortURL.url}
                              >
                                {shortURL.url}
                              </a>
                            </td>
                            <td className="p-3 text-muted-foreground text-sm">{shortURL.domain}</td>
                            <td className="p-3 text-muted-foreground font-mono text-sm">{shortURL.slug}</td>
                            <td className="p-3 text-muted-foreground text-sm">{formatDate(shortURL.created_at)}</td>
                            <td className="p-3">
                              <Link to={`/short-urls/${shortURL.id}`}>
                                <Button variant="outline" size="icon" title="View More" className="h-8 w-8">
                                  <Eye className="h-3.5 w-3.5" />
                                </Button>
                              </Link>
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                  {totalPages > 1 && (
                    <div className="flex items-center justify-between mt-6 pt-4 border-t border-border">
                      <div className="text-xs text-muted-foreground uppercase tracking-wider">
                        Page {currentPage} of {totalPages}
                      </div>
                      <div className="flex gap-2">
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => setCurrentPage((p) => Math.max(1, p - 1))}
                          disabled={currentPage === 1 || loading}
                          className="text-xs uppercase tracking-wider"
                        >
                          <ChevronLeft className="mr-2 h-3.5 w-3.5" />
                          Previous
                        </Button>
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => setCurrentPage((p) => Math.min(totalPages, p + 1))}
                          disabled={currentPage === totalPages || loading}
                          className="text-xs uppercase tracking-wider"
                        >
                          Next
                          <ChevronRight className="ml-2 h-3.5 w-3.5" />
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

