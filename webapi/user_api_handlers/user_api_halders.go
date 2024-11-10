package userapi

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/babbage88/go-infra/auth/hashing"
	jwt_auth "github.com/babbage88/go-infra/auth/tokens"
	infra_db "github.com/babbage88/go-infra/database/infra_db"
	db_models "github.com/babbage88/go-infra/database/models"
	"github.com/babbage88/go-infra/utils/env_helper"
	"github.com/babbage88/go-infra/webutils/cors"
)

func UserCreateHandler(envars *env_helper.EnvVars, db *sql.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			cors.EnableCors(&w)
			return
		}

		cors.EnableCors(&w)

		w.Header().Set("Content-Type", "application/json")

		var u db_models.User
		json.NewDecoder(r.Body).Decode(&u)
		fmt.Printf("The user request value %v", u)

		dbuser, err := infra_db.GetUserByUsername(db, u.Username)
		if err != nil {
			slog.Error("Error getting user from database", slog.String("Error", err.Error()))
		}
		verify_pw := hashing.VerifyPassword(u.Password, dbuser.Password)

		if verify_pw {
			token, err := jwt_auth.CreateTokenanAddToDb(envars, db, dbuser.Id, u.Role, u.Email)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				slog.Error("Error verifying password", slog.String("Error", err.Error()))
			}
			jsonResponse, _ := json.Marshal(token)
			w.WriteHeader(http.StatusOK)
			w.Write(jsonResponse)
			return
		} else {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprint(w, "Invalid credentials")
		}
	}
}
