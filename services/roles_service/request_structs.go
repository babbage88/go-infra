package rolesservice

// swagger:parameters GetAllAppPermissionMappings
type GetRolePermissionMappingsRequest struct {
	// Optional role ID filter.
	//
	// In: query
	RoleID string `json:"roleId"`

	// Optional role name filter.
	//
	// In: query
	RoleName string `json:"roleName"`
}
