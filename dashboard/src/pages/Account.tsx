import { useState, useEffect } from "react"
import { useNavigate } from "react-router-dom"
import { LogOut, Key, Trash2, Copy, Check, Plus } from "lucide-react"
import { toast } from "sonner"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Navbar } from "@/components/Navbar"
import { getMe, createAPIKey, listAPIKeys, deleteAPIKey } from "@/services/api"
import { UserResponse, APIKeyListItem, CreateAPIKeyResponse } from "@/types/api"
import { removeStoredToken } from "@/lib/auth"

const AVAILABLE_SCOPES = [
  { value: "shorten_url", label: "Shorten URL", description: "Create new short URLs" },
  { value: "read_urls", label: "Read URLs", description: "View and list short URLs" },
  { value: "write_urls", label: "Write URLs", description: "Update and delete short URLs" },
]

export function Account() {
  const navigate = useNavigate()
  const [user, setUser] = useState<UserResponse | null>(null)
  const [loading, setLoading] = useState<boolean>(true)
  const [error, setError] = useState<string>("")
  const [showLogoutConfirm, setShowLogoutConfirm] = useState<boolean>(false)
  
  // API Key management state
  const [apiKeys, setApiKeys] = useState<APIKeyListItem[]>([])
  const [loadingApiKeys, setLoadingApiKeys] = useState<boolean>(false)
  const [showCreateForm, setShowCreateForm] = useState<boolean>(false)
  const [selectedScopes, setSelectedScopes] = useState<string[]>([])
  const [creating, setCreating] = useState<boolean>(false)
  const [newApiKey, setNewApiKey] = useState<CreateAPIKeyResponse | null>(null)
  const [copied, setCopied] = useState<boolean>(false)
  const [deleteKeyId, setDeleteKeyId] = useState<string | null>(null)

  useEffect(() => {
    const loadUser = async () => {
      setLoading(true)
      setError("")
      try {
        const userData = await getMe()
        setUser(userData)
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to load account information")
      } finally {
        setLoading(false)
      }
    }
    loadUser()
  }, [])

  useEffect(() => {
    const loadAPIKeys = async () => {
      setLoadingApiKeys(true)
      try {
        const response = await listAPIKeys()
        setApiKeys(response.keys)
      } catch (err) {
        // Silently fail - API keys might not be available
        console.error("Failed to load API keys:", err)
      } finally {
        setLoadingApiKeys(false)
      }
    }
    loadAPIKeys()
  }, [])

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

  const handleLogoutClick = () => {
    setShowLogoutConfirm(true)
  }

  const handleLogoutCancel = () => {
    setShowLogoutConfirm(false)
  }

  const handleLogoutConfirm = () => {
    removeStoredToken()
    navigate("/login", { replace: true })
  }

  const handleScopeToggle = (scope: string) => {
    setSelectedScopes((prev) =>
      prev.includes(scope) ? prev.filter((s) => s !== scope) : [...prev, scope]
    )
  }

  const handleCreateAPIKey = async () => {
    if (selectedScopes.length === 0) {
      toast.error("Please select at least one scope")
      return
    }

    setCreating(true)
    try {
      const response = await createAPIKey(selectedScopes)
      setNewApiKey(response)
      setShowCreateForm(false)
      setSelectedScopes([])
      // Reload API keys list
      const updatedKeys = await listAPIKeys()
      setApiKeys(updatedKeys.keys)
      toast.success("API key created successfully")
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Failed to create API key")
    } finally {
      setCreating(false)
    }
  }

  const handleCopyKey = async () => {
    if (newApiKey?.key) {
      await navigator.clipboard.writeText(newApiKey.key)
      setCopied(true)
      toast.success("API key copied to clipboard")
      setTimeout(() => setCopied(false), 2000)
    }
  }

  const handleDeleteClick = (id: string) => {
    setDeleteKeyId(id)
  }

  const handleDeleteConfirm = async () => {
    if (!deleteKeyId) return

    try {
      await deleteAPIKey(deleteKeyId)
      setApiKeys((prev) => prev.filter((key) => key.id !== deleteKeyId))
      setDeleteKeyId(null)
      toast.success("API key deleted successfully")
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Failed to delete API key")
    }
  }

  const handleDeleteCancel = () => {
    setDeleteKeyId(null)
  }

  const handleCloseNewKeyDialog = () => {
    setNewApiKey(null)
    setCopied(false)
  }

  return (
    <>
      <div className="min-h-screen bg-background pt-20">
        <Navbar />
        <div className="p-8">
          <div className="max-w-4xl mx-auto space-y-8">
            <h1 className="text-4xl font-bold">Account</h1>

            {loading ? (
              <Card className="bg-white dark:bg-background">
                <CardContent className="pt-6">
                  <div className="text-center py-8 text-muted-foreground">Loading...</div>
                </CardContent>
              </Card>
            ) : error ? (
              <Card className="bg-white dark:bg-background">
                <CardContent className="pt-6">
                  <div className="text-center py-8">
                    <p className="text-red-600 mb-4">{error}</p>
                  </div>
                </CardContent>
              </Card>
            ) : user ? (
              <Card className="bg-white dark:bg-background">
                <CardHeader>
                  <CardTitle>Account Information</CardTitle>
                  <CardDescription>
                    Your account details and information
                  </CardDescription>
                </CardHeader>
                <CardContent className="space-y-6">
                  <div className="space-y-4">
                    <div>
                      <label className="block text-sm font-medium mb-1.5">User ID</label>
                      <div className="px-3 py-2 bg-muted rounded-md text-sm font-mono">
                        {user.user_id}
                      </div>
                    </div>

                    {user.username && (
                      <div>
                        <label className="block text-sm font-medium mb-1.5">Username</label>
                        <div className="px-3 py-2 bg-muted rounded-md text-sm">
                          {user.username}
                        </div>
                      </div>
                    )}

                    <div>
                      <label className="block text-sm font-medium mb-1.5">Status</label>
                      <div className="px-3 py-2 bg-muted rounded-md text-sm">
                        <span
                          className={`inline-flex items-center px-2 py-1 rounded-full text-xs font-medium ${
                            user.active
                              ? "bg-green-100 text-green-800"
                              : "bg-red-100 text-red-800"
                          }`}
                        >
                          {user.active ? "Active" : "Inactive"}
                        </span>
                      </div>
                    </div>

                    <div>
                      <label className="block text-sm font-medium mb-1.5">Created At</label>
                      <div className="px-3 py-2 bg-muted rounded-md text-sm">
                        {formatDate(user.created_at)}
                      </div>
                    </div>

                    <div>
                      <label className="block text-sm font-medium mb-1.5">Updated At</label>
                      <div className="px-3 py-2 bg-muted rounded-md text-sm">
                        {formatDate(user.updated_at)}
                      </div>
                    </div>
                  </div>
                  <div className="pt-4 border-t">
                    <Button onClick={handleLogoutClick} variant="destructive">
                      <LogOut className="h-4 w-4 mr-2" />
                      Logout
                    </Button>
                  </div>
                </CardContent>
              </Card>
            ) : null}

            {/* API Keys Management Section */}
            <Card className="bg-white dark:bg-background">
              <CardHeader>
                <div className="flex items-center justify-between">
                  <div>
                    <CardTitle className="flex items-center gap-2">
                      <Key className="h-5 w-5" />
                      API Keys
                    </CardTitle>
                    <CardDescription>
                      Manage your API keys for programmatic access
                    </CardDescription>
                  </div>
                  {!showCreateForm && (
                    <Button
                      onClick={() => setShowCreateForm(true)}
                      variant="outline"
                      size="sm"
                    >
                      <Plus className="h-4 w-4 mr-2" />
                      Create Key
                    </Button>
                  )}
                </div>
              </CardHeader>
              <CardContent className="space-y-4">
                {showCreateForm && (
                  <div className="p-4 border border-border rounded-md space-y-4">
                    <div>
                      <label className="block text-sm font-medium mb-2">
                        Select Scopes (at least one required)
                      </label>
                      <div className="space-y-2">
                        {AVAILABLE_SCOPES.map((scope) => (
                          <label
                            key={scope.value}
                            className="flex items-start gap-3 p-3 border border-border rounded-md hover:bg-muted/50 cursor-pointer transition-colors"
                          >
                            <input
                              type="checkbox"
                              checked={selectedScopes.includes(scope.value)}
                              onChange={() => handleScopeToggle(scope.value)}
                              className="mt-1 h-4 w-4"
                            />
                            <div className="flex-1">
                              <div className="text-sm font-medium">{scope.label}</div>
                              <div className="text-xs text-muted-foreground">
                                {scope.description}
                              </div>
                            </div>
                          </label>
                        ))}
                      </div>
                    </div>
                    <div className="flex gap-2">
                      <Button
                        onClick={handleCreateAPIKey}
                        disabled={creating || selectedScopes.length === 0}
                        className="flex-1"
                      >
                        {creating ? "Creating..." : "Create API Key"}
                      </Button>
                      <Button
                        onClick={() => {
                          setShowCreateForm(false)
                          setSelectedScopes([])
                        }}
                        variant="outline"
                      >
                        Cancel
                      </Button>
                    </div>
                  </div>
                )}

                {loadingApiKeys ? (
                  <div className="text-center py-8 text-muted-foreground">Loading API keys...</div>
                ) : apiKeys.length === 0 ? (
                  <div className="text-center py-8 text-muted-foreground">
                    No API keys found. Create one to get started.
                  </div>
                ) : (
                  <div className="space-y-3">
                    {apiKeys.map((key) => (
                      <div
                        key={key.id}
                        className="p-4 border border-border rounded-md flex items-center justify-between"
                      >
                        <div className="flex-1">
                          <div className="flex items-center gap-2 mb-1">
                            <span className="text-sm font-mono text-muted-foreground">
                              {key.id.substring(0, 8)}...
                            </span>
                            <span className="text-xs text-muted-foreground">
                              Created {formatDate(key.created_at)}
                            </span>
                          </div>
                          <div className="flex flex-wrap gap-1 mt-2">
                            {key.scopes.map((scope) => {
                              const scopeInfo = AVAILABLE_SCOPES.find((s) => s.value === scope)
                              return (
                                <span
                                  key={scope}
                                  className="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200"
                                >
                                  {scopeInfo?.label || scope}
                                </span>
                              )
                            })}
                          </div>
                        </div>
                        <Button
                          onClick={() => handleDeleteClick(key.id)}
                          variant="outline"
                          size="sm"
                          className="ml-4"
                        >
                          <Trash2 className="h-4 w-4" />
                        </Button>
                      </div>
                    ))}
                  </div>
                )}
              </CardContent>
            </Card>
          </div>
        </div>
      </div>

      {/* New API Key Display Dialog */}
      {newApiKey && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-[60]">
          <Card className="max-w-2xl w-full mx-4">
            <CardHeader>
              <CardTitle>API Key Created</CardTitle>
              <CardDescription>
                Your API key has been created. Copy it now - you won't be able to see it again!
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div>
                <label className="block text-sm font-medium mb-2">API Key</label>
                <div className="flex gap-2">
                  <div className="flex-1 px-3 py-2 bg-muted rounded-md text-sm font-mono break-all">
                    {newApiKey.key}
                  </div>
                  <Button
                    onClick={handleCopyKey}
                    variant="outline"
                    size="sm"
                    className="shrink-0"
                  >
                    {copied ? (
                      <>
                        <Check className="h-4 w-4 mr-2" />
                        Copied
                      </>
                    ) : (
                      <>
                        <Copy className="h-4 w-4 mr-2" />
                        Copy
                      </>
                    )}
                  </Button>
                </div>
              </div>
              <div>
                <label className="block text-sm font-medium mb-2">Scopes</label>
                <div className="flex flex-wrap gap-2">
                  {newApiKey.scopes.map((scope) => {
                    const scopeInfo = AVAILABLE_SCOPES.find((s) => s.value === scope)
                    return (
                      <span
                        key={scope}
                        className="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200"
                      >
                        {scopeInfo?.label || scope}
                      </span>
                    )
                  })}
                </div>
              </div>
              <div className="pt-4 border-t">
                <Button onClick={handleCloseNewKeyDialog} className="w-full">
                  I've copied the key
                </Button>
              </div>
            </CardContent>
          </Card>
        </div>
      )}

      {/* Delete API Key Confirmation Dialog */}
      {deleteKeyId && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-[60]">
          <Card className="max-w-md w-full mx-4">
            <CardHeader>
              <CardTitle>Delete API Key</CardTitle>
              <CardDescription>
                Are you sure you want to delete this API key? This action cannot be undone.
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

      {/* Logout Confirmation Dialog */}
      {showLogoutConfirm && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-[60]">
          <Card className="max-w-md w-full mx-4">
            <CardHeader>
              <CardTitle>Logout</CardTitle>
              <CardDescription>
                Are you sure you want to logout? You will need to login again to access the dashboard.
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="flex gap-3">
                <Button
                  onClick={handleLogoutConfirm}
                  variant="destructive"
                  className="flex-1"
                >
                  <LogOut className="h-4 w-4 mr-2" />
                  Logout
                </Button>
                <Button
                  onClick={handleLogoutCancel}
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
    </>
  )
}

