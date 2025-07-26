package external_applications

import (
	"testing"

	"github.com/google/uuid"
)

// TestExternalApplicationsInterface ensures the interface is properly defined
func TestExternalApplicationsInterface(t *testing.T) {
	// This test ensures that ExternalApplicationsService implements the ExternalApplications interface
	var _ ExternalApplications = (*ExternalApplicationsService)(nil)
}

// TestCreateExternalApplicationRequest validates the request structure
func TestCreateExternalApplicationRequest(t *testing.T) {
	req := CreateExternalApplicationRequest{
		Name:           "test-app",
		EndpointUrl:    "https://api.test.com",
		AppDescription: "Test application for integration",
	}

	if req.Name != "test-app" {
		t.Errorf("Expected name to be 'test-app', got %s", req.Name)
	}

	if req.EndpointUrl != "https://api.test.com" {
		t.Errorf("Expected endpoint URL to be 'https://api.test.com', got %s", req.EndpointUrl)
	}

	if req.AppDescription != "Test application for integration" {
		t.Errorf("Expected app description to be 'Test application for integration', got %s", req.AppDescription)
	}
}

// TestUpdateExternalApplicationRequest validates the update request structure
func TestUpdateExternalApplicationRequest(t *testing.T) {
	req := UpdateExternalApplicationRequest{
		Name:           "updated-app",
		EndpointUrl:    "https://api.updated.com",
		AppDescription: "Updated test application",
	}

	if req.Name != "updated-app" {
		t.Errorf("Expected name to be 'updated-app', got %s", req.Name)
	}

	if req.EndpointUrl != "https://api.updated.com" {
		t.Errorf("Expected endpoint URL to be 'https://api.updated.com', got %s", req.EndpointUrl)
	}

	if req.AppDescription != "Updated test application" {
		t.Errorf("Expected app description to be 'Updated test application', got %s", req.AppDescription)
	}
}

// TestExternalApplicationDao validates the DAO structure
func TestExternalApplicationDao(t *testing.T) {
	testID := uuid.New()
	dao := ExternalApplicationDao{
		Id:             testID,
		Name:           "test-dao",
		EndpointUrl:    "https://api.dao.com",
		AppDescription: "Test DAO application",
	}

	if dao.Id != testID {
		t.Errorf("Expected ID to be %s, got %s", testID, dao.Id)
	}

	if dao.Name != "test-dao" {
		t.Errorf("Expected name to be 'test-dao', got %s", dao.Name)
	}

	if dao.EndpointUrl != "https://api.dao.com" {
		t.Errorf("Expected endpoint URL to be 'https://api.dao.com', got %s", dao.EndpointUrl)
	}

	if dao.AppDescription != "Test DAO application" {
		t.Errorf("Expected app description to be 'Test DAO application', got %s", dao.AppDescription)
	}
}
