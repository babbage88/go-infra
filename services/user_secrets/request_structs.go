package user_secrets

import (
	"time"

	"github.com/google/uuid"
)

// swagger:parameters deleteUserSecretByID
type DeleteSecretByIdRequest struct {
	// In: path
	ID string `json:"ID"`
}

// swagger:parameters createUserSecret
type CreateSecretRequestWrapper struct {
	// in:body
	Body CreateSecretRequest `json:"body"`
}

type CreateSecretRequest struct {
	ApplicationID uuid.UUID `json:"application_id"`
	Secret        string    `json:"secret"`
	Expiration    time.Time `json:"expiration"`
}

// swagger:parameters getUserSecretByID
type RetrievedSecretRequest struct {
	// ID of secret
	//
	// In: path
	ID string `json:"ID"`
}

// swagger:response RetrievedSecretResponse
type RetrievedSecretResponse struct {
	// in: body
	Body struct {
		ID                  uuid.UUID `json:"id"`
		UserID              uuid.UUID `json:"user_id"`
		ExternalApplication uuid.UUID `json:"external_application_id"`
		Expiration          time.Time `json:"expiration,omitempty"`
		Token               string    `json:"token"`
	}
}

// swagger:parameters GetUserSecretEntries
type GetUserSecretEntriesRequest struct {
	// In: path
	USERID uuid.UUID
}

// swagger:response GetUserSecretEntriesResponse
type GetUserSecretEntriesResponseWrapper struct {
	// in: body
	Body []UserSecretEntry `json:"userSecretEntries"`
}

// swagger:parameters GetUserSecretEntriesByAppId
type GetUserSecretEntriesByAppIdRequest struct {
	// In: path
	USERID uuid.UUID
	// In: path
	APPID uuid.UUID
}
