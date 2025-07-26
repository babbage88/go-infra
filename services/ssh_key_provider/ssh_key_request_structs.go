package ssh_key_provider

// swagger:parameters deleteSshKey
// Request to delete an SSH key and its associated secret
type DeleteSshKeyRequest struct {
	// ID of the SSH key to delete
	// in: path
	// required: true
	ID string `json:"id"`
}

// swagger:response DeleteSshKeyResponse
// Response after deleting an SSH key and its secret
type DeleteSshKeyResponse struct {
	// in: body
	Body struct {
		// Success message
		Message string `json:"message"`
	}
}
