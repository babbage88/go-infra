package authapi

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Type", "application/json")
		authHeader := strings.Split(r.Header.Get("Authorization"), "Bearer ")
		if authHeader == nil {
			slog.Error("Auth Header is nil")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Auth Header is nil"})
			return
		}
		if len(authHeader) != 2 {
			slog.Error("Malformed token")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Malformed Token"})
			return
		} else {
			jwtToken := authHeader[1]
			token, err := jwt.Parse(jwtToken, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				// Retrieve the secret key from environment variables
				SECRETKEY := os.Getenv("JWT_KEY")
				if SECRETKEY == "" {
					return nil, fmt.Errorf("secret key not found")
				}
				return []byte(SECRETKEY), nil
			})

			if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
				ctx := context.WithValue(r.Context(), "props", claims)
				next.ServeHTTP(w, r.WithContext(ctx))
			} else {
				slog.Error("Error validating token", slog.String("Error", err.Error()))
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized"})
			}
		}

		slog.Info("Token has been verified.", slog.String("Host", r.URL.Host), slog.String("Path", r.URL.Path))
	}
}

func AuthMiddlewareRequirePermission(ua AuthService, permissionName string, next http.HandlerFunc) http.HandlerFunc {
	slog.Info("Starting AuthMiddlewareRequirePermissions", slog.String("Required Perm", permissionName))

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		authHeader := strings.Split(r.Header.Get("Authorization"), "Bearer ")
		if authHeader == nil || len(authHeader) != 2 {
			slog.Error("Malformed or missing Authorization header")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized"})
			return
		}

		jwtToken := authHeader[1]
		token, err := jwt.Parse(jwtToken, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			SECRETKEY := os.Getenv("JWT_KEY")
			if SECRETKEY == "" {
				return nil, fmt.Errorf("secret key not found")
			}
			return []byte(SECRETKEY), nil
		})

		if err != nil || !token.Valid {
			slog.Error("Error validating token", slog.String("Error", err.Error()))
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized"})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || !token.Valid {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid claims"})
			return
		}

		roleIDsInterface, ok := claims["role_ids"].([]interface{})
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Role IDs missing"})
			return
		}

		var roleIDs uuid.UUIDs
		for _, roleID := range roleIDsInterface {
			roleIDStr, ok := roleID.(string)
			if !ok {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]string{"error": "Invalid role ID format"})
				return
			}

			parsedUUID, err := uuid.Parse(roleIDStr)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]string{"error": "Invalid UUID format"})
				return
			}

			roleIDs = append(roleIDs, parsedUUID)
		}

		slog.Info("Verifying permission for role IDs",
			slog.Any("roleIDs", roleIDs),
			slog.String("PermName", permissionName))

		hasPermission, err := ua.VerifyUserRolesForPermission(roleIDs, permissionName)
		if err != nil {
			slog.Error("Error verifying user permission", slog.String("Error", err.Error()))
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Internal Server Error"})
			return
		}

		if !hasPermission {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Permission Denied"})
			return
		}

		slog.Info("User permission verified successfully.",
			slog.Any("RoleIDs", roleIDs),
			slog.String("Permission", permissionName))

		next.ServeHTTP(w, r)
	}
}
