package authapi

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/babbage88/go-infra/database/infra_db_pg"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type LocalAuthService struct {
	DbConn *pgxpool.Pool `json:"dbConn"`
}

func NewLoginRequest(username string, password string, isHashed bool) *UserLoginRequest {
	return &UserLoginRequest{UserName: username, Password: password}
}

func (ua *LocalAuthService) VerifyUser(userid uuid.UUID) bool {
	queries := infra_db_pg.New(ua.DbConn)
	qry, err := queries.GetUserById(context.Background(), userid)
	if err != nil {
		slog.Error("Error querying database for user", slog.String("Error", err.Error()), slog.String("UserName", fmt.Sprint(userid)))
	}
	return qry.Enabled
}

func (us *LocalAuthService) VerifyUserPermission(ueid uuid.UUID, permissionName string) (bool, error) {
	params := infra_db_pg.VerifyUserPermissionByIdParams{
		UserId:     pgtype.UUID{Bytes: ueid, Valid: true},
		Permission: pgtype.Text{String: permissionName, Valid: true},
	}
	queries := infra_db_pg.New(us.DbConn)
	qry, err := queries.VerifyUserPermissionById(context.Background(), params)
	if err != nil {
		slog.Error("error verifying user permissions", slog.String("error", err.Error()))
		return false, err
	}
	return qry, err
}

func (us *LocalAuthService) VerifyUserPermissionByRole(roleId uuid.UUID, permissionName string) (bool, error) {
	params := infra_db_pg.VerifyUserPermissionByRoleIdParams{
		RoleId:     roleId,
		Permission: pgtype.Text{String: permissionName, Valid: true},
	}

	queries := infra_db_pg.New(us.DbConn)
	qry, err := queries.VerifyUserPermissionByRoleId(context.Background(), params)
	if err != nil {
		slog.Error("error verifying user permissions", slog.String("error", err.Error()))
		return false, err
	}
	return qry, err
}

func (a *LocalAuthService) CreateAuthTokenOnLogin(id uuid.UUID, roleIds uuid.UUIDs, email string) (AuthToken, error) {
	tokens := AuthToken{UserID: id}

	jwt_algo := os.Getenv("JWT_ALGORITHM")
	signingMethod := jwt.GetSigningMethod(jwt_algo)
	// Create token
	token := jwt.New(signingMethod)

	// Set claims
	// This is the information which frontend can use
	// The backend can also decode the token and get admin etc.
	claims := token.Claims.(jwt.MapClaims)
	claims["sub"] = id
	claims["name"] = email
	claims["role_ids"] = roleIds
	claims["exp"] = time.Now().Add(time.Minute * 15).Unix()

	// Generate encoded token and send it as response.
	// The signing string should be secret (a generated UUID works too)
	t, err := token.SignedString([]byte(os.Getenv("JWT_KEY")))
	if err != nil {
		return tokens, err
	}
	tokens.Token = t

	refreshToken := jwt.New(jwt.SigningMethodHS256)
	rtClaims := refreshToken.Claims.(jwt.MapClaims)
	rtClaims["sub"] = id
	rtClaims["exp"] = time.Now().Add(time.Hour * 48).Unix()

	rt, err := refreshToken.SignedString([]byte(os.Getenv("JWT_KEY")))
	if err != nil {
		return tokens, err
	}

	tokens.RefreshToken = rt
	return tokens, nil
}

func (a *LocalAuthService) Login(loginReq *UserLoginRequest) UserLoginResponse {
	var response UserLoginResponse
	var result LoginResult
	username := pgtype.Text{String: loginReq.UserName, Valid: true}

	queries := infra_db_pg.New(a.DbConn)
	qry, err := queries.GetUserLogin(context.Background(), username)
	result.PasswordValid = VerifyPassword(loginReq.Password, qry.Password.String)

	if err != nil {
		slog.Error("Error querying database for user", slog.String("UserName", loginReq.UserName))
	}

	if !result.PasswordValid {
		slog.Error("Supplied password does not match the password stored in database", slog.String("User", loginReq.UserName))
		result.Success = false
		result.Error = errors.New("password does not match")
		result.UserEnabled = qry.Enabled
		response.Result = result
		return response
	}

	if !qry.Enabled {
		slog.Error("User is disabled", slog.String("User", loginReq.UserName))
		result.Success = false
		result.UserEnabled = qry.Enabled
		result.Error = errors.New("user is diabled")
		response.Result = result
		return response
	}
	slog.Info("Login was Successful")
	result.Success = true
	result.Error = nil
	result.UserNameMatches = true

	response.Result = result
	response.UserInfo.ParseUserRowFromDb(qry)

	return response
}

func (a *LocalAuthService) VerifyToken(tokenString string) error {
	jwtKey := []byte(os.Getenv("JWT_KEY"))
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil {
		return err
	}

	if !token.Valid {
		return fmt.Errorf("invalid token")
	}

	return nil
}

func (ua *LocalAuthService) ParseAccessToken(accessToken string) *InfraJWTClaim {
	jwtKey := os.Getenv("JWT_KEY")
	parsedAccessToken, _ := jwt.ParseWithClaims(accessToken, &InfraJWTClaim{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtKey), nil
	})

	return parsedAccessToken.Claims.(*InfraJWTClaim)
}

func (ua *LocalAuthService) VerifyUserRolesForPermission(roleIds uuid.UUIDs, permissionName string) (bool, error) {
	var lastError error // Store any encountered errors for logging or debugging

	for _, roleId := range roleIds {
		hasPermission, err := ua.VerifyUserPermissionByRole(roleId, permissionName)
		if err != nil {
			// Save the error but continue checking other roles
			slog.Error("Error encountered while verifying permissions", slog.String("roleId", roleId.String()), slog.String("error", err.Error()))
			lastError = err
			continue
		}
		if hasPermission {
			return true, err
		}
	}

	if lastError != nil {
		// Log the error for debugging purposes
		slog.Error("Error occurred while verifying permissions for roles", slog.String("Error", lastError.Error()))
	}
	// Return false if no roles grant the permission
	return false, lastError
}
