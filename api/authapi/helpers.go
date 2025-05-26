package authapi

import (
	"context"
	"errors"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// GetUserIDFromContext extracts the user ID from the JWT "sub" claim in context
func GetUserIDFromContext(ctx context.Context) (uuid.UUID, error) {
	claims, ok := ctx.Value(ClaimsContextKey).(jwt.MapClaims)
	if !ok {
		return uuid.Nil, errors.New("no jwt claims in context")
	}

	sub, ok := claims["sub"].(string)
	if !ok {
		return uuid.Nil, errors.New("no sub claim in token")
	}

	id, err := uuid.Parse(sub)
	if err != nil {
		return uuid.Nil, err
	}

	return id, nil
}

// GetRoleIDsFromContext extracts the "role_ids" from the JWT in context
func GetRoleIDsFromContext(ctx context.Context) []uuid.UUID {
	claims, ok := ctx.Value(ClaimsContextKey).(jwt.MapClaims)
	if !ok {
		return nil
	}

	rawRoleIDs, ok := claims["role_ids"].([]interface{})
	if !ok {
		return nil
	}

	var ids []uuid.UUID
	for _, id := range rawRoleIDs {
		if idStr, ok := id.(string); ok {
			parsed, err := uuid.Parse(idStr)
			if err == nil {
				ids = append(ids, parsed)
			}
		}
	}
	return ids
}
