import { Navigate } from "react-router-dom"
import { useAuth } from "@clerk/clerk-react"
import { isAuthenticated } from "../lib/auth"
import { useEffect, useState } from "react"
import { getAuthProvider } from "../services/api"
import { useClerkReady } from "./auth/ClerkProvider"
import type { AuthProviderResponse } from "../types/api"

interface ProtectedRouteProps {
  children: React.ReactNode
}

/**
 * Clerk-protected route component
 * Only used when auth_provider is "clerk" and ClerkProvider is active
 */
function ClerkProtectedRoute({ children }: ProtectedRouteProps) {
  const { isSignedIn, isLoaded } = useAuth()

  if (!isLoaded) {
    return null
  }

  if (!isSignedIn) {
    return <Navigate to="/login" replace />
  }

  return <>{children}</>
}

/**
 * JWT-protected route component
 */
function JWTProtectedRoute({ children }: ProtectedRouteProps) {
  if (!isAuthenticated()) {
    return <Navigate to="/login" replace />
  }

  return <>{children}</>
}

/**
 * ProtectedRoute component that checks authentication status
 * Supports both Clerk and JWT authentication
 * Redirects to /login if user is not authenticated
 */
export function ProtectedRoute({ children }: ProtectedRouteProps) {
  const [authProvider, setAuthProvider] = useState<"local" | "external_jwt" | "clerk" | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const { isClerkReady, authProvider: contextAuthProvider } = useClerkReady()

  useEffect(() => {
    // Fetch auth provider to determine which auth method to check
    getAuthProvider()
      .then((response: AuthProviderResponse) => {
        setAuthProvider(response.auth_provider)
        setIsLoading(false)
      })
      .catch((err) => {
        console.error("Failed to fetch auth provider:", err)
        setIsLoading(false)
      })
  }, [])

  // Use context auth provider if available (more reliable)
  const effectiveAuthProvider = contextAuthProvider || authProvider

  // Wait for auth provider to load
  if (isLoading) {
    return null
  }

  // If using Clerk, wait for ClerkProvider to be ready
  if (effectiveAuthProvider === "clerk" && !isClerkReady) {
    return null
  }

  // Route to appropriate protected route component
  // When authProvider is "clerk", ClerkProvider should be active (checked via context)
  if (effectiveAuthProvider === "clerk") {
    return <ClerkProtectedRoute>{children}</ClerkProtectedRoute>
  } else {
    return <JWTProtectedRoute>{children}</JWTProtectedRoute>
  }
}

