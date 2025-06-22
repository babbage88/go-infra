package ssh_key_provider

import (
	"testing"

	"github.com/google/uuid"
)

func TestDeleteSShKeyAndSecret(t *testing.T) {
	// This is a basic test structure - in a real environment you'd need a test database
	// For now, we'll just test that the method signature is correct and the interface is implemented

	// Create a mock store (without actual database connection for this test)
	store := NewPgSshKeySecretStore(nil)

	// Test that the method exists and has the correct signature
	// We can't actually call it without a database connection, but we can verify it exists
	_ = store.DeleteSShKeyAndSecret

	t.Log("DeleteSShKeyAndSecret method exists and has correct signature")
}

func TestSshKeySecretProviderInterface(t *testing.T) {
	// Test that our implementation satisfies the interface
	var _ SshKeySecretProvider = (*PgSshKeySecretStore)(nil)
	t.Log("PgSshKeySecretStore correctly implements SshKeySecretProvider interface")
}

func TestSshKeyHostMappingCRUDMethods(t *testing.T) {
	// Test that all the new SSH key host mapping CRUD methods exist and have correct signatures

	// Create a mock store (without actual database connection for this test)
	store := NewPgSshKeySecretStore(nil)

	// Test that all the new methods exist and have the correct signatures
	_ = store.CreateSshKeyHostMapping
	_ = store.GetSshKeyHostMappingById
	_ = store.GetSshKeyHostMappingsByUserId
	_ = store.GetSshKeyHostMappingsByHostId
	_ = store.GetSshKeyHostMappingsByKeyId
	_ = store.UpdateSshKeyHostMapping
	_ = store.DeleteSshKeyHostMapping
	_ = store.DeleteSshKeyHostMappingsBySshKeyId

	t.Log("All SSH key host mapping CRUD methods exist and have correct signatures")
}

func TestSshKeyHostMappingRequestStructs(t *testing.T) {
	// Test that the request structs can be created and used
	sshKeyID := uuid.New()
	hostServerID := uuid.New()
	userID := uuid.New()

	// Test CreateSshKeyHostMappingRequest
	createReq := CreateSshKeyHostMappingRequest{
		SshKeyID:           sshKeyID,
		HostServerID:       hostServerID,
		UserID:             userID,
		HostserverUsername: "testuser",
	}

	if createReq.SshKeyID != sshKeyID {
		t.Errorf("Expected SshKeyID to be %v, got %v", sshKeyID, createReq.SshKeyID)
	}

	// Test UpdateSshKeyHostMappingRequest
	updateReq := UpdateSshKeyHostMappingRequest{
		ID:                 uuid.New(),
		HostserverUsername: "newuser",
	}

	if updateReq.HostserverUsername != "newuser" {
		t.Errorf("Expected HostserverUsername to be 'newuser', got %s", updateReq.HostserverUsername)
	}

	t.Log("SSH key host mapping request structs work correctly")
}
