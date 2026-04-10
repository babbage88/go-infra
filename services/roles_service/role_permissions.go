package rolesservice

import (
	"context"
	"time"

	"github.com/babbage88/go-infra/database/infra_db_pg"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// swagger:model GetRolePermissionMappingsResponse
type GetRolePermissionMappingsResponse struct {
	Id           uuid.UUID `json:"id"`
	RoleId       uuid.UUID `json:"roleId"`
	RoleName     string    `json:"roleName"`
	PermissionId uuid.UUID `json:"permissionId"`
	Permission   string    `json:"permission"`
	CreatedAt    time.Time `json:"createdAt"`
	LastModified time.Time `json:"lastModified"`
}

// swagger:response GetRolePermissionMappingsResponse
type GetRolesPermissionMappingResponseWrapper struct {
	// in: body
	Body []GetRolePermissionMappingsResponse `json:"body"`
}

type RolePermissionsService interface {
	GetRolePermissionMappings() ([]GetRolePermissionMappingsResponse, error)
	GetRolePermissionMappingsByRoleID(roleID uuid.UUID) ([]GetRolePermissionMappingsResponse, error)
	GetRolePermissionMappingsByRoleName(roleName string) ([]GetRolePermissionMappingsResponse, error)
}

type RoleCRUDService struct {
	DbConn *pgxpool.Pool
}

func (rps *RoleCRUDService) GetRolePermissionMappings() ([]GetRolePermissionMappingsResponse, error) {
	queries := infra_db_pg.New(rps.DbConn)
	dbrolemappings, err := queries.GetAllRolePermissions(context.Background())
	if err != nil {
		return nil, err
	}

	response := make([]GetRolePermissionMappingsResponse, 0, len(dbrolemappings))
	for _, mapping := range dbrolemappings {
		response = append(response, mapRolePermissionMapping(
			mapping.RoleId,
			mapping.Role,
			mapping.PermissionId,
			mapping.Permission,
		))
	}
	return response, nil
}

func (rps *RoleCRUDService) GetRolePermissionMappingsByRoleID(roleID uuid.UUID) ([]GetRolePermissionMappingsResponse, error) {
	queries := infra_db_pg.New(rps.DbConn)
	dbrolemappings, err := queries.GetRolePermissionMappingByRoleId(context.Background(), roleID)
	if err != nil {
		return nil, err
	}

	response := make([]GetRolePermissionMappingsResponse, 0, len(dbrolemappings))
	for _, mapping := range dbrolemappings {
		response = append(response, mapRolePermissionMapping(
			mapping.RoleId,
			mapping.Role,
			mapping.PermissionId,
			mapping.Permission,
		))
	}
	return response, nil
}

func (rps *RoleCRUDService) GetRolePermissionMappingsByRoleName(roleName string) ([]GetRolePermissionMappingsResponse, error) {
	queries := infra_db_pg.New(rps.DbConn)
	dbrolemappings, err := queries.GetRolePermissionMappingByRoleName(context.Background(), roleName)
	if err != nil {
		return nil, err
	}

	response := make([]GetRolePermissionMappingsResponse, 0, len(dbrolemappings))
	for _, mapping := range dbrolemappings {
		response = append(response, mapRolePermissionMapping(
			mapping.RoleId,
			mapping.Role,
			mapping.PermissionId,
			mapping.Permission,
		))
	}
	return response, nil
}

func mapRolePermissionMapping(roleID uuid.UUID, roleName string, permissionID pgtype.UUID, permission pgtype.Text) GetRolePermissionMappingsResponse {
	response := GetRolePermissionMappingsResponse{
		Id:           uuid.Nil,
		RoleId:       roleID,
		RoleName:     roleName,
		PermissionId: uuid.Nil,
		Permission:   "",
		CreatedAt:    time.Time{},
		LastModified: time.Time{},
	}

	if permissionID.Valid {
		response.PermissionId = permissionID.Bytes
	}

	if permission.Valid {
		response.Permission = permission.String
	}

	return response
}
