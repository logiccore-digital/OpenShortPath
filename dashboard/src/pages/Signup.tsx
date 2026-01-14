import { useState, useEffect } from "react"
import { useNavigate, Link } from "react-router-dom"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { getAuthProvider, signup } from "../services/api"
import { setStoredToken, isAuthenticated } from "../lib/auth"
import type { AuthProviderResponse } from "../types/api"

export function Signup() {
  const navigate = useNavigate()
  const [authProvider, setAuthProvider] = useState<"local" | "external_jwt" | null>(null)
  const [enableSignup, setEnableSignup] = useState(false)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [username, setUsername] = useState("")
  const [password, setPassword] = useState("")
  const [confirmPassword, setConfirmPassword] = useState("")
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
        setEnableSignup(response.enable_signup)
        setLoading(false)
        
        // Redirect if signup is not enabled or auth provider is not local
        if (response.auth_provider !== "local" || !response.enable_signup) {
          navigate("/login", { replace: true })
        }
      })
      .catch((err) => {
        setError(err.message || "Failed to fetch authentication provider")
        setLoading(false)
      })
  }, [navigate])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError(null)

    // Validate passwords match
    if (password !== confirmPassword) {
      setError("Passwords do not match")
      return
    }

    // Validate password length
    if (password.length < 8) {
      setError("Password must be at least 8 characters long")
      return
    }

    setSubmitting(true)

    try {
      const response = await signup(username, password)
      setStoredToken(response.token)
      // Redirect to dashboard home
      navigate("/", { replace: true })
    } catch (err) {
      setError(err instanceof Error ? err.message : "Signup failed")
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

  // Don't render if signup is not enabled
  if (authProvider !== "local" || !enableSignup) {
    return null
  }

  return (
    <div className="min-h-screen bg-background flex items-center justify-center p-8">
      <Card className="w-full max-w-md">
        <CardHeader>
          <CardTitle>Create Account</CardTitle>
          <CardDescription>
            Sign up for a new account to access the dashboard
          </CardDescription>
        </CardHeader>
        <CardContent>
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
                autoComplete="username"
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
                minLength={8}
                className="w-full px-3 py-2 border border-input bg-background rounded-md focus:outline-none focus:ring-2 focus:ring-ring"
                disabled={submitting}
                autoComplete="new-password"
              />
              <p className="text-xs text-muted-foreground">
                Password must be at least 8 characters long
              </p>
            </div>
            <div className="space-y-2">
              <label htmlFor="confirmPassword" className="text-sm font-medium">
                Confirm Password
              </label>
              <input
                id="confirmPassword"
                type="password"
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
                required
                minLength={8}
                className="w-full px-3 py-2 border border-input bg-background rounded-md focus:outline-none focus:ring-2 focus:ring-ring"
                disabled={submitting}
                autoComplete="new-password"
              />
            </div>
            <Button type="submit" className="w-full" disabled={submitting}>
              {submitting ? "Creating account..." : "Sign Up"}
            </Button>
            <div className="text-center text-sm">
              <span className="text-muted-foreground">Already have an account? </span>
              <Link
                to="/login"
                className="text-primary hover:underline"
              >
                Login
              </Link>
            </div>
          </form>
        </CardContent>
      </Card>
    </div>
  )
}
