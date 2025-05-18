package user_secrets

import (
	"time"

	"github.com/google/uuid"
)

type ExternalApplicationAuthToken struct {
	Id                    uuid.UUID `json:"id"`
	UserID                uuid.UUID `json:"user_id"`
	ExternalApplicationId uuid.UUID `json:"appId"`
	Token                 []byte    `json:"token"`
	Expiration            time.Time `json:"expiration"`
	CreatedAt             time.Time `json:"created_at"`
	LastModified          time.Time `json:"last_modified"`
}

type ExternalApplication struct {
	Id   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}
