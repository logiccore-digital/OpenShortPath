import { useState, useEffect } from "react"
import { useNavigate } from "react-router-dom"
import { LogOut } from "lucide-react"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Navbar } from "@/components/Navbar"
import { getMe } from "@/services/api"
import { UserResponse } from "@/types/api"
import { removeStoredToken } from "@/lib/auth"

export function Account() {
  const navigate = useNavigate()
  const [user, setUser] = useState<UserResponse | null>(null)
  const [loading, setLoading] = useState<boolean>(true)
  const [error, setError] = useState<string>("")
  const [showLogoutConfirm, setShowLogoutConfirm] = useState<boolean>(false)

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
          </div>
        </div>
      </div>

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

