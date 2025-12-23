import { useState, useEffect } from "react"
import { useNavigate } from "react-router-dom"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { getAuthProvider, login } from "../services/api"
import { setStoredToken, isAuthenticated } from "../lib/auth"
import type { AuthProviderResponse } from "../types/api"

export function Login() {
  const navigate = useNavigate()
  const [authProvider, setAuthProvider] = useState<"local" | "external_jwt" | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [username, setUsername] = useState("")
  const [password, setPassword] = useState("")
  const [submitting, setSubmitting] = useState(false)

  useEffect(() => {
    // Redirect if already authenticated
    if (isAuthenticated()) {
      navigate("/", { replace: true })
      return
    }

    // Fetch auth provider on mount
    getAuthProvider()
      .then((response: AuthProviderResponse) => {
        setAuthProvider(response.auth_provider)
        setLoading(false)
      })
      .catch((err) => {
        setError(err.message || "Failed to fetch authentication provider")
        setLoading(false)
      })
  }, [navigate])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError(null)
    setSubmitting(true)

    try {
      const response = await login(username, password)
      setStoredToken(response.token)
      // Redirect to dashboard home
      navigate("/", { replace: true })
    } catch (err) {
      setError(err instanceof Error ? err.message : "Login failed")
      setSubmitting(false)
    }
  }

  if (loading) {
    return (
      <div className="min-h-screen bg-background flex items-center justify-center p-8">
        <Card className="w-full max-w-md">
          <CardContent className="pt-6">
            <p className="text-center text-muted-foreground">Loading...</p>
          </CardContent>
        </Card>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-background flex items-center justify-center p-8">
      <Card className="w-full max-w-md">
        <CardHeader>
          <CardTitle>Login</CardTitle>
          <CardDescription>
            {authProvider === "local"
              ? "Enter your credentials to access the dashboard"
              : "External authentication provider"}
          </CardDescription>
        </CardHeader>
        <CardContent>
          {authProvider === "local" ? (
            <form onSubmit={handleSubmit} className="space-y-4">
              {error && (
                <div className="p-3 text-sm text-red-600 bg-red-50 border border-red-200 rounded-md">
                  {error}
                </div>
              )}
              <div className="space-y-2">
                <label htmlFor="username" className="text-sm font-medium">
                  Username
                </label>
                <input
                  id="username"
                  type="text"
                  value={username}
                  onChange={(e) => setUsername(e.target.value)}
                  required
                  className="w-full px-3 py-2 border border-input bg-background rounded-md focus:outline-none focus:ring-2 focus:ring-ring"
                  disabled={submitting}
                />
              </div>
              <div className="space-y-2">
                <label htmlFor="password" className="text-sm font-medium">
                  Password
                </label>
                <input
                  id="password"
                  type="password"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  required
                  className="w-full px-3 py-2 border border-input bg-background rounded-md focus:outline-none focus:ring-2 focus:ring-ring"
                  disabled={submitting}
                />
              </div>
              <Button type="submit" className="w-full" disabled={submitting}>
                {submitting ? "Logging in..." : "Login"}
              </Button>
            </form>
          ) : (
            <div className="space-y-4">
              <p className="text-muted-foreground">
                You are using an external authentication provider. Please obtain a JWT token externally and set it in your browser's localStorage.
              </p>
              <p className="text-sm text-muted-foreground">
                To set the token manually, open your browser's developer console and run:
              </p>
              <pre className="p-3 bg-muted rounded-md text-xs overflow-x-auto">
                localStorage.setItem('jwt_token', 'your-token-here')
              </pre>
              <Button
                onClick={() => navigate("/", { replace: true })}
                variant="outline"
                className="w-full"
              >
                Go to Dashboard
              </Button>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}

