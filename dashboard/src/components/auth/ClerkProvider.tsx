import { useEffect, useState, ReactNode, createContext, useContext } from "react"
import { ClerkProvider as ClerkProviderBase, useAuth } from "@clerk/clerk-react"
import { getAuthProvider } from "../../services/api"
import { setClerkTokenGetter } from "../../lib/auth"
import type { AuthProviderResponse } from "../../types/api"

interface ConditionalClerkProviderProps {
  children: ReactNode
}

// Context to track when Clerk is ready
interface ClerkReadyContextType {
  isClerkReady: boolean
  authProvider: "local" | "external_jwt" | "clerk" | null
  setClerkReady?: (ready: boolean) => void
}

const ClerkReadyContext = createContext<ClerkReadyContextType>({
  isClerkReady: false,
  authProvider: null,
})

export function useClerkReady() {
  return useContext(ClerkReadyContext)
}

/**
 * Inner component that sets up Clerk token getter
 * This component is only rendered when ClerkProvider is active, so useAuth() is safe
 */
function ClerkTokenSetup({ children, setClerkReady }: { children: ReactNode; setClerkReady: (ready: boolean) => void }) {
  const { getToken, isSignedIn, isLoaded } = useAuth()

  useEffect(() => {
    // Mark Clerk as ready once it's loaded
    if (isLoaded) {
      setClerkReady(true)
    }
  }, [isLoaded, setClerkReady])

  useEffect(() => {
    // Set the Clerk token getter function
    setClerkTokenGetter(async () => {
      if (isSignedIn && isLoaded) {
        try {
          // Get the session token - Clerk tokens are JWT format
          const token = await getToken()
          if (!token) {
            console.warn("Clerk getToken() returned null")
            return null
          }
          return token
        } catch (error) {
          console.error("Failed to get Clerk token:", error)
          return null
        }
      }
      return null
    })

    // Cleanup: remove token getter when component unmounts
    return () => {
      setClerkTokenGetter(null)
    }
  }, [getToken, isSignedIn, isLoaded])

  return <>{children}</>
}

/**
 * Conditional Clerk Provider that only initializes Clerk when auth_provider is "clerk"
 * Fetches the auth provider configuration and uses the publishable_key from the API
 */
export function ConditionalClerkProvider({ children }: ConditionalClerkProviderProps) {
  const [clerkPublishableKey, setClerkPublishableKey] = useState<string | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [authProvider, setAuthProvider] = useState<"local" | "external_jwt" | "clerk" | null>(null)
  const [isClerkReady, setIsClerkReady] = useState(false)

  useEffect(() => {
    // Fetch auth provider configuration
    getAuthProvider()
      .then((response: AuthProviderResponse) => {
        setAuthProvider(response.auth_provider)
        if (response.auth_provider === "clerk" && response.clerk_publishable_key) {
          setClerkPublishableKey(response.clerk_publishable_key)
          // Clerk will be ready after ClerkProvider is mounted
          // We'll set this in ClerkTokenSetup after ClerkProvider is active
        } else {
          // Clear Clerk token getter if not using Clerk
          setClerkTokenGetter(null)
          setIsClerkReady(true) // Not using Clerk, so we're "ready"
        }
        setIsLoading(false)
      })
      .catch((err) => {
        console.error("Failed to fetch auth provider:", err)
        setIsLoading(false)
        setIsClerkReady(true)
      })
  }, [])

  // Show loading state while fetching auth provider
  if (isLoading) {
    return (
      <ClerkReadyContext.Provider value={{ isClerkReady: false, authProvider: null, setClerkReady: undefined }}>
        {children}
      </ClerkReadyContext.Provider>
    )
  }

  // If auth provider is Clerk and we have a publishable key, wrap with ClerkProvider
  if (authProvider === "clerk" && clerkPublishableKey) {
    return (
      <ClerkProviderBase publishableKey={clerkPublishableKey}>
        <ClerkTokenSetup setClerkReady={setIsClerkReady}>
          <ClerkReadyContext.Provider value={{ isClerkReady: isClerkReady, authProvider: "clerk", setClerkReady: setIsClerkReady }}>
            {children}
          </ClerkReadyContext.Provider>
        </ClerkTokenSetup>
      </ClerkProviderBase>
    )
  }

  // Otherwise, render children without Clerk provider
  return (
    <ClerkReadyContext.Provider value={{ isClerkReady: true, authProvider, setClerkReady: undefined }}>
      {children}
    </ClerkReadyContext.Provider>
  )
}
