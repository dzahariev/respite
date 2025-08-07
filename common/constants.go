package common

// Define a custom type for context keys
type contextKey string

const (
	GLOBAL = "global"

	LoggerKey                 contextKey = "LoggerKey"
	RequestContextKey         contextKey = "RequestContextKey"
	CurrentUserIDKey          contextKey = "CurrentUserIDKey"
	CurrentUserPermissionsKey contextKey = "CurrentUserPermissionsKey"
)
