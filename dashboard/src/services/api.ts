import {
  AuthProviderResponse,
  LoginRequest,
  LoginResponse,
  ShortenRequest,
  ShortURL,
  ListShortURLsResponse,
  UpdateShortURLRequest,
  ApiError,
  DomainsResponse,
  UserResponse,
  CreateAPIKeyRequest,
  CreateAPIKeyResponse,
  ListAPIKeysResponse,
} from "../types/api"
import { getStoredToken, removeStoredToken } from "../lib/auth"

const API_BASE_URL = "/api/v1"

/**
 * Base API client with automatic token attachment and error handling
 */
async function apiRequest<T>(
  endpoint: string,
  options: RequestInit = {}
): Promise<T> {
  const token = getStoredToken()
  
  // Build headers object
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
  }
  
  // Merge existing headers if they're a plain object
  if (options.headers) {
    if (options.headers instanceof Headers) {
      options.headers.forEach((value, key) => {
        headers[key] = value
      })
    } else if (Array.isArray(options.headers)) {
      options.headers.forEach(([key, value]) => {
        headers[key] = value
      })
    } else {
      Object.assign(headers, options.headers)
    }
  }
  
  // Attach Bearer token if available
  if (token) {
    headers["Authorization"] = `Bearer ${token}`
  }
  
  const response = await fetch(`${API_BASE_URL}${endpoint}`, {
    ...options,
    headers,
  })
  
  // Handle 401 Unauthorized - clear token and redirect to login
  if (response.status === 401) {
    removeStoredToken()
    window.location.href = "/dashboard/login"
    throw new Error("Unauthorized")
  }
  
  // Handle other error statuses
  if (!response.ok) {
    let errorMessage = `API request failed: ${response.status} ${response.statusText}`
    
    try {
      const errorData: ApiError = await response.json()
      errorMessage = errorData.error || errorMessage
      if (errorData.details) {
        errorMessage += ` - ${errorData.details}`
      }
    } catch {
      // If response is not JSON, use default error message
    }
    
    throw new Error(errorMessage)
  }
  
  // Handle 204 No Content (for DELETE requests)
  if (response.status === 204) {
    return undefined as T
  }
  
  return response.json()
}

/**
 * Get authentication provider type
 */
export async function getAuthProvider(): Promise<AuthProviderResponse> {
  return apiRequest<AuthProviderResponse>("/auth-provider")
}

/**
 * Get list of available domains for shortening
 */
export async function getDomains(): Promise<string[]> {
  const response = await apiRequest<DomainsResponse>("/domains")
  return response.domains
}

/**
 * Login with username and password (local auth only)
 */
export async function login(username: string, password: string): Promise<LoginResponse> {
  const request: LoginRequest = { username, password }
  return apiRequest<LoginResponse>("/login", {
    method: "POST",
    body: JSON.stringify(request),
  })
}

/**
 * Create a short URL
 */
export async function shorten(
  domain: string,
  url: string,
  slug?: string
): Promise<ShortURL> {
  const request: ShortenRequest = { domain, url, slug }
  return apiRequest<ShortURL>("/shorten", {
    method: "POST",
    body: JSON.stringify(request),
  })
}

/**
 * List user's short URLs with pagination
 */
export async function listShortURLs(
  page: number = 1,
  limit: number = 20
): Promise<ListShortURLsResponse> {
  const params = new URLSearchParams({
    page: page.toString(),
    limit: limit.toString(),
  })
  return apiRequest<ListShortURLsResponse>(`/short-urls?${params.toString()}`)
}

/**
 * Get a single short URL by ID
 */
export async function getShortURL(id: string): Promise<ShortURL> {
  return apiRequest<ShortURL>(`/short-urls/${id}`)
}

/**
 * Update a short URL by ID
 */
export async function updateShortURL(
  id: string,
  updates: UpdateShortURLRequest
): Promise<ShortURL> {
  return apiRequest<ShortURL>(`/short-urls/${id}`, {
    method: "PUT",
    body: JSON.stringify(updates),
  })
}

/**
 * Delete a short URL by ID
 */
export async function deleteShortURL(id: string): Promise<void> {
  return apiRequest<void>(`/short-urls/${id}`, {
    method: "DELETE",
  })
}

/**
 * Get current authenticated user's information
 */
export async function getMe(): Promise<UserResponse> {
  return apiRequest<UserResponse>("/me")
}

/**
 * Create a new API key
 */
export async function createAPIKey(scopes: string[]): Promise<CreateAPIKeyResponse> {
  const request: CreateAPIKeyRequest = { scopes }
  return apiRequest<CreateAPIKeyResponse>("/api-keys", {
    method: "POST",
    body: JSON.stringify(request),
  })
}

/**
 * List user's API keys
 */
export async function listAPIKeys(): Promise<ListAPIKeysResponse> {
  return apiRequest<ListAPIKeysResponse>("/api-keys")
}

/**
 * Delete an API key by ID
 */
export async function deleteAPIKey(id: string): Promise<void> {
  return apiRequest<void>(`/api-keys/${id}`, {
    method: "DELETE",
  })
}

