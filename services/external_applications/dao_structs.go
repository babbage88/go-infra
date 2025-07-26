package external_applications

import (
	"time"

	"github.com/google/uuid"
)

// swagger:model ExternalApplicationDao
type ExternalApplicationDao struct {
	Id             uuid.UUID `json:"id"`
	Name           string    `json:"name"`
	CreatedAt      time.Time `json:"createdAt"`
	LastModified   time.Time `json:"lastModified"`
	EndpointUrl    string    `json:"endpointUrl,omitempty"`
	AppDescription string    `json:"appDescription,omitempty"`
}

// swagger:model CreateExternalApplicationRequest
type CreateExternalApplicationRequest struct {
	Name           string `json:"name" validate:"required"`
	EndpointUrl    string `json:"endpointUrl,omitempty"`
	AppDescription string `json:"appDescription,omitempty"`
}

// swagger:model UpdateExternalApplicationRequest
type UpdateExternalApplicationRequest struct {
	Name           string `json:"name,omitempty"`
	EndpointUrl    string `json:"endpointUrl,omitempty"`
	AppDescription string `json:"appDescription,omitempty"`
}
