import { jwtDecode } from "jwt-decode"

const TOKEN_STORAGE_KEY = "jwt_token"

interface JWTPayload {
  sub?: string
  exp?: number
  iat?: number
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
 */
export function isAuthenticated(): boolean {
  const token = getStoredToken()
  
  if (!token) {
    return false
  }
  
  return !isTokenExpired(token)
}

