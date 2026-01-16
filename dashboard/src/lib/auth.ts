import { jwtDecode } from "jwt-decode"

const TOKEN_STORAGE_KEY = "jwt_token"

interface JWTPayload {
  sub?: string
  exp?: number
  iat?: number
}

// Global function to get Clerk token - will be set by ClerkProvider when Clerk is enabled
let getClerkTokenFn: (() => Promise<string | null>) | null = null

/**
 * Set the function to get Clerk token (called by ClerkProvider)
 */
export function setClerkTokenGetter(fn: (() => Promise<string | null>) | null): void {
  getClerkTokenFn = fn
}

/**
 * Get Clerk token if available
 */
export async function getClerkToken(): Promise<string | null> {
  if (getClerkTokenFn) {
    return getClerkTokenFn()
  }
  return null
}

/**
 * Retrieve JWT token from localStorage
 */
export function getStoredToken(): string | null {
  return localStorage.getItem(TOKEN_STORAGE_KEY)
}

/**
 * Store JWT token in localStorage
 */
export function setStoredToken(token: string): void {
  localStorage.setItem(TOKEN_STORAGE_KEY, token)
}

/**
 * Remove JWT token from localStorage
 */
export function removeStoredToken(): void {
  localStorage.removeItem(TOKEN_STORAGE_KEY)
}

/**
 * Check if a JWT token is expired
 * Decodes the token without verification and checks the 'exp' claim
 */
export function isTokenExpired(token: string): boolean {
  try {
    const decoded = jwtDecode<JWTPayload>(token)
    
    // If no exp claim, consider it not expired (server may not use exp)
    if (!decoded.exp) {
      return false
    }
    
    // Check if expiration time is in the past
    const currentTime = Math.floor(Date.now() / 1000)
    return decoded.exp < currentTime
  } catch (error) {
    // If token can't be decoded, consider it expired/invalid
    return true
  }
}

/**
 * Check if user is authenticated
 * Returns true if a valid token exists and is not expired
 * Note: For Clerk authentication, this should be checked via Clerk's useAuth hook
 */
export function isAuthenticated(): boolean {
  const token = getStoredToken()
  
  if (!token) {
    return false
  }
  
  return !isTokenExpired(token)
}

/**
 * Get the current authentication token (Clerk or JWT)
 * Returns Clerk token if available, otherwise falls back to stored JWT token
 */
export async function getAuthToken(): Promise<string | null> {
  // Try to get Clerk token first
  try {
    const clerkToken = await getClerkToken()
    if (clerkToken) {
      return clerkToken
    }
  } catch (error) {
    console.error("Error getting Clerk token:", error)
  }
  
  // Fall back to stored JWT token
  return getStoredToken()
}

