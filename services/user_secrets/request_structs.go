package user_secrets

import (
	"time"

	"github.com/google/uuid"
)

// swagger:parameters createUserSecret
type CreateSecretRequest struct {
	// The application ID the token is for
	// required: true
	ApplicationID uuid.UUID `json:"application_id"`

	// The secret string, such as a bearer token
	// required: true
	Secret string `json:"secret"`
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
