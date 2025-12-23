// API Type Definitions matching server response types

export interface AuthProviderResponse {
  auth_provider: "local" | "external_jwt"
}

export interface LoginRequest {
  username: string
  password: string
}

export interface LoginResponse {
  token: string
}

export interface ShortenRequest {
  domain: string
  url: string
  slug?: string
}

export interface ShortURL {
  id: string
  domain: string
  slug: string
  url: string
  user_id: string
  created_at: string
  updated_at: string
}

export interface ListShortURLsResponse {
  urls: ShortURL[]
  page: number
  limit: number
  total: number
  total_pages: number
}

export interface UpdateShortURLRequest {
  url?: string
  slug?: string
  domain?: string
}

export interface ApiError {
  error: string
  details?: string
}

export interface DomainsResponse {
  domains: string[]
}

