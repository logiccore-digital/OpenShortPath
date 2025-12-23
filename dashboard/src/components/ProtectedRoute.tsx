import { Navigate } from "react-router-dom"
import { isAuthenticated } from "../lib/auth"

interface ProtectedRouteProps {
  children: React.ReactNode
}

/**
 * ProtectedRoute component that checks authentication status
 * Redirects to /login if user is not authenticated
 */
export function ProtectedRoute({ children }: ProtectedRouteProps) {
  if (!isAuthenticated()) {
    return <Navigate to="/login" replace />
  }
  
  return <>{children}</>
}

