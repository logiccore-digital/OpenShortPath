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
  namespace_id?: string
}

export interface ShortURL {
  id: string
  domain: string
  slug: string
  url: string
  user_id: string
  namespace_id?: string
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
  namespace_id?: string
}

export interface ApiError {
  error: string
  details?: string
}

export interface DomainsResponse {
  domains: string[]
}

export interface UserResponse {
  user_id: string
  username?: string
  active: boolean
  plan?: string
  monthly_link_limit?: number
  monthly_links_used?: number
  monthly_link_reset?: string
  rate_limit_per_hour?: number
  rate_limit_remaining?: number
  rate_limit_reset?: string
  created_at: string
  updated_at: string
}

export interface CreateAPIKeyRequest {
  scopes: string[]
}

export interface CreateAPIKeyResponse {
  id: string
  key: string
  scopes: string[]
  created_at: string
}

export interface APIKeyListItem {
  id: string
  scopes: string[]
  created_at: string
}

export interface ListAPIKeysResponse {
  keys: APIKeyListItem[]
}

export interface Namespace {
  id: string
  name: string
  domain: string
  user_id: string
  created_at: string
  updated_at: string
}

export interface CreateNamespaceRequest {
  name: string
  domain: string
}

export interface UpdateNamespaceRequest {
  name?: string
  domain?: string
}

export interface ListNamespacesResponse {
  namespaces: Namespace[]
  page: number
  limit: number
  total: number
  total_pages: number
}

