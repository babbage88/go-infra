package authapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/babbage88/go-infra/utils/env_helper"
	"github.com/golang-jwt/jwt/v5"
)

func LoginCookie(ua_service *UserAuthService) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		uname := r.Form.Get("username")
		pw := r.Form.Get("password")
		loginReq := &UserLoginRequest{UserName: uname, Password: pw}

		LoginResult := loginReq.Login(ua_service.DbConn)
		if !LoginResult.Result.Success {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": LoginResult.Result.Error.Error()})
			return
		}

		token, err := ua_service.CreateAuthTokenOnLogin(LoginResult.UserInfo.Id, LoginResult.UserInfo.Role, LoginResult.UserInfo.Email)
		if err != nil {
			slog.Error("Error creating auth token", slog.String("Error", err.Error()))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Set the JWT as an HTTP-only cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "auth_token",
			Value:    token.Token, // Use the token value
			Path:     "/logincookie",
			MaxAge:   3600, // 1 hour
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteLaxMode,
		})

		// Set the JWT as an HTTP-only cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "refresh_token",
			Value:    token.RefreshToken,
			Path:     "/logincookie",
			MaxAge:   3600, // 1 hour
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteLaxMode,
		})

		w.WriteHeader(http.StatusOK)
	}
}

func AuthCookieMiddleware(envars *env_helper.EnvVars, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Get the JWT from the cookie
		cookie, err := r.Cookie("auth_token")
		if err != nil {
			if errors.Is(err, http.ErrNoCookie) {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			slog.Error("Error reading cookie", slog.String("Error", err.Error()))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Parse and validate the JWT
		token, err := jwt.Parse(cookie.Value, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
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
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
		}
	}
}
