package constants

// ContextKeyUserID is the key used to store the user ID in the Gin context
// This is set by the JWT middleware when a valid token with a 'sub' claim is provided
const ContextKeyUserID = "user_id"

// ContextKeyScopes is the key used to store the scopes in the Gin context
// This is set by the API key middleware when a valid API key is provided
const ContextKeyScopes = "scopes"

// ContextKeyAuthMethod is the key used to store the authentication method in the Gin context
// This is set by the JWT or API key middleware to indicate how the user authenticated
const ContextKeyAuthMethod = "auth_method"

// Authentication method values
const AuthMethodJWT = "jwt"
const AuthMethodAPIKey = "api_key"

// Plan types
const PlanHobbyist = "hobbyist"
const PlanVerifiedAccess = "verified_access"
const PlanPro = "pro"

// Rate limit types
const RateLimitTypeIP = "ip"
const RateLimitTypeUser = "user"

