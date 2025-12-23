import { useState } from "react"
import { useNavigate, useLocation, Link } from "react-router-dom"
import { LayoutDashboard, User, LogOut } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { removeStoredToken } from "../lib/auth"

export function Navbar() {
  const navigate = useNavigate()
  const location = useLocation()
  const [showLogoutConfirm, setShowLogoutConfirm] = useState<boolean>(false)

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

  const isActive = (path: string) => {
    return location.pathname === path
  }

  return (
    <>
      <nav className="fixed top-0 left-0 right-0 z-50 border-b bg-background">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center h-16">
            <div className="flex items-center gap-6">
              <h1 className="text-xl font-semibold">OpenShortPath</h1>
              <div className="flex items-center gap-4">
                <Link
                  to="/"
                  className={`flex items-center gap-1.5 text-sm font-medium transition-colors hover:text-primary ${
                    isActive("/") ? "text-primary" : "text-muted-foreground"
                  }`}
                >
                  <LayoutDashboard className="h-4 w-4" />
                  Dashboard
                </Link>
                <Link
                  to="/account"
                  className={`flex items-center gap-1.5 text-sm font-medium transition-colors hover:text-primary ${
                    isActive("/account") ? "text-primary" : "text-muted-foreground"
                  }`}
                >
                  <User className="h-4 w-4" />
                  Account
                </Link>
              </div>
            </div>
            <div className="flex items-center">
              <Button variant="outline" onClick={handleLogoutClick}>
                <LogOut className="h-4 w-4 mr-2" />
                Logout
              </Button>
            </div>
          </div>
        </div>
      </nav>

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

