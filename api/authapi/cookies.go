package authapi

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	accessTokenCookieName  = "infractl_access_token"
	refreshTokenCookieName = "infractl_refresh_token"
)

func setAuthCookies(w http.ResponseWriter, accessToken string, refreshToken string, accessExpiration time.Time) {
	refreshExpiration := time.Now().Add(time.Minute * time.Duration(getRefreshTokenExpirationMinutes()))

	http.SetCookie(w, buildAuthCookie(accessTokenCookieName, accessToken, accessExpiration))
	http.SetCookie(w, buildAuthCookie(refreshTokenCookieName, refreshToken, refreshExpiration))
}

func clearAuthCookies(w http.ResponseWriter) {
	expired := time.Unix(0, 0)
	http.SetCookie(w, buildAuthCookie(accessTokenCookieName, "", expired))
	http.SetCookie(w, buildAuthCookie(refreshTokenCookieName, "", expired))
}

func buildAuthCookie(name string, value string, expires time.Time) *http.Cookie {
	maxAge := int(time.Until(expires).Seconds())
	if value == "" || maxAge < 0 {
		maxAge = -1
	}

	return &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		Secure:   shouldUseSecureCookies(),
		SameSite: getAuthCookieSameSite(),
		Expires:  expires,
		MaxAge:   maxAge,
	}
}

func GetAccessTokenFromRequest(r *http.Request) string {
	if authHeader := strings.TrimSpace(r.Header.Get("Authorization")); strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
		return strings.TrimSpace(authHeader[len("Bearer "):])
	}

	if cookie, err := r.Cookie(accessTokenCookieName); err == nil {
		return strings.TrimSpace(cookie.Value)
	}

	return ""
}

func GetRefreshTokenFromRequest(r *http.Request) string {
	if cookie, err := r.Cookie(refreshTokenCookieName); err == nil {
		return strings.TrimSpace(cookie.Value)
	}

	return ""
}

func getAuthCookieSameSite() http.SameSite {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("AUTH_COOKIE_SAMESITE"))) {
	case "none":
		return http.SameSiteNoneMode
	case "strict":
		return http.SameSiteStrictMode
	case "lax", "":
		return http.SameSiteLaxMode
	default:
		return http.SameSiteLaxMode
	}
}

func shouldUseSecureCookies() bool {
	value := strings.ToLower(strings.TrimSpace(os.Getenv("AUTH_COOKIE_SECURE")))
	return value == "1" || value == "true" || value == "yes"
}

func getRefreshTokenExpirationMinutes() int64 {
	expireMinutesInt, err := parseEnvInt64("REFRESH_TOKEN_EXPIRIRATION_MINUTES", defaultRefreshTokenExpirationMinutes)
	if err != nil {
		return defaultRefreshTokenExpirationMinutes
	}

	if expireMinutesInt <= 0 {
		return defaultRefreshTokenExpirationMinutes
	}

	return expireMinutesInt
}

func parseEnvInt64(name string, fallback int64) (int64, error) {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return fallback, nil
	}

	var value int64
	if _, err := fmt.Sscan(raw, &value); err != nil {
		return fallback, err
	}

	return value, nil
}
