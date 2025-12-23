import { useNavigate } from "react-router-dom"
import { Button } from "@/components/ui/button"
import { removeStoredToken } from "../lib/auth"

export function Navbar() {
  const navigate = useNavigate()

  const handleLogout = () => {
    removeStoredToken()
    navigate("/login", { replace: true })
  }

  return (
    <nav className="border-b bg-background">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex justify-between items-center h-16">
          <div className="flex items-center">
            <h1 className="text-xl font-semibold">OpenShortPath</h1>
          </div>
          <div className="flex items-center">
            <Button variant="outline" onClick={handleLogout}>
              Logout
            </Button>
          </div>
        </div>
      </div>
    </nav>
  )
}

