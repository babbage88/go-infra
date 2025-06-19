package ssh_key_provider

import (
	"testing"
	"time"

	"github.com/babbage88/go-infra/services/user_secrets"
	"github.com/google/uuid"
)

// MockUserSecretProvider implements user_secrets.UserSecretProvider for testing
type MockUserSecretProvider struct {
	secrets map[uuid.UUID]bool // tracks which secrets exist
}

func NewMockUserSecretProvider() *MockUserSecretProvider {
	return &MockUserSecretProvider{
		secrets: make(map[uuid.UUID]bool),
	}
}

func (m *MockUserSecretProvider) StoreSecret(plaintextSecret string, userId, appId uuid.UUID, expiry time.Time) error {
	// Generate a mock secret ID
	secretId := uuid.New()
	m.secrets[secretId] = true
	return nil
}

func (m *MockUserSecretProvider) RetrieveSecret(secretId uuid.UUID) (*user_secrets.RetrievedUserSecret, error) {
	if !m.secrets[secretId] {
		return nil, nil
	}
	return &user_secrets.RetrievedUserSecret{}, nil
}

func (m *MockUserSecretProvider) GetUserSecretEntries(userId uuid.UUID) ([]user_secrets.UserSecretEntry, error) {
	return []user_secrets.UserSecretEntry{}, nil
}

func (m *MockUserSecretProvider) GetUserSecretEntriesByAppId(userId uuid.UUID, appId uuid.UUID) ([]user_secrets.UserSecretEntry, error) {
	return []user_secrets.UserSecretEntry{}, nil
}

func (m *MockUserSecretProvider) GetUserSecretEntriesByAppName(userId uuid.UUID, appName string) ([]user_secrets.UserSecretEntry, error) {
	return []user_secrets.UserSecretEntry{}, nil
}

func (m *MockUserSecretProvider) DeleteSecret(secretId uuid.UUID) error {
	if !m.secrets[secretId] {
		return nil // Secret doesn't exist, consider it already deleted
	}
	delete(m.secrets, secretId)
	return nil
}

func TestDeleteSShKeyAndSecret(t *testing.T) {
	// This is a basic test structure - in a real environment you'd need a test database
	// For now, we'll just test that the method signature is correct and the interface is implemented

	mockSecretProvider := NewMockUserSecretProvider()

	// Create a mock store (without actual database connection for this test)
	store := &PgSshKeySecretStore{
		DbConn:         nil, // Would be a test database connection
		SecretProvider: mockSecretProvider,
	}

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
