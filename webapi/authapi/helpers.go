package authapi

import (
	"context"
	"errors"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func GetUserIDFromContext(ctx context.Context) (uuid.UUID, error) {
	claims, ok := ctx.Value("props").(jwt.MapClaims)
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

func GetRoleIDsFromContext(ctx context.Context) uuid.UUIDs {
	claims, ok := ctx.Value("props").(jwt.MapClaims)
	if !ok {
		return nil
	}

	rawRoleIDs, ok := claims["role_ids"].([]interface{})
	if !ok {
		return nil
	}

	var ids uuid.UUIDs
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
