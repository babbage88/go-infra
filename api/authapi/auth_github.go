package authapi

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/babbage88/go-infra/database/infra_db_pg"
	"github.com/babbage88/go-infra/services/user_crud_svc"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	defaultGitHubCallbackPath    = "/auth/github/callback"
	defaultGitHubFrontendRoute   = "/login/github/callback"
	defaultGitHubUserRoleName    = "User"
	gitHubStateExpirationMinutes = 10
	gitHubAuthorizationEndpoint  = "https://github.com/login/oauth/authorize"
	gitHubAccessTokenEndpoint    = "https://github.com/login/oauth/access_token"
	gitHubUserProfileEndpoint    = "https://api.github.com/user"
	gitHubUserEmailsEndpoint     = "https://api.github.com/user/emails"
	gitHubAuthorizationScope     = "read:user user:email"
	gitHubUserAgent              = "compunity-github-auth"
)

type gitHubOAuthState struct {
	RedirectURI string `json:"redirect_uri"`
	Nonce       string `json:"nonce"`
}

type gitHubAccessTokenResponse struct {
	AccessToken string `json:"access_token"`
	Error       string `json:"error"`
}

type gitHubUserProfile struct {
	Login string `json:"login"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type gitHubUserEmail struct {
	Email    string `json:"email"`
	Primary  bool   `json:"primary"`
	Verified bool   `json:"verified"`
}

func GitHubLoginStartHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientID := os.Getenv("GITHUB_CLIENT_ID")
		if clientID == "" {
			http.Error(w, "GITHUB_CLIENT_ID is not configured", http.StatusInternalServerError)
			return
		}

		redirectURI := getFrontendRedirectURI(r)
		stateToken, err := createGitHubStateToken(redirectURI)
		if err != nil {
			slog.Error("failed to create github oauth state", slog.String("error", err.Error()))
			http.Error(w, "failed to initialize github login", http.StatusInternalServerError)
			return
		}

		query := url.Values{}
		query.Set("client_id", clientID)
		query.Set("redirect_uri", getGitHubBackendRedirectURI(r))
		query.Set("scope", gitHubAuthorizationScope)
		query.Set("state", stateToken)

		http.Redirect(w, r, gitHubAuthorizationEndpoint+"?"+query.Encode(), http.StatusTemporaryRedirect)
	})
}

func GitHubLoginCallbackHandler(authService AuthService, userCRUDService *user_crud_svc.UserCRUDService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		redirectURI := getFrontendRedirectURI(r)

		if authService == nil || userCRUDService == nil {
			redirectWithGitHubAuthError(w, r, redirectURI, "github auth is unavailable")
			return
		}

		if githubError := r.URL.Query().Get("error"); githubError != "" {
			redirectWithGitHubAuthError(w, r, redirectURI, githubError)
			return
		}

		stateToken := r.URL.Query().Get("state")
		if stateToken == "" {
			redirectWithGitHubAuthError(w, r, redirectURI, "missing oauth state")
			return
		}

		state, err := parseGitHubStateToken(stateToken)
		if err != nil {
			slog.Warn("github oauth state validation failed", slog.String("error", err.Error()))
			redirectWithGitHubAuthError(w, r, redirectURI, "invalid oauth state")
			return
		}
		if state.RedirectURI != "" {
			redirectURI = state.RedirectURI
		}

		code := r.URL.Query().Get("code")
		if code == "" {
			redirectWithGitHubAuthError(w, r, redirectURI, "missing oauth code")
			return
		}

		accessToken, err := exchangeGitHubCode(r.Context(), code, getGitHubBackendRedirectURI(r))
		if err != nil {
			slog.Error("github oauth code exchange failed", slog.String("error", err.Error()))
			redirectWithGitHubAuthError(w, r, redirectURI, "failed to exchange github code")
			return
		}

		profile, err := fetchGitHubProfile(r.Context(), accessToken)
		if err != nil {
			slog.Error("github profile lookup failed", slog.String("error", err.Error()))
			redirectWithGitHubAuthError(w, r, redirectURI, "failed to load github profile")
			return
		}

		user, err := findOrCreateGitHubUser(r.Context(), authService, userCRUDService, profile)
		if err != nil {
			slog.Error("github user provisioning failed", slog.String("error", err.Error()))
			redirectWithGitHubAuthError(w, r, redirectURI, "failed to provision user")
			return
		}

		tokenPair, err := authService.CreateAuthTokenOnLogin(user.Id, user.RoleIds, user.Email)
		if err != nil {
			slog.Error("failed to create auth token after github login", slog.String("error", err.Error()))
			redirectWithGitHubAuthError(w, r, redirectURI, "failed to create session")
			return
		}

		tokenPair.Username = user.UserName
		tokenPair.Email = user.Email
		redirectWithGitHubAuthSuccess(w, r, redirectURI, tokenPair)
	})
}

func findOrCreateGitHubUser(ctx context.Context, authService AuthService, userCRUDService *user_crud_svc.UserCRUDService, profile gitHubUserProfile) (*user_crud_svc.UserDao, error) {
	if profile.Email != "" {
		user, err := authService.GetUserByUsernameOrEmail(profile.Email)
		switch {
		case err == nil:
			return ensureGitHubUserEnabled(user)
		case !errors.Is(err, sql.ErrNoRows):
			return nil, err
		}
	}

	if profile.Login != "" {
		user, err := authService.GetUserByUsernameOrEmail(profile.Login)
		switch {
		case err == nil:
			return ensureGitHubUserEnabled(user)
		case !errors.Is(err, sql.ErrNoRows):
			return nil, err
		}
	}

	if profile.Email == "" {
		return nil, fmt.Errorf("github account did not provide an email address")
	}

	username, err := nextAvailableGitHubUsername(userCRUDService, profile)
	if err != nil {
		return nil, err
	}

	newUser, err := userCRUDService.NewUser(username, randomGitHubPassword(), profile.Email)
	if err != nil {
		return nil, err
	}

	roleName := os.Getenv("GITHUB_AUTH_DEFAULT_ROLE")
	if roleName == "" {
		roleName = defaultGitHubUserRoleName
	}

	roleID, err := infra_db_pg.New(userCRUDService.DbConn).GetRoleIdByName(ctx, roleName)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve default github role %q: %w", roleName, err)
	}

	if err := userCRUDService.UpdateUserRoleMapping(newUser.Id, roleID); err != nil {
		return nil, fmt.Errorf("failed to assign role %q to github user: %w", roleName, err)
	}

	return userCRUDService.GetUserById(newUser.Id)
}

func ensureGitHubUserEnabled(user *user_crud_svc.UserDao) (*user_crud_svc.UserDao, error) {
	if user == nil {
		return nil, fmt.Errorf("user lookup returned nil")
	}
	if !user.Enabled || user.IsDeleted {
		return nil, fmt.Errorf("user %s is disabled or deleted", user.UserName)
	}
	return user, nil
}

func nextAvailableGitHubUsername(userCRUDService *user_crud_svc.UserCRUDService, profile gitHubUserProfile) (string, error) {
	base := strings.TrimSpace(profile.Login)
	if base == "" {
		base = strings.TrimSpace(profile.Name)
	}
	base = normalizeGitHubUsername(base)
	if base == "" {
		base = "github-user"
	}

	candidates := []string{base}
	for i := range 5 {
		candidates = append(candidates, fmt.Sprintf("%s-%s", base, randomSuffix(4+i)))
	}

	for _, candidate := range candidates {
		_, err := userCRUDService.GetUserByName(candidate)
		if errors.Is(err, sql.ErrNoRows) {
			return candidate, nil
		}
		if err != nil {
			return "", err
		}
	}

	return "", fmt.Errorf("failed to generate a unique github username")
}

func normalizeGitHubUsername(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return ""
	}

	var builder strings.Builder
	lastWasDash := false
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			builder.WriteRune(r)
			lastWasDash = false
		case r == '-' || r == '_' || r == ' ':
			if !lastWasDash && builder.Len() > 0 {
				builder.WriteByte('-')
				lastWasDash = true
			}
		}
	}

	return strings.Trim(builder.String(), "-")
}

func randomGitHubPassword() string {
	return "github-oauth-" + randomSuffix(24)
}

func randomSuffix(length int) string {
	byteCount := length
	if byteCount < 1 {
		byteCount = 8
	}

	raw := make([]byte, byteCount)
	if _, err := rand.Read(raw); err != nil {
		return uuid.NewString()
	}

	encoded := base64.RawURLEncoding.EncodeToString(raw)
	if len(encoded) < length {
		return encoded
	}

	return encoded[:length]
}

func createGitHubStateToken(frontendRedirectURI string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"redirect_uri": frontendRedirectURI,
		"nonce":        randomSuffix(16),
		"exp":          time.Now().Add(gitHubStateExpirationMinutes * time.Minute).Unix(),
	})

	return token.SignedString([]byte(os.Getenv("JWT_KEY")))
}

func parseGitHubStateToken(stateToken string) (*gitHubOAuthState, error) {
	token, err := jwt.Parse(stateToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("JWT_KEY")), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid github oauth state")
	}

	redirectURI, _ := claims["redirect_uri"].(string)
	nonce, _ := claims["nonce"].(string)

	return &gitHubOAuthState{
		RedirectURI: redirectURI,
		Nonce:       nonce,
	}, nil
}

func exchangeGitHubCode(ctx context.Context, code string, redirectURI string) (string, error) {
	values := url.Values{}
	values.Set("client_id", os.Getenv("GITHUB_CLIENT_ID"))
	values.Set("client_secret", os.Getenv("GITHUB_CLIENT_SECRET"))
	values.Set("code", code)
	values.Set("redirect_uri", redirectURI)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, gitHubAccessTokenEndpoint, strings.NewReader(values.Encode()))
	if err != nil {
		return "", err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", gitHubUserAgent)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		return "", fmt.Errorf("github token endpoint returned status %d", resp.StatusCode)
	}

	var tokenResponse gitHubAccessTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		return "", err
	}
	if tokenResponse.Error != "" {
		return "", fmt.Errorf("github token endpoint returned error %q", tokenResponse.Error)
	}
	if tokenResponse.AccessToken == "" {
		return "", fmt.Errorf("github token endpoint returned an empty access token")
	}

	return tokenResponse.AccessToken, nil
}

func fetchGitHubProfile(ctx context.Context, accessToken string) (gitHubUserProfile, error) {
	profile, err := fetchGitHubUserProfile(ctx, accessToken)
	if err != nil {
		return profile, err
	}

	if strings.TrimSpace(profile.Email) == "" {
		email, err := fetchGitHubPrimaryEmail(ctx, accessToken)
		if err != nil {
			return profile, err
		}
		profile.Email = email
	}

	profile.Email = strings.TrimSpace(profile.Email)
	profile.Login = strings.TrimSpace(profile.Login)
	profile.Name = strings.TrimSpace(profile.Name)

	return profile, nil
}

func fetchGitHubUserProfile(ctx context.Context, accessToken string) (gitHubUserProfile, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, gitHubUserProfileEndpoint, nil)
	if err != nil {
		return gitHubUserProfile{}, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("User-Agent", gitHubUserAgent)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return gitHubUserProfile{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		return gitHubUserProfile{}, fmt.Errorf("github user endpoint returned status %d", resp.StatusCode)
	}

	var profile gitHubUserProfile
	if err := json.NewDecoder(resp.Body).Decode(&profile); err != nil {
		return gitHubUserProfile{}, err
	}

	return profile, nil
}

func fetchGitHubPrimaryEmail(ctx context.Context, accessToken string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, gitHubUserEmailsEndpoint, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("User-Agent", gitHubUserAgent)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		return "", fmt.Errorf("github email endpoint returned status %d", resp.StatusCode)
	}

	var emails []gitHubUserEmail
	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return "", err
	}

	for _, email := range emails {
		if email.Primary && email.Verified && strings.TrimSpace(email.Email) != "" {
			return email.Email, nil
		}
	}

	for _, email := range emails {
		if email.Verified && strings.TrimSpace(email.Email) != "" {
			return email.Email, nil
		}
	}

	return "", fmt.Errorf("github account does not have a verified email address")
}

func redirectWithGitHubAuthSuccess(w http.ResponseWriter, r *http.Request, redirectURI string, authToken AuthToken) {
	fragment := url.Values{}
	fragment.Set("accessToken", authToken.Token)
	fragment.Set("refreshToken", authToken.RefreshToken)
	fragment.Set("user_id", authToken.UserID.String())
	fragment.Set("userName", authToken.Username)
	fragment.Set("email", authToken.Email)

	redirectTo := redirectURI + "#" + fragment.Encode()
	http.Redirect(w, r, redirectTo, http.StatusTemporaryRedirect)
}

func redirectWithGitHubAuthError(w http.ResponseWriter, r *http.Request, redirectURI string, message string) {
	fragment := url.Values{}
	fragment.Set("error", message)
	http.Redirect(w, r, redirectURI+"#"+fragment.Encode(), http.StatusTemporaryRedirect)
}

func getGitHubBackendRedirectURI(r *http.Request) string {
	if configured := os.Getenv("GITHUB_REDIRECT_URL"); configured != "" {
		return configured
	}

	scheme := "https"
	if r.TLS == nil {
		scheme = "http"
	}

	return fmt.Sprintf("%s://%s%s", scheme, r.Host, defaultGitHubCallbackPath)
}

func getFrontendRedirectURI(r *http.Request) string {
	if redirectURI := strings.TrimSpace(r.URL.Query().Get("redirect_uri")); redirectURI != "" {
		return redirectURI
	}

	if frontendBase := strings.TrimRight(os.Getenv("FRONTEND_URL"), "/"); frontendBase != "" {
		return frontendBase + defaultGitHubFrontendRoute
	}

	scheme := "https"
	if r.TLS == nil {
		scheme = "http"
	}

	return fmt.Sprintf("%s://%s%s", scheme, r.Host, defaultGitHubFrontendRoute)
}
