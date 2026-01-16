import { useState, useEffect } from "react"
import { useNavigate, Link } from "react-router-dom"
import { SignIn, useAuth } from "@clerk/clerk-react"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { getAuthProvider, login } from "../services/api"
import { setStoredToken, isAuthenticated } from "../lib/auth"
import { useClerkReady } from "../components/auth/ClerkProvider"
import { useTheme } from "../components/theme/ThemeProvider"
import type { AuthProviderResponse } from "../types/api"

/**
 * Clerk login component - only used when Clerk is enabled
 */
function ClerkLogin() {
  const navigate = useNavigate()
  const { isSignedIn, isLoaded } = useAuth()
  const { isDark } = useTheme()

  useEffect(() => {
    if (isLoaded && isSignedIn) {
      navigate("/", { replace: true })
    }
  }, [isLoaded, isSignedIn, navigate])

  if (!isLoaded) {
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
      <div className="w-full max-w-md">
        <SignIn
          routing="path"
          path="/dashboard/login"
          signUpUrl="/dashboard/signup"
          afterSignInUrl="/dashboard"
          appearance={{
            baseTheme: isDark ? "dark" : "light",
            variables: {
              colorBackground: isDark ? "hsl(0, 0%, 5%)" : "hsl(0, 0%, 100%)",
              colorInputBackground: isDark ? "hsl(0, 0%, 10%)" : "hsl(0, 0%, 100%)",
              colorText: isDark ? "hsl(0, 0%, 95%)" : "hsl(0, 0%, 9%)",
              colorTextSecondary: isDark ? "hsl(0, 0%, 75%)" : "hsl(0, 0%, 45%)",
              colorPrimary: isDark ? "hsl(142, 71%, 45%)" : "hsl(0, 0%, 9%)",
              colorInputText: isDark ? "hsl(0, 0%, 95%)" : "hsl(0, 0%, 9%)",
              colorDanger: isDark ? "hsl(0, 62.8%, 50%)" : "hsl(0, 84.2%, 60.2%)",
              colorSuccess: isDark ? "hsl(142, 71%, 45%)" : "hsl(142, 71%, 45%)",
              colorWarning: isDark ? "hsl(43, 74%, 66%)" : "hsl(43, 74%, 66%)",
              colorNeutral: isDark ? "hsl(0, 0%, 20%)" : "hsl(0, 0%, 90%)",
              borderRadius: "0.5rem",
              fontFamily: "'Open Sans', ui-sans-serif, system-ui, sans-serif",
            },
            elements: {
              rootBox: "mx-auto",
              card: "shadow-none bg-transparent",
              headerTitle: isDark 
                ? "text-white font-semibold" 
                : "text-[hsl(0,0%,9%)] font-semibold",
              headerSubtitle: isDark 
                ? "text-[hsl(0,0%,75%)]" 
                : "text-[hsl(0,0%,45%)]",
              dividerLine: isDark 
                ? "bg-[hsl(0,0%,20%)]" 
                : "bg-[hsl(0,0%,90%)]",
              dividerText: isDark 
                ? "text-[hsl(0,0%,60%)]" 
                : "text-[hsl(0,0%,45%)]",
              socialButtonsBlockButton: isDark 
                ? "bg-[hsl(0,0%,12%)] text-white border-[hsl(0,0%,25%)] hover:bg-[hsl(0,0%,18%)] hover:border-[hsl(0,0%,30%)] transition-colors" 
                : "bg-white text-[hsl(0,0%,9%)] border-[hsl(0,0%,90%)] hover:bg-[hsl(0,0%,96%)] transition-colors",
              socialButtonsBlockButtonText: isDark 
                ? "text-white font-medium" 
                : "text-[hsl(0,0%,9%)] font-medium",
              formButtonPrimary: isDark 
                ? "bg-[hsl(142,71%,45%)] text-[hsl(0,0%,5%)] hover:bg-[hsl(142,71%,50%)] font-medium transition-colors" 
                : "bg-[hsl(0,0%,9%)] text-white hover:bg-[hsl(0,0%,12%)] font-medium transition-colors",
              formFieldLabel: isDark 
                ? "text-[hsl(0,0%,90%)] font-medium" 
                : "text-[hsl(0,0%,9%)] font-medium",
              formFieldInput: isDark 
                ? "bg-[hsl(0,0%,10%)] text-white border-[hsl(0,0%,25%)] focus:border-[hsl(142,71%,45%)] focus:ring-2 focus:ring-[hsl(142,71%,45%)]/20" 
                : "bg-white text-[hsl(0,0%,9%)] border-[hsl(0,0%,90%)] focus:border-[hsl(0,0%,9%)] focus:ring-2 focus:ring-[hsl(0,0%,9%)]/20",
              formFieldInputShowPasswordButton: isDark 
                ? "text-[hsl(0,0%,70%)] hover:text-white" 
                : "text-[hsl(0,0%,45%)] hover:text-[hsl(0,0%,9%)]",
              formFieldErrorText: isDark 
                ? "text-[hsl(0,62.8%,60%)]" 
                : "text-[hsl(0,84.2%,60.2%)]",
              footerActionLink: isDark 
                ? "text-[hsl(142,71%,50%)] hover:text-[hsl(142,71%,55%)] font-medium" 
                : "text-[hsl(0,0%,9%)] hover:text-[hsl(0,0%,12%)] font-medium",
              footerActionText: isDark 
                ? "text-[hsl(0,0%,70%)]" 
                : "text-[hsl(0,0%,45%)]",
              identityPreviewText: isDark 
                ? "text-white" 
                : "text-[hsl(0,0%,9%)]",
              identityPreviewEditButton: isDark 
                ? "text-[hsl(142,71%,50%)] hover:text-[hsl(142,71%,55%)]" 
                : "text-[hsl(0,0%,9%)] hover:text-[hsl(0,0%,12%)]",
              formResendCodeLink: isDark 
                ? "text-[hsl(142,71%,50%)] hover:text-[hsl(142,71%,55%)]" 
                : "text-[hsl(0,0%,9%)] hover:text-[hsl(0,0%,12%)]",
              otpCodeFieldInput: isDark 
                ? "bg-[hsl(0,0%,10%)] text-white border-[hsl(0,0%,25%)] focus:border-[hsl(142,71%,45%)]" 
                : "bg-white text-[hsl(0,0%,9%)] border-[hsl(0,0%,90%)] focus:border-[hsl(0,0%,9%)]",
              alertText: isDark 
                ? "text-[hsl(0,0%,90%)]" 
                : "text-[hsl(0,0%,9%)]",
              formButtonReset: isDark 
                ? "text-[hsl(0,0%,70%)] hover:text-white" 
                : "text-[hsl(0,0%,45%)] hover:text-[hsl(0,0%,9%)]",
              badge: isDark 
                ? "bg-[hsl(0,0%,15%)] text-[hsl(0,0%,90%)] border-[hsl(0,0%,25%)]" 
                : "bg-[hsl(0,0%,96%)] text-[hsl(0,0%,9%)] border-[hsl(0,0%,90%)]",
            },
          } as any}
        />
      </div>
    </div>
  )
}

export function Login() {
  const navigate = useNavigate()
  const { isClerkReady, authProvider: contextAuthProvider } = useClerkReady()
  const [authProvider, setAuthProvider] = useState<"local" | "external_jwt" | "clerk" | null>(null)
  const [enableSignup, setEnableSignup] = useState(false)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [username, setUsername] = useState("")
  const [password, setPassword] = useState("")
  const [submitting, setSubmitting] = useState(false)

  useEffect(() => {
    // Fetch auth provider on mount
    getAuthProvider()
      .then((response: AuthProviderResponse) => {
        setAuthProvider(response.auth_provider)
        setEnableSignup(response.enable_signup)
        setLoading(false)
      })
      .catch((err) => {
        setError(err.message || "Failed to fetch authentication provider")
        setLoading(false)
      })
  }, [])

  // Use context auth provider if available (more reliable)
  const effectiveAuthProvider = contextAuthProvider || authProvider

  // Handle JWT authentication redirect
  useEffect(() => {
    if (effectiveAuthProvider !== "clerk" && isAuthenticated()) {
      navigate("/", { replace: true })
    }
  }, [effectiveAuthProvider, navigate])

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

  // Show Clerk sign-in UI when Clerk is enabled and ready
  if (effectiveAuthProvider === "clerk") {
    // Wait for Clerk to be ready before rendering ClerkLogin
    if (!isClerkReady) {
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
    return <ClerkLogin />
  }

  return (
    <div className="min-h-screen bg-background flex items-center justify-center p-8">
      <Card className="w-full max-w-md">
        <CardHeader>
          <CardTitle>Login</CardTitle>
          <CardDescription>
            {effectiveAuthProvider === "local"
              ? "Enter your credentials to access the dashboard"
              : "External authentication provider"}
          </CardDescription>
        </CardHeader>
        <CardContent>
          {effectiveAuthProvider === "local" ? (
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
              {enableSignup && (
                <div className="text-center text-sm">
                  <span className="text-muted-foreground">Don't have an account? </span>
                  <Link
                    to="/signup"
                    className="text-primary hover:underline"
                  >
                    Sign up
                  </Link>
                </div>
              )}
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

