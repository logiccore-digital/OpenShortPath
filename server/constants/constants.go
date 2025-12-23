package constants

// ContextKeyUserID is the key used to store the user ID in the Gin context
// This is set by the JWT middleware when a valid token with a 'sub' claim is provided
const ContextKeyUserID = "user_id"

// ContextKeyScopes is the key used to store the scopes in the Gin context
// This is set by the API key middleware when a valid API key is provided
const ContextKeyScopes = "scopes"

