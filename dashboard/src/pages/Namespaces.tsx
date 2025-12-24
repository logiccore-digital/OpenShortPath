import { useState, useEffect } from "react"
import { Folder, Plus, Edit, Trash2, ChevronLeft, ChevronRight, ChevronDown, X } from "lucide-react"
import { toast } from "sonner"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Navbar } from "@/components/Navbar"
import { getDomains, createNamespace, listNamespaces, updateNamespace, deleteNamespace } from "@/services/api"
import { Namespace } from "@/types/api"

export function Namespaces() {
  const [domains, setDomains] = useState<string[]>([])
  const [selectedDomain, setSelectedDomain] = useState<string>("")
  const [name, setName] = useState<string>("")
  const [namespaces, setNamespaces] = useState<Namespace[]>([])
  const [currentPage, setCurrentPage] = useState<number>(1)
  const [totalPages, setTotalPages] = useState<number>(1)
  const [total, setTotal] = useState<number>(0)
  const [loading, setLoading] = useState<boolean>(false)
  const [submitting, setSubmitting] = useState<boolean>(false)
  const [editingId, setEditingId] = useState<string | null>(null)
  const [editName, setEditName] = useState<string>("")
  const [editDomain, setEditDomain] = useState<string>("")
  const [deletingId, setDeletingId] = useState<string | null>(null)

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
        toast.error(err instanceof Error ? err.message : "Failed to load domains")
      }
    }
    loadDomains()
  }, [])

  // Load namespaces on mount and when page changes
  useEffect(() => {
    const loadNamespaces = async () => {
      setLoading(true)
      try {
        const response = await listNamespaces(currentPage, limit)
        setNamespaces(response.namespaces)
        setTotalPages(response.total_pages)
        setTotal(response.total)
      } catch (err) {
        toast.error(err instanceof Error ? err.message : "Failed to load namespaces")
      } finally {
        setLoading(false)
      }
    }
    loadNamespaces()
  }, [currentPage])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setSubmitting(true)

    try {
      await createNamespace(name, selectedDomain)
      toast.success("Namespace created successfully!")
      // Reset form
      setName("")
      // Refresh the list
      const response = await listNamespaces(currentPage, limit)
      setNamespaces(response.namespaces)
      setTotalPages(response.total_pages)
      setTotal(response.total)
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : "Failed to create namespace"
      toast.error(errorMessage)
    } finally {
      setSubmitting(false)
    }
  }

  const handleEdit = (namespace: Namespace) => {
    setEditingId(namespace.id)
    setEditName(namespace.name)
    setEditDomain(namespace.domain)
  }

  const handleCancelEdit = () => {
    setEditingId(null)
    setEditName("")
    setEditDomain("")
  }

  const handleSaveEdit = async () => {
    if (!editingId) return

    try {
      await updateNamespace(editingId, {
        name: editName,
        domain: editDomain,
      })
      toast.success("Namespace updated successfully!")
      handleCancelEdit()
      // Refresh the list
      const response = await listNamespaces(currentPage, limit)
      setNamespaces(response.namespaces)
      setTotalPages(response.total_pages)
      setTotal(response.total)
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : "Failed to update namespace"
      toast.error(errorMessage)
    }
  }

  const handleDeleteClick = (id: string) => {
    setDeletingId(id)
  }

  const handleDeleteCancel = () => {
    setDeletingId(null)
  }

  const handleDeleteConfirm = async () => {
    if (!deletingId) return

    try {
      await deleteNamespace(deletingId)
      toast.success("Namespace deleted successfully!")
      setDeletingId(null)
      // Refresh the list
      const response = await listNamespaces(currentPage, limit)
      setNamespaces(response.namespaces)
      setTotalPages(response.total_pages)
      setTotal(response.total)
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : "Failed to delete namespace"
      toast.error(errorMessage)
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

  const validateName = (name: string): boolean => {
    // Lowercase alphanumerical with optional hyphens or underscores, max 32 characters
    const regex = /^[a-z0-9_-]+$/
    return regex.test(name) && name.length > 0 && name.length <= 32
  }

  return (
    <div className="min-h-screen bg-background">
      <Navbar />
      <div className="px-6 md:px-8 lg:px-12 pb-6 md:pb-8 lg:pb-12 pt-24 transition-colors duration-300">
        <div className="max-w-7xl mx-auto w-full space-y-8">

          {/* Section 1: Create Namespace */}
          <Card>
            <CardHeader>
              <CardTitle>Create Namespace</CardTitle>
              <CardDescription>
                Create a new namespace to organize your short URLs. Namespace names must be lowercase alphanumerical with optional hyphens or underscores, and must be 32 characters or less.
              </CardDescription>
            </CardHeader>
            <CardContent>
              <form onSubmit={handleSubmit} className="space-y-2">
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
                  
                  {/* Name input */}
                  <input
                    id="name"
                    type="text"
                    value={name}
                    onChange={(e) => setName(e.target.value.toLowerCase())}
                    required
                    placeholder="my-namespace"
                    maxLength={32}
                    className="flex-1 p-4 bg-input text-foreground border-y sm:border-y border-x sm:border-l border-t-0 sm:border-t border-border focus:outline-none focus:border-emerald-500 transition-colors placeholder:text-muted-foreground text-sm h-[57px]"
                    disabled={submitting}
                    pattern="[a-z0-9_-]+"
                  />
                  
                  {/* Submit button */}
                  <Button 
                    type="submit" 
                    disabled={submitting || !validateName(name) || !selectedDomain}
                    className="bg-emerald-500 hover:bg-emerald-400 text-black px-6 h-[57px] font-bold flex items-center justify-center gap-2 transition-colors disabled:opacity-50 disabled:cursor-not-allowed whitespace-nowrap"
                  >
                    {submitting ? (
                      <>
                        <Plus className="h-4 w-4 animate-pulse" />
                        <span className="hidden sm:inline">Creating</span>
                        <span className="sm:hidden">Creating</span>
                      </>
                    ) : (
                      <>
                        <Plus className="h-4 w-4" />
                        <span className="hidden sm:inline">Create</span>
                        <span className="sm:hidden">Create</span>
                      </>
                    )}
                  </Button>
                </div>
                {name && !validateName(name) && (
                  <p className="text-xs text-red-500 mt-1">
                    Namespace name must be lowercase alphanumerical with optional hyphens or underscores, and must be 32 characters or less.
                  </p>
                )}
              </form>
            </CardContent>
          </Card>

          {/* Section 2: Namespaces List */}
          <Card className="bg-white dark:bg-background border-0">
            <CardHeader>
              <CardTitle>Your Namespaces</CardTitle>
              <CardDescription>
                {total > 0 ? `Showing ${namespaces.length} of ${total} namespaces` : "No namespaces yet"}
              </CardDescription>
            </CardHeader>
            <CardContent>
              {loading ? (
                <div className="text-center py-8 text-muted-foreground text-sm">Loading...</div>
              ) : namespaces.length === 0 ? (
                <div className="text-center py-8 text-muted-foreground text-sm leading-relaxed">
                  No namespaces found. Create your first one above!
                </div>
              ) : (
                <>
                  <div className="overflow-x-auto">
                    <table className="w-full border-collapse">
                      <thead>
                        <tr className="border-b border-border">
                          <th className="text-left p-3 font-bold text-sm">Name</th>
                          <th className="text-left p-3 font-bold text-sm">Domain</th>
                          <th className="text-left p-3 font-bold text-sm">Created</th>
                          <th className="text-left p-3 font-bold text-sm">Actions</th>
                        </tr>
                      </thead>
                      <tbody>
                        {namespaces.map((namespace) => (
                          <tr key={namespace.id} className="border-b border-border hover:bg-muted/30 transition-colors">
                            {editingId === namespace.id ? (
                              <>
                                <td className="p-3">
                                  <input
                                    type="text"
                                    value={editName}
                                    onChange={(e) => setEditName(e.target.value.toLowerCase())}
                                    maxLength={32}
                                    className="w-full p-2 bg-input text-foreground border border-border focus:outline-none focus:border-emerald-500 text-sm"
                                    pattern="[a-z0-9_-]+"
                                  />
                                </td>
                                <td className="p-3">
                                  <div className="relative">
                                    <select
                                      value={editDomain}
                                      onChange={(e) => setEditDomain(e.target.value)}
                                      className="w-full p-2 bg-input text-foreground border border-border focus:outline-none focus:border-emerald-500 appearance-none cursor-pointer text-sm pr-8"
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
                                      className="absolute right-2 top-1/2 -translate-y-1/2 text-muted-foreground pointer-events-none"
                                    />
                                  </div>
                                </td>
                                <td className="p-3 text-muted-foreground text-sm">{formatDate(namespace.created_at)}</td>
                                <td className="p-3">
                                  <div className="flex gap-2">
                                    <Button
                                      variant="outline"
                                      size="sm"
                                      onClick={handleSaveEdit}
                                      disabled={!validateName(editName) || !editDomain}
                                      className="h-8 text-xs"
                                    >
                                      Save
                                    </Button>
                                    <Button
                                      variant="outline"
                                      size="sm"
                                      onClick={handleCancelEdit}
                                      className="h-8 text-xs"
                                    >
                                      <X className="h-3.5 w-3.5" />
                                    </Button>
                                  </div>
                                </td>
                              </>
                            ) : (
                              <>
                                <td className="p-3">
                                  <div className="flex items-center gap-2">
                                    <Folder className="h-4 w-4 text-emerald-500" />
                                    <span className="font-mono text-sm">{namespace.name}</span>
                                  </div>
                                </td>
                                <td className="p-3 text-muted-foreground text-sm">{namespace.domain}</td>
                                <td className="p-3 text-muted-foreground text-sm">{formatDate(namespace.created_at)}</td>
                                <td className="p-3">
                                  <div className="flex gap-2">
                                    <Button
                                      variant="outline"
                                      size="icon"
                                      onClick={() => handleEdit(namespace)}
                                      title="Edit"
                                      className="h-8 w-8"
                                    >
                                      <Edit className="h-3.5 w-3.5" />
                                    </Button>
                                    <Button
                                      variant="outline"
                                      size="icon"
                                      onClick={() => handleDeleteClick(namespace.id)}
                                      title="Delete"
                                      className="h-8 w-8 text-red-500 hover:text-red-600 hover:border-red-500"
                                    >
                                      <Trash2 className="h-3.5 w-3.5" />
                                    </Button>
                                  </div>
                                </td>
                              </>
                            )}
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

      {/* Delete Confirmation Dialog */}
      {deletingId && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-[60]">
          <Card className="max-w-md w-full mx-4">
            <CardHeader>
              <CardTitle>Delete Namespace</CardTitle>
              <CardDescription>
                Are you sure you want to delete this namespace? This will also delete all short URLs associated with it. This action cannot be undone.
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="flex gap-3">
                <Button
                  onClick={handleDeleteConfirm}
                  variant="destructive"
                  className="flex-1"
                >
                  <Trash2 className="h-4 w-4 mr-2" />
                  Delete
                </Button>
                <Button
                  onClick={handleDeleteCancel}
                  variant="outline"
                  className="flex-1"
                >
                  Cancel
                </Button>
              </div>
            </CardContent>
          </Card>
        </div>
      )}
    </div>
  )
}

