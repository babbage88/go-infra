package authapi

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// ContextKey type to avoid collisions
type ContextKey string

const ClaimsContextKey ContextKey = "jwtClaims"
const userIDContextKey = ContextKey("userID")

// AuthMiddleware extracts and validates the JWT and stores claims in context
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		claims, err := parseAndValidateToken(r)
		if err != nil {
			slog.Error("Unauthorized request", slog.String("error", err.Error()))
			http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), ClaimsContextKey, claims)
		slog.Info("Token has been verified.", slog.String("Path", r.URL.Path))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// AuthMiddlewareRequirePermission adds permission check on top of AuthMiddleware
func AuthMiddlewareRequirePermission(ua AuthService, permissionName string, next http.Handler) http.Handler {
	return AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		claimsVal := r.Context().Value(ClaimsContextKey)
		claims, ok := claimsVal.(jwt.MapClaims)
		if !ok {
			http.Error(w, `{"error": "Invalid token claims"}`, http.StatusUnauthorized)
			return
		}

		roleIDsInterface, ok := claims["role_ids"].([]interface{})
		if !ok {
			http.Error(w, `{"error": "Role IDs missing or invalid"}`, http.StatusUnauthorized)
			return
		}

		var roleIDs []uuid.UUID
		for _, roleID := range roleIDsInterface {
			roleIDStr, ok := roleID.(string)
			if !ok {
				http.Error(w, `{"error": "Invalid role ID format"}`, http.StatusBadRequest)
				return
			}

			parsedUUID, err := uuid.Parse(roleIDStr)
			if err != nil {
				http.Error(w, `{"error": "Invalid UUID format"}`, http.StatusBadRequest)
				return
			}
			roleIDs = append(roleIDs, parsedUUID)
		}

		slog.Info("Checking permission for role IDs",
			slog.Any("roleIDs", roleIDs),
			slog.String("permission", permissionName))

		hasPermission, err := ua.VerifyUserRolesForPermission(roleIDs, permissionName)
		if err != nil {
			slog.Error("Permission check failed", slog.String("error", err.Error()))
			http.Error(w, `{"error": "Internal Server Error"}`, http.StatusInternalServerError)
			return
		}

		if !hasPermission {
			http.Error(w, `{"error": "Permission Denied"}`, http.StatusUnauthorized)
			return
		}

		slog.Info("Permission granted", slog.String("permission", permissionName))
		next.ServeHTTP(w, r)
	}))
}

// AuthMiddlewareRequireRoleOrPermission grants access if the user has the named role
// or if one of their roles grants the named permission.
func AuthMiddlewareRequireRoleOrPermission(ua AuthService, roleName string, permissionName string, next http.Handler) http.Handler {
	return AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		userID, err := GetUserIDFromContext(r.Context())
		if err != nil {
			http.Error(w, `{"error": "Unable to determine user from token"}`, http.StatusUnauthorized)
			return
		}

		user, err := ua.GetUserById(userID)
		if err != nil {
			slog.Error("Failed to load user for role check", slog.String("userID", userID.String()), slog.String("error", err.Error()))
			http.Error(w, `{"error": "Unable to load user roles"}`, http.StatusUnauthorized)
			return
		}

		for _, userRole := range user.Roles {
			if strings.EqualFold(strings.TrimSpace(userRole), strings.TrimSpace(roleName)) {
				slog.Info("Role-based access granted", slog.String("role", roleName), slog.String("path", r.URL.Path))
				next.ServeHTTP(w, r)
				return
			}
		}

		claimsVal := r.Context().Value(ClaimsContextKey)
		claims, ok := claimsVal.(jwt.MapClaims)
		if !ok {
			http.Error(w, `{"error": "Invalid token claims"}`, http.StatusUnauthorized)
			return
		}

		roleIDsInterface, ok := claims["role_ids"].([]interface{})
		if !ok {
			http.Error(w, `{"error": "Role IDs missing or invalid"}`, http.StatusUnauthorized)
			return
		}

		var roleIDs []uuid.UUID
		for _, roleID := range roleIDsInterface {
			roleIDStr, ok := roleID.(string)
			if !ok {
				http.Error(w, `{"error": "Invalid role ID format"}`, http.StatusBadRequest)
				return
			}

			parsedUUID, err := uuid.Parse(roleIDStr)
			if err != nil {
				http.Error(w, `{"error": "Invalid UUID format"}`, http.StatusBadRequest)
				return
			}
			roleIDs = append(roleIDs, parsedUUID)
		}

		hasPermission, err := ua.VerifyUserRolesForPermission(roleIDs, permissionName)
		if err != nil {
			slog.Error("Permission check failed", slog.String("permission", permissionName), slog.String("error", err.Error()))
			http.Error(w, `{"error": "Internal Server Error"}`, http.StatusInternalServerError)
			return
		}

		if !hasPermission {
			http.Error(w, `{"error": "Permission Denied"}`, http.StatusUnauthorized)
			return
		}

		slog.Info("Permission-based access granted", slog.String("permission", permissionName), slog.String("path", r.URL.Path))
		next.ServeHTTP(w, r)
	}))
}

// parseAndValidateToken extracts the token from the request and returns the claims if valid
func parseAndValidateToken(r *http.Request) (jwt.MapClaims, error) {
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return nil, fmt.Errorf("missing or malformed Authorization header")
	}
	jwtToken := strings.TrimPrefix(authHeader, "Bearer ")

	secret := os.Getenv("JWT_KEY")
	if secret == "" {
		return nil, fmt.Errorf("JWT_KEY environment variable is not set")
	}

	token, err := jwt.Parse(jwtToken, func(token *jwt.Token) (interface{}, error) {
		// Only HMAC signing methods are supported
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid token: %v", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}
