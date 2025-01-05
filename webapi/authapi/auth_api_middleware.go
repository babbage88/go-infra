package authapi

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/babbage88/go-infra/utils/env_helper"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware(envars *env_helper.EnvVars, next http.HandlerFunc) http.HandlerFunc {
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
				SECRETKEY := envars.GetVarMapValue("JWT_KEY")
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

func AuthMiddlewareRequirePermission(ua *UserAuthService, permissionName string, next http.HandlerFunc) http.HandlerFunc {
	slog.Info("Starting AuthMiddlewareRequirePermissions", slog.String("Required Perm", permissionName))
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
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Malformed Token"})
			w.Write([]byte(`{"error": "Malformed Token"}`))
			return
		}

		jwtToken := authHeader[1]
		token, err := jwt.Parse(jwtToken, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			SECRETKEY := ua.Envars.GetVarMapValue("JWT_KEY")
			if SECRETKEY == "" {
				return nil, fmt.Errorf("secret key not found")
			}
			return []byte(SECRETKEY), nil
		})

		if err != nil || !token.Valid {
			slog.Error("Error validating token", slog.String("Error", err.Error()))
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error": "Unauthorized"}`))
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || !token.Valid {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error": "Invalid claims"}`))
			return
		}

		roleIDsInterface, ok := claims["role_ids"].([]interface{})
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error": "Role IDs missing"}`))
			return
		}

		var roleIDs []int32
		for _, roleID := range roleIDsInterface {
			if id, ok := roleID.(float64); ok { // JWT encodes numbers as float64
				roleIDs = append(roleIDs, int32(id))
			}
		}

		slog.Info("Verifying permission for role id", slog.String("roleID", fmt.Sprint(roleIDs)), slog.String("PermName", permissionName))
		hasPermission, err := ua.VerifyUserRolesForPermission(roleIDs, permissionName)
		if err != nil {
			slog.Error("Error verifying user permission", slog.String("Error", err.Error()))
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": "Internal Server Error"}`))
			return
		}

		if !hasPermission {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error": "Permission Denied"}`))
			return
		}

		slog.Info("User permission verified successfully.", slog.Any("RoleIDs", roleIDs), slog.String("Permission", permissionName))
		next.ServeHTTP(w, r)
	}
}
