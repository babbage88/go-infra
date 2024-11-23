package userapi

type CreateNewUserRequest struct {
	NewUsername     string `json:"newUsername"`
	NewUserPassword string `json:"newPassword"`
	NewUserEmail    string `json:"newEmail"`
	NewUserRole     string `json:"newUserRole"`
}
