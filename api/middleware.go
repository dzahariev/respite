package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/dzahariev/respite/common"

	"github.com/gofrs/uuid/v5"
)

// Static is a Wrapper for static resources
func (server *Server) Static() http.Handler {
	return http.FileServer(http.Dir("./public"))
}

// Health is a Wrapper for static resources
func (server *Server) Health() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "OK")
	}
}

// Public is a Wrapper for public resources
func (server *Server) Public(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		next(w, r)
	}
}

// Protected is a Wrapper for protected and Global resources
func (server *Server) Protected(permission string, resource common.Resource, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := common.GetLogger(ctx)

		// Parse token
		authHeader := r.Header.Get("Authorization")
		if len(authHeader) < 7 {
			logger.Error("Unauthorized request, missing or invalid Authorization header")
			ERROR(w, http.StatusUnauthorized, fmt.Errorf("unauthorized, missing bearer authorization header"))
			return
		}
		authType := strings.ToLower(authHeader[:6])
		if authType != "bearer" {
			logger.Error("Unauthorized request, invalid Authorization header type", "type", authType)
			ERROR(w, http.StatusUnauthorized, fmt.Errorf("unauthorized, invalid bearer authorization header"))
			return
		}
		// Verify token is valid
		tokenString := authHeader[7:]
		tokenString = strings.TrimSpace(tokenString)
		err := server.AuthClient.RetrospectToken(ctx, tokenString)
		if err != nil {
			logger.Error("Unauthorized request, invalid token", "error", err)
			ERROR(w, http.StatusUnauthorized, err)
			return
		}
		// Create user if not exists
		userFromInfo, err := server.AuthClient.GetUserFromToken(ctx, tokenString)
		if err != nil {
			logger.Error("Unauthorized request, cannot get user from token", "error", err)
			ERROR(w, http.StatusUnauthorized, err)
			return
		}
		loadedUser, _ := server.DBLoadUser(ctx, string(userFromInfo.ID.String())) // we ignore the error as it is expected if user do not exists
		if loadedUser == nil {
			err := server.DBSaveUser(ctx, userFromInfo)
			if err != nil {
				logger.Error("Error saving user from token", "error", err)
				ERROR(w, http.StatusUnauthorized, err)
				return
			}
		}
		loadedUser, err = server.DBLoadUser(ctx, string(userFromInfo.ID.String()))
		if err != nil {
			logger.Error("Error loading user from token", "error", err)
			ERROR(w, http.StatusUnauthorized, err)
			return
		}

		// Create new context with current user
		ctxWithUser := context.WithValue(ctx, common.CurrentUserKey, loadedUser)
		// Get roles from token
		roles, err := server.AuthClient.GetRolesFromToken(ctxWithUser, tokenString)
		if err != nil {
			logger.Error("Unauthorized request, cannot get roles from token", "error", err)
			ERROR(w, http.StatusUnauthorized, err)
			return
		}
		var permissions []string
		for _, role := range roles {
			permissions = append(permissions, server.RoleToPermissions[role]...)
		}
		// Create new context with current user permissions
		ctxWithUserPerm := context.WithValue(ctxWithUser, common.CurrentUserPermissionsKey, permissions)

		// Replace request context
		rWithUserPerm := r.WithContext(ctxWithUserPerm)

		requestContext := common.NewRequestContext(rWithUserPerm, server.DB, resource, server.Resources)
		ctxWithUserPermRC := context.WithValue(ctxWithUserPerm, common.RequestContextKey, requestContext)

		// Replace request context
		rWithUserPermRC := r.WithContext(ctxWithUserPermRC)

		// Check permissions
		if havePermission(resource.Name, permission, permissions) {
			next(w, rWithUserPermRC)
		} else {
			// lack of permissions
			logger.Error("Unauthorized request, no permission for resource", "resource", resource.Name, "permission", permission)
			ERROR(w, http.StatusUnauthorized, fmt.Errorf("unauthorized, no permission for %s.%s", resource.Name, permission))
			return
		}
	}
}

// Middleware to add request_id logger into context
func loggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := uuid.Must(uuid.NewV4()).String()
		logger := slog.Default().With("request_id", reqID)
		ctx := context.WithValue(r.Context(), common.LoggerKey, logger)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ContentTypeJSON set the content type to JSON
func ContentTypeJSON(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next(w, r)
	}
}

// JSON returns data as JSON stream
func JSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.WriteHeader(statusCode)
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		fmt.Fprintf(w, "%s", err.Error())
	}
}

// ERROR returns error as JSON representation
func ERROR(w http.ResponseWriter, statusCode int, err error) {
	if err != nil {
		JSON(w, statusCode, struct {
			Error string `json:"error"`
		}{
			Error: err.Error(),
		})
		return
	}
	JSON(w, http.StatusBadRequest, nil)
}

// havePermission is to check if the permission for the resource is present in the list of permissions
func havePermission(resource, permission string, permissions []string) bool {
	for _, currentPermission := range permissions {
		resourcePermission := fmt.Sprintf("%s.%s", resource, permission)
		if strings.EqualFold(currentPermission, resourcePermission) {
			return true
		}
	}
	return false
}
