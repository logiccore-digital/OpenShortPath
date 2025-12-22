package constants

// ContextKeyUserID is the key used to store the user ID in the Gin context
// This is set by the JWT middleware when a valid token with a 'sub' claim is provided
const ContextKeyUserID = "user_id"

