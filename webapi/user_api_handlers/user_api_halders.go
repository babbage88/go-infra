package userapi

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/babbage88/go-infra/services"
	"github.com/babbage88/go-infra/webutils/cors"
)

// swagger:route POST /create/user createuser idOfcreateUserEndpoint
// Create a new user.
//
// security:
// - bearer:
// responses:
//   200: UserDao

func CreateUser(uc_service *services.UserCRUDService) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		cors.HandlerCorsAndOptions(w, r)

		if r.Method != http.MethodPost {
			slog.Error("Invalid request method", slog.String("Method", r.Method))
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		var newUserReq CreateNewUserRequest

		err := json.NewDecoder(r.Body).Decode(&newUserReq)
		if err != nil {
			slog.Error("Failed to decode request body", slog.String("Error", err.Error()))
			http.Error(w, "Bad request: "+err.Error(), http.StatusBadRequest)
			return
		}

		newUser, err := uc_service.NewUser(
			newUserReq.NewUsername,
			newUserReq.NewUserPassword,
			newUserReq.NewUserEmail,
			newUserReq.NewUserRole)
		if err != nil {
			slog.Error("Error creating new user", slog.String("Error", err.Error()))
			http.Error(w, "Error createing new user "+err.Error(), http.StatusInternalServerError)
		}

		jsonResponse, err := json.Marshal(newUser)
		if err != nil {
			slog.Error("Failed to marshal JSON response", slog.String("Error", err.Error()))
			http.Error(w, "Failed to marshal JSON response: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(jsonResponse)
	}
}
