import { useState } from "react"
import { useNavigate, useLocation, Link } from "react-router-dom"
import { Terminal, LayoutDashboard, User, LogOut, Folder } from "lucide-react"
import { toast } from "sonner"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { removeStoredToken } from "../lib/auth"
import { useTheme } from "./theme/ThemeProvider"

export function Navbar() {
  const navigate = useNavigate()
  const location = useLocation()
  const { isDark, toggleTheme } = useTheme()
  const [showLogoutConfirm, setShowLogoutConfirm] = useState<boolean>(false)

  const handleLogoutClick = () => {
    setShowLogoutConfirm(true)
  }

  const handleLogoutCancel = () => {
    setShowLogoutConfirm(false)
  }

  const handleLogoutConfirm = () => {
    removeStoredToken()
    toast.success("Logged out successfully!")
    navigate("/login", { replace: true })
  }

  const isActive = (path: string) => {
    return location.pathname === path
  }

  return (
    <>
      <nav className="fixed top-0 left-0 right-0 z-50 border-b border-border bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
        <div className="max-w-full mx-auto px-6 md:px-8 lg:px-12">
          <div className="flex justify-between items-center h-16 transition-colors duration-300">
            <h1 className="hidden sm:flex text-2xl font-bold tracking-tighter items-center gap-2">
              <Terminal size={20} className="text-emerald-500" />
              OpenShortPath_
            </h1>
            <div className="flex items-center gap-4">
              <button
                onClick={toggleTheme}
                className="text-xs uppercase tracking-wider text-muted-foreground hover:text-foreground transition-colors"
              >
                [{isDark ? 'light_mode' : 'dark_mode'}]
              </button>
              <Link
                to="/"
                className={`flex items-center gap-1.5 text-xs uppercase tracking-wider transition-colors ${
                  isActive("/") ? "text-foreground" : "text-muted-foreground hover:text-foreground"
                }`}
              >
                <LayoutDashboard className="h-3.5 w-3.5" />
                <span className="hidden sm:inline">Dashboard</span>
              </Link>
              <Link
                to="/namespaces"
                className={`flex items-center gap-1.5 text-xs uppercase tracking-wider transition-colors ${
                  isActive("/namespaces") ? "text-foreground" : "text-muted-foreground hover:text-foreground"
                }`}
              >
                <Folder className="h-3.5 w-3.5" />
                <span className="hidden sm:inline">Namespaces</span>
              </Link>
              <Link
                to="/account"
                className={`flex items-center gap-1.5 text-xs uppercase tracking-wider transition-colors ${
                  isActive("/account") ? "text-foreground" : "text-muted-foreground hover:text-foreground"
                }`}
              >
                <User className="h-3.5 w-3.5" />
                <span className="hidden sm:inline">Account</span>
              </Link>
              <Button 
                variant="outline" 
                onClick={handleLogoutClick}
                className="text-xs uppercase tracking-wider h-auto py-1.5 px-3"
              >
                <LogOut className="h-3.5 w-3.5 mr-1.5" />
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

