package user_secrets

import (
	"time"

	"github.com/google/uuid"
)

// swagger:parameters createUserSecret
type CreateSecretRequestWrapper struct {
	// in:body
	Body CreateSecretRequest `json:"body"`
}

type CreateSecretRequest struct {
	ApplicationID uuid.UUID `json:"application_id"`
	Secret        string    `json:"secret"`
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
	// in:body
	Body struct {
		ID                  uuid.UUID `json:"id"`
		UserID              uuid.UUID `json:"user_id"`
		ExternalApplication uuid.UUID `json:"external_application_id"`
		Expiration          time.Time `json:"expiration,omitempty"`
		Token               string    `json:"token"`
	}
}
